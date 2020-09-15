### 什么是go pprof
pprof是Go的性能分析工具，在程序运行过程中，可以记录程序的运行信息，可以是CPU使用情况、内存使用情况、goroutine运行情况等，当需要性能调优或者定位Bug时候，这些记录的信息是相当重要。

### 基本使用
使用pprof有多种方式，Go已经现成封装好了1个：net/http/pprof，使用简单的几行命令，就可以开启pprof，记录运行信息，并且提供了Web服务，能够通过浏览器和命令行2种方式获取运行数据。

demo
```
// 展示内存增长和pprof，并不是泄露
package main

import (
    "fmt"
    "net/http"
    _ "net/http/pprof"
    "os"
    "time"
)

// 运行一段时间：fatal error: runtime: out of memory
func main() {
    // 开启pprof
    go func() {
        ip := "0.0.0.0:6060"
        if err := http.ListenAndServe(ip, nil); err != nil {
            fmt.Printf("start pprof failed on %s\n", ip)
            os.Exit(1)
        }
    }()

    tick := time.Tick(time.Second / 100)
    var buf []byte
    for range tick {
        buf = append(buf, make([]byte, 1024*1024)...)
    }
}
```
#### 浏览器方式
输入网址ip:port/debug/pprof/打开pprof主页，从上到下依次是5类profile信息：
1. block：goroutine的阻塞信息，本例就截取自一个goroutine阻塞的demo，但block为0，没掌握block的用法
2. goroutine：所有goroutine的信息，下面的full goroutine stack dump是输出所有goroutine的调用栈，是goroutine的debug=2，后面会详细介绍。
3. heap：堆内存的信息
4. mutex：锁的信息
5. threadcreate：线程信息

#### 命令行方式
当连接在服务器终端上的时候，是没有浏览器可以使用的，Go提供了命令行的方式，能够获取以上5类信息，这种方式用起来更方便。

使用命令go tool pprof url可以获取指定的profile文件，此命令会发起http请求，然后下载数据到本地，之后进入交互式模式，就像gdb一样，可以使用命令查看运行信息，以下是5类请求的方式：

```
# 下载cpu profile，默认从当前开始收集30s的cpu使用情况，需要等待30s
go tool pprof http://localhost:6060/debug/pprof/profile   # 30-second CPU profile
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=120     # wait 120s

# 下载heap profile
go tool pprof http://localhost:6060/debug/pprof/heap      # heap profile

# 下载goroutine profile
go tool pprof http://localhost:6060/debug/pprof/goroutine # goroutine profile

# 下载block profile
go tool pprof http://localhost:6060/debug/pprof/block     # goroutine blocking profile

# 下载mutex profile
go tool pprof http://localhost:6060/debug/pprof/mutex
```

上面这个demo会不断的申请内存，把它编译运行起来，然后执行：
```
$ go tool pprof http://localhost:6060/debug/pprof/heap

Fetching profile over HTTP from http://localhost:6060/debug/pprof/heap
Saved profile in /home/ubuntu/pprof/pprof.demo.alloc_objects.alloc_space.inuse_objects.inuse_space.001.pb.gz       //<--- 下载到的内存profile文件
File: demo // 程序名称
Build ID: a9069a125ee9c0df3713b2149ca859e8d4d11d5a
Type: inuse_space
Time: May 16, 2019 at 8:55pm (CST)
Entering interactive mode (type "help" for commands, "o" for options)
(pprof)
(pprof)
(pprof) help  // 使用help打印所有可用命令
  Commands:
    callgrind        Outputs a graph in callgrind format
    comments         Output all profile comments
    disasm           Output assembly listings annotated with samples
    dot              Outputs a graph in DOT format
    eog              Visualize graph through eog
    evince           Visualize graph through evince
    gif              Outputs a graph image in GIF format
    gv               Visualize graph through gv
    kcachegrind      Visualize report in KCachegrind
    list             Output annotated source for functions matching regexp
    pdf              Outputs a graph in PDF format
    peek             Output callers/callees of functions matching regexp
    png              Outputs a graph image in PNG format
    proto            Outputs the profile in compressed protobuf format
    ps               Outputs a graph in PS format
    raw              Outputs a text representation of the raw profile
    svg              Outputs a graph in SVG format
    tags             Outputs all tags in the profile
    text             Outputs top entries in text form
    top              Outputs top entries in text form
    topproto         Outputs top entries in compressed protobuf format
    traces           Outputs all profile samples in text form
    tree             Outputs a text rendering of call graph
    web              Visualize graph through web browser
    weblist          Display annotated source in a web browser
    o/options        List options and their current values
    quit/exit/^D     Exit pprof
    
    ....
```
关于命令，本文只会用到3个，我认为也是最常用的：top、list、traces，分别介绍一下。

#### top
按指标大小列出前10个函数，比如内存是按内存占用多少，CPU是按执行时间多少。
```
(pprof) top
Showing nodes accounting for 814.62MB, 100% of 814.62MB total
      flat  flat%   sum%        cum   cum%
  814.62MB   100%   100%   814.62MB   100%  main.main
         0     0%   100%   814.62MB   100%  runtime.main
```
top会列出5个统计数据：

- flat: 本函数占用的内存量。
- flat%: 本函数内存占使用中内存总量的百分比。
- sum%: 前面每一行flat百分比的和，比如第2行虽然的100% 是 100% + 0%。
- cum: 是累计量，加入main函数调用了函数f，函数f占用的内存量，也会记进来。
- cum%: 是累计量占总量的百分比。

#### list
查看某个函数的代码，以及该函数每行代码的指标信息，如果函数名不明确，会进行模糊匹配，比如list main会列出main.main和runtime.main。
```
(pprof) list main.main  // 精确列出函数
Total: 814.62MB
ROUTINE ======================== main.main in /home/ubuntu/heap/demo2.go
  814.62MB   814.62MB (flat, cum)   100% of Total
         .          .     20:    }()
         .          .     21:
         .          .     22:    tick := time.Tick(time.Second / 100)
         .          .     23:    var buf []byte
         .          .     24:    for range tick {
  814.62MB   814.62MB     25:        buf = append(buf, make([]byte, 1024*1024)...)
         .          .     26:    }
         .          .     27:}
         .          .     28:
(pprof) list main  // 匹配所有函数名带main的函数
Total: 814.62MB
ROUTINE ======================== main.main in /home/ubuntu/heap/demo2.go
  814.62MB   814.62MB (flat, cum)   100% of Total
         .          .     20:    }()
         .          .     21:
..... // 省略几行
         .          .     28:
ROUTINE ======================== runtime.main in /usr/lib/go-1.10/src/runtime/proc.go
         0   814.62MB (flat, cum)   100% of Total
         .          .    193:        // A program compiled with -buildmode=c-archive or c-shared
..... // 省略几行
```
可以看到在main.main中的第25行占用了814.62MB内存，左右2个数据分别是flat和cum，含义和top中解释的一样。

#### traces
打印所有调用栈，以及调用栈的指标信息。
```
(pprof) traces
File: demo2
Type: inuse_space
Time: May 16, 2019 at 7:08pm (CST)
-----------+-------------------------------------------------------
     bytes:  813.46MB
  813.46MB   main.main
             runtime.main
-----------+-------------------------------------------------------
     bytes:  650.77MB
         0   main.main
             runtime.main
....... // 省略几十行
```
每个- - - - - 隔开的是一个调用栈，能看到runtime.main调用了main.main，并且main.main中占用了813.46MB内存。

### 什么是内存泄露
内存泄露指的是程序运行过程中已不再使用的内存，没有被释放掉，导致这些内存无法被使用，直到程序结束这些内存才被释放的问题。

Go虽然有GC来回收不再使用的堆内存，减轻了开发人员对内存的管理负担，但这并不意味着Go程序不再有内存泄露问题。在Go程序中，如果没有Go语言的编程思维，也不遵守良好的编程实践，就可能埋下隐患，造成内存泄露问题。

### 怎么发现内存泄露
在Go中发现内存泄露有2种方法，一个是通用的监控工具，另一个是go pprof：

- 监控工具：固定周期对进程的内存占用情况进行采样，数据可视化后，根据内存占用走势（持续上升），很容易发现是否发生内存泄露。
- go pprof：适合没有监控工具的情况，使用Go提供的pprof工具判断是否发生内存泄露。
