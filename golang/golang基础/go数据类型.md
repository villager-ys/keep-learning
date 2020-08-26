#### 类型转化问题
1. go语言不允许隐士类型转换
2. 别名和原有类型也不能进行隐士类型转换


需要转换使用强转

```
package test

import "testing"

type MyInt int32
func TestType(t *testing.T)  {
	var a int32 = 1
	var b int64 =2
	a = int32(b)
	var c MyInt
	c = MyInt(a)
	t.Log(a,b,c)
}

```

#### 类型预定义值：
1. math.MaxInt64
2. math.MaxFloat64
3. math.MaxUint64

#### 指针类型：
1. 不支持指针运算
2. string是值类型，默认的初始化值是空字符串，而不是nil
