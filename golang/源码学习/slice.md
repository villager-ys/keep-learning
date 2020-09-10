### 数据结构
```
type slice struct {
    array unsafe.Pointer // 指向数组
    len   int            // 元素个数
    cap   int            // slice第一个元素到底层数组的最后一个元素的长度
}

注：slice大小为多少呢？ 24 byte
```
### 创建
```
func makeslice(et *_type, len, cap int) slice {

    maxElements := maxSliceCap(et.size)

    // 检验slice的长度、容量
    if len < 0 || uintptr(len) > maxElements {
        panicmakeslicelen()
    }

    if cap < len || uintptr(cap) > maxElements {
        panicmakeslicecap()
    }

    p := mallocgc(et.size*uintptr(cap), et, true)
    return slice{p, len, cap}
}


// 计算最大容量
func maxSliceCap(elemsize uintptr) uintptr {

    // 如果元素类型大小在maxElems长度内，取对应大小
    if elemsize < uintptr(len(maxElems)) {
        return maxElems[elemsize]
    }
    // maxAlloc:允许用户分配的最大内存空间
    return maxAlloc / elemsize
}

var maxElems = [...]uintptr{
    ^uintptr(0), // 就是64位无符号的0,按位异或 等于 18446744073709551615
    maxAlloc / 1, maxAlloc / 2, maxAlloc / 3, maxAlloc / 4,
    maxAlloc / 5, maxAlloc / 6, maxAlloc / 7, maxAlloc / 8,
    maxAlloc / 9, maxAlloc / 10, maxAlloc / 11, maxAlloc / 12,
    maxAlloc / 13, maxAlloc / 14, maxAlloc / 15, maxAlloc / 16,
    maxAlloc / 17, maxAlloc / 18, maxAlloc / 19, maxAlloc / 20,
    maxAlloc / 21, maxAlloc / 22, maxAlloc / 23, maxAlloc / 24,
    maxAlloc / 25, maxAlloc / 26, maxAlloc / 27, maxAlloc / 28,
    maxAlloc / 29, maxAlloc / 30, maxAlloc / 31, maxAlloc / 32,
}
```
这里为什么需要一个maxElems数组呢？ 估计是为了计算更快
### 扩容
```
func growslice(et *_type, old slice, cap int) slice {
    if raceenabled {
        callerpc := getcallerpc()
        racereadrangepc(old.array, uintptr(old.len*int(et.size)), callerpc, funcPC(growslice))
    }
    if msanenabled {
        msanread(old.array, uintptr(old.len*int(et.size)))
    }

    // 什么时候size == 0呢？ type zero struct{}
    // 见 https://stackoverflow.com/questions/57085905/which-types-size-is-zero-in-slice-of-golang#57086264
    if et.size == 0 {

        // 容量比以前小， 报错
        if cap < old.cap {
            panic(errorString("growslice: cap out of range"))
        }
        // append should not create a slice with nil pointer but non-zero len.
        // We assume that append doesn't need to preserve old.array in this case.
        return slice{unsafe.Pointer(&zerobase), old.len, cap}
    }

    newcap := old.cap
    doublecap := newcap + newcap
    // CASE 1: 如果申请的容量大于老容量的2倍， 则直接扩容至cap
    // CASE 2.1: 老长度 小于 1024， 则扩容至2倍老容量
    // CASE 2.2: 一直循环 newcap += newcap / 4 （5/4倍），直到newcap不小于申请的容量，
    if cap > doublecap {
        newcap = cap
    } else {
        if old.len < 1024 {
            newcap = doublecap
        } else {
            // Check 0 < newcap to detect overflow
            // and prevent an infinite loop.
            for 0 < newcap && newcap < cap {
                newcap += newcap / 4
            }
            // Set newcap to the requested cap when
            // the newcap calculation overflowed.
            if newcap <= 0 {
                newcap = cap
            }
        }
    }

    var overflow bool
    var lenmem, newlenmem, capmem uintptr
    // 新容量计算
    switch {
    case et.size == 1:
        lenmem = uintptr(old.len)
        newlenmem = uintptr(cap)
        capmem = roundupsize(uintptr(newcap))
        overflow = uintptr(newcap) > maxAlloc
        newcap = int(capmem)
    case et.size == sys.PtrSize:
        lenmem = uintptr(old.len) * sys.PtrSize
        newlenmem = uintptr(cap) * sys.PtrSize
        capmem = roundupsize(uintptr(newcap) * sys.PtrSize)
        overflow = uintptr(newcap) > maxAlloc/sys.PtrSize
        newcap = int(capmem / sys.PtrSize)
    case isPowerOfTwo(et.size):
        var shift uintptr
        if sys.PtrSize == 8 {
            // Mask shift for better code generation.
            shift = uintptr(sys.Ctz64(uint64(et.size))) & 63
        } else {
            shift = uintptr(sys.Ctz32(uint32(et.size))) & 31
        }
        lenmem = uintptr(old.len) << shift
        newlenmem = uintptr(cap) << shift
        capmem = roundupsize(uintptr(newcap) << shift)
        overflow = uintptr(newcap) > (maxAlloc >> shift)
        newcap = int(capmem >> shift)
    default:
        lenmem = uintptr(old.len) * et.size
        newlenmem = uintptr(cap) * et.size
        capmem = roundupsize(uintptr(newcap) * et.size)
        overflow = uintptr(newcap) > maxSliceCap(et.size)
        newcap = int(capmem / et.size)
    }

    if cap < old.cap || overflow || capmem > maxAlloc {
        panic(errorString("growslice: cap out of range"))
    }

    var p unsafe.Pointer
    if et.kind&kindNoPointers != 0 { 
        // 在老的切片后面继续扩充容量
        p = mallocgc(capmem, nil, false)
        // // 将 lenmem 这个多个 bytes 从 old.array地址 拷贝到 p 的地址处
        memmove(p, old.array, lenmem)
        // 先将 P 地址加上新的容量得到新切片容量的地址，然后将新切片容量地址后面的 capmem-newlenmem 个 bytes 这块内存初始化。为之后继续 append() 操作腾出空间。
        memclrNoHeapPointers(add(p, newlenmem), capmem-newlenmem)
    } else {
        // 重新申请新的数组给新切片
        // 重新申请 capmen 这个大的内存地址，并且初始化为0值
        p = mallocgc(capmem, et, true)
        if !writeBarrier.enabled {
            // 如果还不能打开写锁，那么只能把 lenmem 大小的 bytes 字节从 old.array 拷贝到 p 的地址处
            memmove(p, old.array, lenmem)
        } else {
            // 循环拷贝老的切片的值
            for i := uintptr(0); i < lenmem; i += et.size {
                typedmemmove(et, add(p, i), add(old.array, i))
            }
        }
    }

    return slice{p, old.len, newcap}
}
```
上述就是扩容的实现。主要需要关注的有两点，一个是扩容时候的策略，还有一个就是扩容是生成全新的内存地址还是在原来的地址后追加。

扩容策略:
- 首先判断，如果新申请容量（cap）大于2倍的旧容量（old.cap），最终容量（newcap）就是新申请的容量（cap）
- 否则判断，如果旧切片的长度小于1024，则最终容量(newcap)就是旧容量(old.cap)的两倍，即（newcap=doublecap）
- 否则判断，如果旧切片长度大于等于1024，则最终容量（newcap）从旧容量（old.cap）开始循环增加原来的 1/4(1.25)，即（newcap=old.cap,for {newcap += newcap/4}）直到最终容量（newcap）大于等于新申请的容量(cap)，即（newcap >= cap）
- 如果最终容量（cap）计算值溢出，则最终容量（cap）就是新申请容量（cap）

扩容地址问题：扩容之后可能还是原来的数组，因为可能底层数组还有空间

### slice copy
```
func slicecopy(to, fm slice, width uintptr) int {
	// 如果源切片或者目标切片有一个长度为0，那么就不需要拷贝，直接 return 
	if fm.len == 0 || to.len == 0 {
		return 0
	}
	// n 记录下源切片或者目标切片较短的那一个的长度
	n := fm.len
	if to.len < n {
		n = to.len
	}
	// 如果入参 width = 0，也不需要拷贝了，返回较短的切片的长度
	if width == 0 {
		return n
	}
	// 如果开启了竞争检测
	if raceenabled {
		callerpc := getcallerpc(unsafe.Pointer(&to))
		pc := funcPC(slicecopy)
		racewriterangepc(to.array, uintptr(n*int(width)), callerpc, pc)
		racereadrangepc(fm.array, uintptr(n*int(width)), callerpc, pc)
	}
	// 如果开启了 The memory sanitizer (msan)
	if msanenabled {
		msanwrite(to.array, uintptr(n*int(width)))
		msanread(fm.array, uintptr(n*int(width)))
	}

	size := uintptr(n) * width
	if size == 1 { 
		// TODO: is this still worth it with new memmove impl?
		// 如果只有一个元素，那么指针直接转换即可
		*(*byte)(to.array) = *(*byte)(fm.array) // known to be a byte pointer
	} else {
		// 如果不止一个元素，那么就把 size 个 bytes 从 fm.array 地址开始，拷贝到 to.array 地址之后
		memmove(to.array, fm.array, size)
	}
	return n
}
```
在这个方法中，slicecopy 方法会把源切片值(即 fm Slice )中的元素复制到目标切片(即 to Slice )中，并返回被复制的元素个数，copy 的两个类型必须一致。slicecopy 方法最终的复制结果取决于较短的那个切片，当较短的切片复制完成，整个复制过程就全部完成了。

