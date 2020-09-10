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
    if et.kind&kindNoPointers != 0 { //指针

        // 申请一块地址
        p = mallocgc(capmem, nil, false)
        // 将以前的数据拷贝到新申请的地址中
        memmove(p, old.array, lenmem)
        // The append() that calls growslice is going to overwrite from old.len to cap (which will be the new length).
        // Only clear the part that will not be overwritten.

        // 擦除未使用的地址
        memclrNoHeapPointers(add(p, newlenmem), capmem-newlenmem)
    } else {
        // Note: can't use rawmem (which avoids zeroing of memory), because then GC can scan uninitialized memory.
        // 申请一块地址
        p = mallocgc(capmem, et, true)
        // 将以前的数据拷贝到新地址中
        if !writeBarrier.enabled {
            memmove(p, old.array, lenmem)
        } else {
            for i := uintptr(0); i < lenmem; i += et.size {
                typedmemmove(et, add(p, i), add(old.array, i))
            }
        }
    }

    return slice{p, old.len, newcap}
}
```
在扩容的时候，明确调用了mallocgc重新申请地址， 为什么网上有些博文说“扩容之后可能还是原来的数组，因为可能底层数组还有空间” ？ 难道是我理解错了？