### 一，获取命令行参数
os.Args
```
	fmt.Printf("参数共有:%d\n", len(os.Args))
	for key, value := range os.Args {
		fmt.Printf("args[%d],%v\n", key, value)
	}
```
### 二，flag
如果命令行输入参数没有顺序，使用os.Args就比较麻烦。

```
// 定义命令行参数对应的变量，这三个变量都是指针类型
var cliName = flag.String("name", "nick", "Input Your Name")
var cliAge = flag.Int("age", 28, "Input Your Age")
var cliGender = flag.String("gender", "male", "Input Your Gender")

// 另一种flag的定义方式
var cliFlag int

flag.IntVar(&cliFlag, "flagname", 1234, "Just for demo")

// 把用户传递的命令行参数解析为对应变量的值
flag.Parse()
```