### 字符串

1,string是数据类型，不是引用或者指针类型

2,string是只读的slice,len表示它的包含的byte数

3,string的byte数可以存放任何数据

```
	var s string //初始化默认值是""
	t.Log(len(s), s)
	s = "hello"
	t.Log(len(s))
	//s[1] = '2' //string 是不可变的byte slice
	s = "\xE4\xB8\xA5" //存储二进制数据
	t.Log(len(s), s)
	s = "中"
	t.Log(len(s)) //byte数
```

### Unicode&UTF-8
1,unicode是一种字符集(doce point)

2,utf-8是unicode的存储实现(转换为字节序列的规则)

byte和rune：

byte和rune实质上就是uint8和int32类型。byte用来强调数据是raw data，而不是数字；而rune用来表示Unicode的code point

```
	s := "hello 你好"
	t.Log(len(s))         // 12 字符串所占字节的长度
	t.Log(len([]byte(s))) // 12 字符串所占字节的长度
	//计算字符串的长度
	//golang中的unicode/utf8包提供了用utf-8获取长度的方法
	t.Log(utf8.RuneCountInString(s)) //8
	//通过rune类型处理unicode字符
	t.Log(len([]rune(s))) //8
	s = "中"
	c := []rune(s)
	t.Logf("中 的Unicode %x", c[0])
	t.Logf("中 的utf-8 %x", s)
```

计算字符串长度：

计算字符串所占字节的长度：

- len(str)
- len([]byte(str))

计算字符串长度而不是字节长度：

- 使用unicode/utf-8包中的RuneCountInString方法:utf8.RuneCountInString(str)
- 将字符串转rune再计算：len([]rune(str))

### 字符串使用range遍历
字符串结合range遍历是自动会获取rune而不是byte

```
	s := "中华有为-华为"
	//字符串range遍历自动会获取rune,而不是byte
	for _, c := range s {
		// %[1]c %[1]d代表都是以第一个遍历进行格式化输出
		t.Logf("%[1]c %[1]d", c)
		//t.Logf("%[1]c %[1]x", c)
	}
```


### 常用字符串函数

1. strings包
2. strcovn包

字符串常用函数比较多，使用时自行查文档

字符串分割和拼接

```
	s := "A-B-C"
	// Split分割获得切片
	parts := strings.Split(s, "-")
	for _, part := range parts {
		t.Logf("%s", part)
	}
	// 拼接
	str := strings.Join(parts, ",")
	t.Log(str)
```

字符串和int转换

```
	// 字符串和int转换
	s := strconv.Itoa(1)
	t.Log("字符串" + s)
	//t.Log(10 + strconv.Atoi("1"))
	//Atoi返回值是两个，不能直接相加
	if n, err := strconv.Atoi("1"); err == nil {
		t.Log(10 + n)
	}
```
需要注意strconv.Atoi("1")返回的是两个值