### 数据结构
```
//runtime/runtime2.go
type _defer struct {
    siz     int32   //参数大小
    started bool    // defer是否被调用过的标识
    sp      uintptr // sp at time of defer
    pc      uintptr
    fn      *funcval // defer 后面跟的function
    _panic  *_panic  // panic that is running defer
    link    *_defer  // 链表结构
}
```
每一个defer关键字在编译阶段都会转换成deferproc，编译器会在函数return之前插入deferreturn。

### deferproc
```
 //runtime/panic.go
 // 创建一个defer, fn: defer后面的function, size: fn的参数大小
func deferproc(siz int32, fn *funcval) { // arguments of fn follow fn
    if getg().m.curg != getg() {
        // go code on the system stack can't defer
        throw("defer on system stack")
    }

    sp := getcallersp()
    // 参数首地址
    argp := uintptr(unsafe.Pointer(&fn)) + unsafe.Sizeof(fn)
    callerpc := getcallerpc()

    // 创建defer
    d := newdefer(siz) 
    if d._panic != nil {
        throw("deferproc: d.panic != nil after newdefer")
    }
    d.fn = fn
    d.pc = callerpc
    d.sp = sp
    switch siz {
    case 0:
        // Do nothing.
    case sys.PtrSize:
        *(*uintptr)(deferArgs(d)) = *(*uintptr)(unsafe.Pointer(argp))
    default:
        memmove(deferArgs(d), unsafe.Pointer(argp), uintptr(siz)) // 参数复制到d的后面
    }

    return0()
}
```
```
func newdefer(siz int32) *_defer {
    var d *_defer
    sc := deferclass(uintptr(siz))
    gp := getg()

    // 这里用到了两级缓存

    // deferpool [5][]*_defer, 可以看出只是部分缓存了， 最常用的部分
    if sc < uintptr(len(p{}.deferpool)) {

        // 这里获取的是 P (G、M、P)
        pp := gp.m.p.ptr()

        // 如果P的deferpool缓存不存在， 则从全局sched的deferpool转移一部分到P中
        if len(pp.deferpool[sc]) == 0 && sched.deferpool[sc] != nil {
            // Take the slow path on the system stack so
            // we don't grow newdefer's stack.
            systemstack(func() {
                lock(&sched.deferlock)
                for len(pp.deferpool[sc]) < cap(pp.deferpool[sc])/2 && sched.deferpool[sc] != nil {
                    d := sched.deferpool[sc]
                    sched.deferpool[sc] = d.link
                    d.link = nil
                    pp.deferpool[sc] = append(pp.deferpool[sc], d)
                }
                unlock(&sched.deferlock)
            })
        }

        // 从 P 中获取本地 defer
        if n := len(pp.deferpool[sc]); n > 0 {
            d = pp.deferpool[sc][n-1]
            pp.deferpool[sc][n-1] = nil
            pp.deferpool[sc] = pp.deferpool[sc][:n-1]
        }
    }

    // 上面缓存只是缓存了常见的一部分， 其余的部分直接内存分配
    if d == nil {
        // Allocate new defer+args.
        systemstack(func() {
            total := roundupsize(totaldefersize(uintptr(siz)))
            d = (*_defer)(mallocgc(total, deferType, true))
        })
    }
    d.siz = siz

    // 亮点来了， G 的 _defer指向了 d
    d.link = gp._defer
    gp._defer = d
    return d
}
```
根据defer参数的大小，会将常见大小缓存至 P 和 sched两级，其他大小则直接生成。
d会添加到G defer链表的首部，所以defer是一个后进先出的链表。

### deferreturn
```
func deferreturn(arg0 uintptr) {
    gp := getg()
    d := gp._defer
    if d == nil {
        return
    }
    sp := getcallersp()

    // G中的defer链表可能包含了很多函数的defer,那么怎么知道获取的这个defer就属于当前函数呢？
    // 如果 defer.sp 等于当前的 sp， 那么说明defer属于当前函数。（sp: 栈顶指针）
    if d.sp != sp {
        return
    }

    switch d.siz {
    case 0:
        // Do nothing.
    case sys.PtrSize:
        *(*uintptr)(unsafe.Pointer(&arg0)) = *(*uintptr)(deferArgs(d))
    default:
        memmove(unsafe.Pointer(&arg0), deferArgs(d), uintptr(d.siz))
    }
    fn := d.fn
    d.fn = nil
    gp._defer = d.link
    freedefer(d) // 重新放回缓存
    // jmpdefer由汇编实现，功能有两个， 1: 执行defer后面的代码， 2.跳转到deferreturn 实现循环
    jmpdefer(fn, uintptr(unsafe.Pointer(&arg0)))
}
```
G 退出时会遍历G上的defer链

### Goexit
```
func Goexit() {
    gp := getg()
    // 遍历 G 的_defer
    for {
        d := gp._defer
        if d == nil {
            break
        }
        // d 已经被调用过了
        if d.started {
            if d._panic != nil {
                d._panic.aborted = true
                d._panic = nil
            }
            d.fn = nil
            gp._defer = d.link
            freedefer(d)
            continue
        }
        // 设置被调用过
        d.started = true
        // 调用defer后面的function
        reflectcall(nil, unsafe.Pointer(d.fn), deferArgs(d), uint32(d.siz), uint32(d.siz))
        if gp._defer != d {
            throw("bad defer entry in Goexit")
        }
        d._panic = nil
        d.fn = nil
        gp._defer = d.link
        freedefer(d)
        // Note: we ignore recovers here because Goexit isn't a panic
    }
    goexit1()
}
```
结论： defer会涉及到内存分类、缓存、多次调用，所以会有一定的性能问题。