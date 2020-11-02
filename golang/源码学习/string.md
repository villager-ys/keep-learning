## string的底层数据结构
代码路径：`runtime/string.go:228`

```go
type stringStruct struct {
	str unsafe.Pointer
	len int
}
```

### demo例子
```go
func main() {
	test := "hello"
	p := (*str)(unsafe.Pointer(&test))
	fmt.Println(&p, p) // 0xc420070018 &{0xa3f71 5}
	c := make([]byte, p.length)
	for i := 0; i < p.length; i++ {
		tmp := uintptr(unsafe.Pointer(p.array))           // 指针类型转换通过unsafe包
		c[i] = *(*byte)(unsafe.Pointer(tmp + uintptr(i))) // 指针运算只能通过uintptr
	}
	fmt.Println(c)         // [104 101 108 108 111]
	fmt.Println(string(c)) // [byte] --> string, "hello"

	test2 := test + " world"  // 字符串是不可变类型，会生成一个新的string实例
	p2 := (*str)(unsafe.Pointer(&test2))
	fmt.Println(&p2, p2) // 0xc420028030 &{0xc42000a2e5 11}
	fmt.Println(test2)   // hello, world
}
```

### 字符串的拼接与修改
#### 1 +操作
string类型是一个不可变类型，那么任何对string的修改都会新生成一个string的实例，看一下runtime包下的拼接函数

```go
func concatstring2(buf *tmpBuf, a [2]string) string {
	return concatstrings(buf, a[:])
}

func concatstrings(buf *tmpBuf, a []string) string {
	idx := 0
	l := 0
	count := 0
	for i, x := range a {
		n := len(x)
		if n == 0 {
			continue
		}
		if l+n < l {
			throw("string concatenation too long")
		}
		l += n
		count++
		idx = i
	}
	if count == 0 {
		return ""
	}

	// If there is just one string and either it is not on the stack
	// or our result does not escape the calling frame (buf != nil),
	// then we can return that string directly.
	if count == 1 && (buf != nil || !stringDataOnStack(a[idx])) {
		return a[idx]
	}
	s, b := rawstringtmp(buf, l)
	for _, x := range a {
		copy(b, x) // copy操作
		b = b[len(x):]
	}
	return s
}
```

分析runtime的concatstrings实现，可以看出+最后新申请buf，拷贝原来的string到buf，最后返回新实例。那么每次的+操作，都会涉及新申请buf，然后是对应的copy。如果反复使用+，就不可避免有大量的申请内存操作，对于大量的拼接，性能就会受到影响了。

#### 2 bytes.Buffer
通过看源码，bytes.Buffer 增长buffer时是按照2倍来增长内存，可以有效避免频繁的申请内存，通过一个例子来看：

```go
func main() {
    var buf bytes.Buffer
    for i := 0; i < 10; i++ {
        buf.WriteString("hi ")
    }
    fmt.Println(buf.String())
}
```
对应的byte.WriteString函数源码
```go
// @file: buffer.go
func (b *Buffer) WriteString(s string) (n int, err error) {
    b.lastRead = opInvalid
    m, ok := b.tryGrowByReslice(len(s))
    if !ok {
        m = b.grow(len(s)) // 高效的增长策略 -> let capacity get twice as large
    }
    return copy(b.buf[m:], s), nil
}

// @file: buffer.go
// let capacity get twice as large !!!
func (b *Buffer) grow(n int) int {
    m := b.Len()
    // If buffer is empty, reset to recover space.
    if m == 0 && b.off != 0 {
        b.Reset()
    }
    // Try to grow by means of a reslice.
    if i, ok := b.tryGrowByReslice(n); ok {
        return i
    }
    // Check if we can make use of bootstrap array.
    if b.buf == nil && n <= len(b.bootstrap) {
        b.buf = b.bootstrap[:n]
        return 0
    }
    c := cap(b.buf)
    if n <= c/2-m {
        // We can slide things down instead of allocating a new
        // slice. We only need m+n <= c to slide, but
        // we instead let capacity get twice as large so we
        // don't spend all our time copying.
        copy(b.buf, b.buf[b.off:])
    } else if c > maxInt-c-n {
        panic(ErrTooLarge)
    } else {
        // Not enough space anywhere, we need to allocate.
        buf := makeSlice(2*c + n)
        copy(buf, b.buf[b.off:])
        b.buf = buf
    }
    // Restore b.off and len(b.buf).
    b.off = 0
    b.buf = b.buf[:m+n]
    return m
}
```
#### 3 strings.Join
这个函数可以一次申请最终string的大小，但是使用得预先准备好所有string，这种场景也是高效的，一个例子：
```go
func main() {
	var strs []string
	for i := 0; i < 10; i++ {
		strs = append(strs, "hi")
	}
	fmt.Println(strings.Join(strs, " "))
}
```
对应库的源码：
```go
// Join concatenates the elements of a to create a single string. The separator string
// sep is placed between elements in the resulting string.
func Join(a []string, sep string) string {
    switch len(a) {
    case 0:
        return ""
    case 1:
        return a[0]
    case 2:
        // Special case for common small values.
        // Remove if golang.org/issue/6714 is fixed
        return a[0] + sep + a[1]
    case 3:
        // Special case for common small values.
        // Remove if golang.org/issue/6714 is fixed
        return a[0] + sep + a[1] + sep + a[2]
    }
	
    // 计算好最终的string的大小
    n := len(sep) * (len(a) - 1)  //
    for i := 0; i < len(a); i++ {
        n += len(a[i])
    }

    b := make([]byte, n)
    bp := copy(b, a[0])
    for _, s := range a[1:] {
        bp += copy(b[bp:], sep)
        bp += copy(b[bp:], s)
    }
    return string(b)
}
```
#### 4 strings.Builder
看到这个名字，就想到了Java的库。其高效也是体现在2倍速的内存增长, WriteString函数利用了slice类型对应append函数的2倍速增长。
```go
func main() {
    var s strings.Builder
    for i := 0; i < 10; i++ {
        s.WriteString("hi ")
    }
    fmt.Println(s.String())
}
```
对应库源码:
```go
@file: builder.go
// WriteString appends the contents of s to b's buffer.
// It returns the length of s and a nil error.
func (b *Builder) WriteString(s string) (int, error) {
    b.copyCheck()
    b.buf = append(b.buf, s...)
    return len(s), nil
}
```