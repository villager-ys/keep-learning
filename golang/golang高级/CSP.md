### CSP
communicating sequential processes

“用于描述两个独立的并发实体通过共享的通讯 channel(管道)进行通信的并发模型。 CSP中channel是第一类对象，它不关注发送消息的实体，而关注与发送消息时使用的channel。”
```
package main

import (
	"fmt"
	"time"
)

func service() string {
	time.Sleep(time.Millisecond * 50)
	return "Done"
}

func otherTask() {
	fmt.Println("do something else")
	time.Sleep(time.Millisecond * 100)
	fmt.Println("Task is Done")
}

func asyncService() chan string {
	////串行
	//fmt.Println(service())
	//otherTask()
	//retCh := make(chan string)
	//定义一个buffered channel
	retCh := make(chan string, 1)
	go func() {
		ret := service()
		fmt.Println("return result")
		retCh <- ret
		fmt.Println("service exited")
	}()
	return retCh
}

func main() {
	retCh := asyncService()
	otherTask()
	fmt.Println(<-retCh)
}
```
执行顺序：
1. 主携程调用asyncService方法(方法中有匿名携程，但是主线程没管它继续运行)
2. 主携程调用otherTask方法，
3. 打印"do something else"，休眠100毫秒
4. 匿名携程启动，调用service方法获取返回值，打印"return result"，往channel里存放service方法返回值，打印"service exited"
5. 主携程继续执行otherTask方法，打印
"Task is Done"
6. 主携程打印channel里的值

![image](../images/result.png)