### map初始化

```
	m1:= map[string]int{"one":1,"two":2,"three":3}
	t.Log(len(m1))
	t.Log(m1["two"])
	m2:= map[int]int{}
	m2[4]=16
	t.Log(len(m2))
	t.Log(m2[4])
	m3:=make(map[int]int,10)
	t.Log(m3)
```

### map访问不存在的key
go语言访问map不存在的key获得value是0，不是nil

```
	m1:=make(map[int]int,10)
	t.Log(m1[0],m1[2])
	m1[2]=0
	t.Log(m1[2])
	if v,ok:=m1[3];ok{
		t.Log("key3存在",v)
	}else{
		t.Log("key3不存在",v)
	}
```
![image](../images/map.png)

### map遍历k,v

```
	m1:= map[string]int{"one":1,"two":2,"three":3}
	for k,v := range m1{
		t.Log(k,v)
	}
```
### map与工厂模式

map的value可以是一个函数，这样可以实现工厂模式(不太理解)

```
	m1 := map[int]func(op int)int{}
	m1[1]= func(op int)int{return op}
	m1[2]= func(op int) int {return op*op}
	m1[3]= func(op int) int {return op*op*op}
	t.Log(m1[1](2),m1[2](2),m1[3](2))
```

### map for set

go语言没有set数据结构，可以使用map[type]boolean来实现set结构

```
	m1 := map[int]bool{}
	m1[1] = true
	n := 3
	if m1[n] {
		t.Logf("%d is exit", n)
	} else {
		t.Logf("%d is not exit", n)
	}
	delete(m1, 1)
	n = 1
	if m1[n] {
		t.Logf("%d is exit", n)
	} else {
		t.Logf("%d is not exit", n)
	}
```
1. 元素唯一性
2. 增加元素
3. 删除元素
4. 元素个数