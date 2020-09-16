#### 数据结构
```
type _panic struct {
	argp      unsafe.Pointer // pointer to arguments of deferred call run during panic; cannot move - known to liblink
	arg       interface{}    // argument to panic
	link      *_panic        // link to earlier panic
	recovered bool           // whether this panic is over
	aborted   bool           // the panic was aborted
}
```
在 panic 中是使用 _panic 作为其基础单元的，每执行一次 panic 语句，都会创建一个 _panic 对象。它包含了一些基础的字段用于存储当前的 panic 调用情况，涉及的字段如下

- argp：指向 defer 延迟调用的参数的指针
- arg：panic 的原因，也就是调用 panic 时传入的参数
- link：指向上一个调用的 _panic，这里说明panic也是一个链表
- recovered：panic 是否已经被处理过，也就是是否被 recover
- aborted：panic 是否被中止

通过查看 link 字段，可得知其是一个链表的数据结构

#### gopanic
panic 编译后主要是调用了 runtime.gopanic
```
func gopanic(e interface{}) {
	gp := getg()
	......
	var p _panic
	p.arg = e
	// 头插法
	p.link = gp._panic
	gp._panic = (*_panic)(noescape(unsafe.Pointer(&p)))

	for {
		d := gp._defer
		if d == nil {
			break
		}

		// If defer was started by earlier panic or Goexit (and, since we're back here, that triggered a new panic),
		// take defer off list. The earlier panic or Goexit will not continue running.
		if d.started {
			if d._panic != nil {
				d._panic.aborted = true
			}
			d._panic = nil
			d.fn = nil
			gp._defer = d.link
			freedefer(d)
			continue
		}

		// Mark defer as started, but keep on list, so that traceback
		// can find and update the defer's argument frame if stack growth
		// or a garbage collection happens before reflectcall starts executing d.fn.
		d.started = true

		// Record the panic that is running the defer.
		// If there is a new panic during the deferred call, that panic
		// will find d in the list and will mark d._panic (this panic) aborted.
		d._panic = (*_panic)(noescape(unsafe.Pointer(&p)))

		p.argp = unsafe.Pointer(getargp(0))
		reflectcall(nil, unsafe.Pointer(d.fn), deferArgs(d), uint32(d.siz), uint32(d.siz))
		p.argp = nil

		// reflectcall did not panic. Remove d.
		if gp._defer != d {
			throw("bad defer entry in panic")
		}
		d._panic = nil
		d.fn = nil
		gp._defer = d.link

		// trigger shrinkage to test stack copy. See stack_test.go:TestStackPanic
		//GC()

		pc := d.pc
		sp := unsafe.Pointer(d.sp) // must be pointer so it gets adjusted during stack copy
		freedefer(d)
		if p.recovered {
			atomic.Xadd(&runningPanicDefers, -1)

			gp._panic = p.link
			// Aborted panics are marked but remain on the g.panic list.
			// Remove them from the list.
			for gp._panic != nil && gp._panic.aborted {
				gp._panic = gp._panic.link
			}
			if gp._panic == nil { // must be done with signal
				gp.sig = 0
			}
			// Pass information about recovering frame to recovery.
			gp.sigcode0 = uintptr(sp)
			gp.sigcode1 = pc
			mcall(recovery)
			throw("recovery failed") // mcall should not return
		}
	}

	preprintpanics(gp._panic)

	fatalpanic(gp._panic) // should not return
	*(*int)(nil) = 0      // not reached
}
```
- 获取指向当前 Goroutine 的指针
- 初始化一个 panic 的基本单位 _panic，并将这个panic头插入当前goroutine的panic链表中。
- 获取当前 Goroutine 上挂载的 _defer（数据结构也是链表）
- 若当前存在 defer 调用，则调用 reflectcall 方法去执行先前 defer 中延迟执行的代码。reflectcall方法若在执行过程中需要运行 recover 将会调用 gorecover 方法。
- 结束前，使用 preprintpanics 方法打印出所涉及的 panic 消息
- 最后调用 fatalpanic 中止应用程序，实际是执行 exit(2) 进行最终退出行为的。

通过对上述代码的执行分析，可得知 panic 方法实际上就是处理当前 Goroutine(g) 上所挂载的 ._panic 链表（所以无法对其他 Goroutine 的异常事件响应），然后对其所属的 defer 链表和 recover 进行检测并处理，最后调用退出命令中止应用程序。

#### recover panic
```
func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("recover: %v", err)
		}
	}()
	
	panic("sim lou.")
}
```
我们看汇编代码，panic是怎么被recover的：
```
"".main STEXT size=118 args=0x0 locals=0x50
	......
	0x003a 00058 (panic_test2.go:6)	CALL	runtime.deferprocStack(SB)
	......
	0x005a 00090 (panic_test2.go:12)	CALL	runtime.gopanic(SB)
	......
	0x0060 00096 (panic_test2.go:6)	CALL	runtime.deferreturn(SB)
	......
"".main.func1 STEXT size=151 args=0x0 locals=0x40
	0x0000 00000 (panic_test2.go:6)	TEXT	"".main.func1(SB), ABIInternal, $64-0
	......
	0x0026 00038 (panic_test2.go:7)	CALL	runtime.gorecover(SB)
	......
	0x0092 00146 (panic_test2.go:6)	JMP	0
```

通过分析底层调用，可得知主要是如下几个方法：

- runtime.deferprocStack
- runtime.gopanic
- runtime.deferreturn
- runtime.gorecover

前面我们说了简单的流程，gopanic 方法会遍历调用当前 Goroutine 下的 defer 链表，若 reflectcall 执行中遇到 recover 就会调用 gorecover 进行处理，该方法代码如下

```
func gorecover(argp uintptr) interface{} {
	// Must be in a function running as part of a deferred call during the panic.
	// Must be called from the topmost function of the call
	// (the function used in the defer statement).
	// p.argp is the argument pointer of that topmost deferred function call.
	// Compare against argp reported by caller.
	// If they match, the caller is the one who can recover.
	gp := getg()
	p := gp._panic
	if p != nil && !p.recovered && argp == uintptr(p.argp) {
		p.recovered = true
		return p.arg
	}
	return nil
}
```
这代码，看上去挺简单的，核心就是修改 recovered 字段。该字段是用于标识当前 panic 是否已经被 recover 处理。但是这和我们想象的并不一样啊，程序是怎么从 panic 流转回去的呢？是不是在核心方法里处理了呢？我们再看看 gopanic 的代码，如下：
```
func gopanic(e interface{}) {
    ...
    for {
        // defer...
        ...
        pc := d.pc
        sp := unsafe.Pointer(d.sp) // must be pointer so it gets adjusted during stack copy
        freedefer(d)

        // recover...
        if p.recovered {
            atomic.Xadd(&runningPanicDefers, -1)

            gp._panic = p.link
            for gp._panic != nil && gp._panic.aborted {
                gp._panic = gp._panic.link
            }
            if gp._panic == nil { 
                gp.sig = 0
            }

            gp.sigcode0 = uintptr(sp)
            gp.sigcode1 = pc
            mcall(recovery)
            throw("recovery failed") 
        }
    }
    ...
}
```
我们回到 gopanic 方法中再仔细看看，发现实际上是包含对 recover 流转的处理代码的。恢复流程如下：
- 判断当前 _panic 中的 recover 是否已标注为处理
- 从 _panic 链表中删除已标注中止的 panic 事件，也就是删除已经被恢复的 panic 事件
- 将相关需要恢复的栈帧信息传递给 recovery 方法的 gp 参数（每个栈帧对应着一个未运行完的函数。栈帧中保存了该函数的返回地址和局部变量）
- 执行 recovery 进行恢复动作

从流程来看，最核心的是 recovery 方法。它承担了异常流转控制的职责。
```
func recovery(gp *g) {
	// Info about defer passed in G struct.
	sp := gp.sigcode0
	pc := gp.sigcode1

	// d's arguments need to be in the stack.
	if sp != 0 && (sp < gp.stack.lo || gp.stack.hi < sp) {
		print("recover: ", hex(sp), " not in [", hex(gp.stack.lo), ", ", hex(gp.stack.hi), "]\n")
		throw("bad recovery")
	}

	// Make the deferproc for this d return again,
	// this time returning 1.  The calling function will
	// jump to the standard return epilogue.
	gp.sched.sp = sp
	gp.sched.pc = pc
	gp.sched.lr = 0
	gp.sched.ret = 1
	gogo(&gp.sched)
}
```
粗略一看，似乎就是很简单的设置了一些值？但实际上设置的是编译器中伪寄存器的值，常常被用于维护上下文等。在这里我们需要结合 gopanic 方法一同观察 recovery 方法。它所使用的栈指针 sp 和程序计数器 pc 是由当前 defer 在调用流程中的 deferproc 传递下来的，因此实际上最后是通过 gogo 方法跳回了 deferproc 方法。另外我们注意到：

gp.sched.ret = 1

在底层中程序将 gp.sched.ret 设置为了 1，也就是没有实际调用 deferproc 方法，直接修改了其返回值。意味着默认它已经处理完成。直接转移到 deferproc 方法的下一条指令去。至此为止，异常状态的流转控制就已经结束了。接下来就是继续走 defer 的流程了.

### 总结：
从 panic 和 recover 这对关键字的实现上可以看出，可恢复的 panic 必须要 recover 的配合。 而且，这个 recover 必须位于同一 goroutine 的直接调用链上（例如，如果 A 依次调用了 B 和 C，而 B 包含了 recover，而 C 发生了 panic，则这时 B 的 panic 无法恢复 C 的 panic； 又例如 A 调用了 B 而 B 又调用了 C，那么 C 发生 panic 时，如果 A 要求了 recover 则仍然可以恢复）， 否则无法对 panic 进行恢复。

当一个 panic 被恢复后，调度并因此中断，会重新进入调度循环，进而继续执行 recover 后面的代码， 包括比 recover 更早的 defer（因为已经执行过得 defer 已经被释放，而尚未执行的 defer 仍在 goroutine 的 defer 链表中）， 或者 recover 所在函数的调用方。