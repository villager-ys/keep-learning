### go调度(gpm)
- G goroutine 我们代码写的go func(){ }
- M 内核线程
- P M调度G的上下文, P中存储了很多G,M通过调用P来获取并执行G。
- schedt - 全局调度器，主要存储了一些空闲的G、M、P,runable的G

大致的启动流程：
1. 创建g0
2. 创建m0
3. m.g0 = g0
4. g0.m = m0
5. 命令行初始化，OS初始化
6. schedinit 调度器初始化，主要做：空间申请、M的最大数量设置、P的数量设置、初始化参数和环境等
7. newproc为main.main创建一个主goroutine
8. mstart,运行主goroutine-->执行main.main

go func(){}我们写的协程-->newproc获取func和参数-->切换到go，使用go栈空间，调用newproc1-->gfget从当前P获取空闲的G，此时如果当前P为空并且全局P不为空，则全局移动32个P到到本地，否则当前P中pop一个G，初始化栈空间-->判断获取G是否成功，不成功就创建一个G，初始化栈空间并加入全局G数组;成功就把参数复制到栈，清除G的运行现场，因为G有可能是从P中获取的，清除原有的数据-->状态设置成runable,设置goid-->runqput加入当前P的的runable队列，如果当前P的本地runable队列已满，则本地移除一半到全局-->有空闲的P且没有自旋的M，makep():添加一个P来执行goroutinue

存放G的P怎么来的？
在程序启动的时候有一个环节是schedinit，会调用procresize生成对应个数的P，这里我们就可以修改变量GOMAXPROCS来动态修改P的个数，所以在procresize中会对P数组进行调整，或新增P或减少P。被减少的P会将自身的runable、runnext、gfee移到全局去。
除了当前P外，所有P都设为Pidle,也就是不和M关联，如果P中没有runable,则将P加入全局空闲P,否则获取全局空闲M和P绑定

M从何而来？
有空闲P且没有自旋的M，makep()-->statrm,全局去获取空闲M，成功就将M和空闲P绑定，否则就创建一个M和空闲P绑定，唤醒M-->执行mstatr/mstart1-->开始循环schedule，从P本地或是全局获取runable的G，获取成功则G与M绑定，G设置成running,执行完成后goexit，将G设置成Dead，将G与M解绑，将G加入P的空闲的链表，如果本地空闲个数大于64则移一半到全局空闲G队列;如果P本地或是全局没有获取到runable的G，则阻塞在本地，全局，网络，其他P中获取，如果获取成功就同样执行上面过程，否则M与P解绑，P加入全局，M加入全局M，M休眠

### g0 
g0 这样一个特殊的 goroutine,用来创建 goroutine、deferproc 函数里新建 _defer、垃圾回收相关的工作（例如 stw、扫描 goroutine 的执行栈、一些标识清扫的工作、栈增长）等等。

### go struct能不能比较

可以能，也可以不能。因为go存在不能使用==判断类型：map、slice，如果struct包含这些类型的字段，则不能比较。这两种类型也不能作为map的key。

### go defer

每一个defer关键字在编译阶段都会转换成deferproc，编译器会在函数return之前插入deferreturn。Goexit时会遍历G上的defer链

_defer会添加到G defer链表的首部，所以defer是一个后进先出的链表。

### select

select执行大致可分为一下几步：
1. 创建select, newselect()创建
2. 注册case,selectsend()：注册写数据到到channel会调用的case;selectrecv():注册读数据到channel会调用的case;selectdefault:注册default的case
3. 执行select, selectgo():1,对case进行洗牌 2,对lockorder(存放case对应channel的地址)里头的元素排序，方便后面去重 3，对所有的channel加锁，同时遍历的时候根据前一个数据来判断是否要加锁去重 4, 循环执行，按顺序执行遍历case,遇到可执行的则跳出去执行,结束select,没有遇到，则阻塞等待channel有输入/输出，此时如果有输入/输入发生，判断是否有可执行的case,有就执行并释放select,否则继续循环
4. 释放select

select的几种场景：
1. 超时判断
2. 再另外一个协程中，如果运行遇到非法操作或不可处理的错误，就向通道发送数据通知主程序停止运行close
3. 判断channel是否堵塞

### context
goroutine管理、信息传递。context的意思是上下文，在线程、协程中都有这个概念，它指的是程序单元的一个运行状态、现场、快照，包含。context在多个goroutine中是并发安全的。

### 主协程如何等其余协程完再操作
- waitgroup
- channel

### slice，len，cap，共享，扩容

slice struct:数组指针，len 元素个数,cap slice第一个元素到底层数组的最后一个元素的长度,扩容是以这个值为标准

扩容(不是并发安全)的实现需要关注的有两点，一个是扩容时候的策略，还有一个就是扩容是生成全新的内存地址还是在原来的地址后追加。

扩容策略:
- 首先判断，如果新申请容量（cap）大于2倍的旧容量（old.cap），最终容量（newcap）就是新申请的容量（cap）
- 否则判断，如果旧切片的长度小于1024，则最终容量(newcap)就是旧容量(old.cap)的两倍，即（newcap=doublecap）
- 否则判断，如果旧切片长度大于等于1024，则最终容量（newcap）从旧容量（old.cap）开始循环增加原来的 1/4(1.25)，即（newcap=old.cap,for {newcap += newcap/4}）直到最终容量（newcap）大于等于新申请的容量(cap)，即（newcap >= cap）
- 如果最终容量（cap）计算值溢出，则最终容量（cap）就是新申请容量（cap）

扩容地址问题：扩容之后可能还是原来的数组，因为可能底层数组还有空间

扩容：每次扩容slice底层都将先分配新的容量的内存空间，再将老的数组拷贝到新的内存空间，因为这个操作不是并发安全的。所以并发进行append操作，读到内存中的老数组可能为同一个，最终导致append的数据丢失。

共享：slice的底层是对数组的引用，因此如果两个切片引用了同一个数组片段，就会形成共享底层数组。当sliec发生内存的重新分配（如扩容）时，会对共享进行隔断。

slice copy: slicecopy 方法会把源切片值(即 fm Slice )中的元素复制到目标切片(即 to Slice )中，并返回被复制的元素个数，copy 的两个类型必须一致。slicecopy 方法最终的复制结果取决于较短的那个切片，当较短的切片复制完成，整个复制过程就全部完成了。

### map如何顺序读取
map的底层是hash table(hmap类型)，对key值进行了hash，并将结果的低八位用于确定key/value存在于哪个bucket（bmap类型）。再将高八位与bucket的tophash进行依次比较，确定是否存在。出现hash冲撞时，会通过bucket的overflow指向另一个bucket，形成一个单向链表。每个bucket存储8个键值对。

如果要实现map的顺序读取，需要使用一个slice来存储map的key并按照顺序进行排序。

### 实现set
利用map，如果要求并发安全，就用sync.map

要注意下set中的delete函数需要使用 delete(map)来实现，但是这个并不会释放内存，除非value也是一个子map。当进行多次delete后，可以使用make来重建map。

### 实现消息队列（多生产者，多消费者）
使用sync.Map来管理topic，用channel来做队列。

### go GC

### 当go服务部署到线上了，发现有内存泄露，该怎么处理

### go的值传递和引用

### go内存逃逸

### 怎么理解go的interface
