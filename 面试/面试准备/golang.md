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

### map
golang中map比较重要的数据结构hmap和bmap,hmap中的buckets代表了存储数据的bucket数组，bmap则代表了map中bucket本体。
每一个bmap最多放8个key和value，最后由一个overflow字段指向下一个bmap。

查找：
- 根据key计算出哈希值
- 根据哈希值低位确定所在bucket
- 根据哈希值高8位确定在bucket中的存储位置
- 当前bucket未找到则查找对应的overflow bucket。
- 对应位置有数据则对比完整的哈希值，确定是否是要查找的数据
- 如果当前处于map进行了扩容，处于数据搬移状态，则优先从oldbuckets查找。

插入：
- 根据key计算出哈希值
- 根据哈希值低位确定所在bucket
- 根据哈希值高8位确定在bucket中的存储位置
- 查找该key是否存在，已存在则更新，不存在则插入

扩容：
- 当链表越来越长，其实是扩容的加载因子达到6.5，buckets就会进行扩容，将原来bucket数组数量扩充一倍，产生一个新的bucket数组，也就是hmap的buckets属性指向的数组。这样hmap中的oldbuckets属性指向的就是旧bucket数组。
- 这里的加载因子LoadFactor是一个阈值，计算方式为（元素个数/桶个数 ）如果超过6.5，将会进行扩容，这个是经过测试才得出的合理的一个阈值。因为，加载因子越小，空间利用率就小，加载因子越大，产生冲突的几率就大。所以6.5是一个平衡的值。
- map的扩容不会立马全部复制，而是渐进式扩容，即首先开辟2倍的内存空间，创建一个新的bucket数组。只有当访问old bucket数组时，才会调用growWork()方法将old bucket中的元素拷贝到新的bucket数组，进行渐进式的扩容。当然旧的数据不会删除，而是去掉引用，等待gc回收

删除：

删除操作并不会释放内存：删除的核心就把对应的tophash设置为empty，而不是直接删除了内存里面的数据。

map是并发不安全的，那么如何实现呢？

新建一个struct，在map的基础上加RWMutex。或者直接使用sync.map

sync.Map

如果你接触过大Java，那你一定对ConcurrentHashMap利用锁分段技术(ConcurrentHashMap是由Segment(可重入锁ReentrantLock)数组结构和HashEntry数组)增加了锁的数目，从而使争夺同一把锁的线程的数目得到控制的原理记忆深刻。
那么Golang的sync.Map是否也是使用了相同的原理呢？sync.Map的原理很简单，使用了空间换时间策略，通过冗余的两个数据结构(read、dirty),实现加锁对性能的影响。
通过引入两个map将读写分离到不同的map，其中read map提供并发读和已存元素原子写，而dirty map则负责读写。 这样read map就可以在不加锁的情况下进行并发读取,当read map中没有读取到值时,再加锁进行后续读取,并累加未命中数,当未命中数大于等于dirty map长度,将dirty map上升为read map。从之前的结构体的定义可以发现，虽然引入了两个map，但是底层数据存储的是指针，指向的是同一份值。

### sync.Map设计点
1. 空间换时间。通过冗余的两个数据结构(read、dirty),实现加锁对性能的影响。
2. 使用只读数据(read)，避免读写冲突。
3. 动态调整，miss次数多了之后，将dirty数据提升为read。
4. double-checking（双重检测）。
5. 延迟删除。 删除一个键值只是打标记，只有在提升dirty的时候才清理删除的数据。
6. 优先从read读取、更新、删除，因为对read的读取不需要锁。
7. 虽然read和dirty有冗余数据，但这些数据是通过指针指向同一个数据，所以尽管Map的value会很大，但是冗余的空间占用还是有限的。

### 实现set
利用map，如果要求并发安全，就用sync.map

要注意下set中的delete函数需要使用 delete(map)来实现，但是这个并不会释放内存，除非value也是一个子map。当进行多次delete后，可以使用make来重建map。

### 实现消息队列（多生产者，多消费者）
使用sync.Map来管理topic，用channel来做队列。

### go GC

golang GC 采用基于标记-清除的三色标记法。大约经历如下几步:

1. Stack scan: 收集根对象（全局变量和 goroutine 栈上的变量），该阶段会开启写屏障(Write Barrier)。
2. Mark: 标记对象，直到标记完所有根对象和根对象可达对象。此时写屏障会记录所有指针的更改(通过 mutator)。
3. Mark Termination: 重新扫描部分全局变量和发生更改的栈变量，完成标记，该阶段会STW(Stop The World)，也是 gc 时造成 go 程序停顿的主要阶段。
4. Sweep: 并发的清除未标记的对象。

什么是三色标记?

三色标记法，是传统标记-清除算法的一种优化，主要思想是增加了一种中间状态，即灰色对象，以减少 STW 时间。
三色标记将对象分为黑色、白色、灰色三种：
- 黑色：已标记的对象，表示对象是根对象可达的。
- 白色：未标记对象，gc开始时所有对象为白色，当gc结束时，如果仍为白色，说明对象不可达，在 sweep 阶段会被清除。
- 灰色：被黑色对象引用到的对象，但其引用的自对象还未被扫描，灰色为标记过程的中间状态，当灰色对象全部被标记完成代表本次标记阶段结束。

### 当go服务部署到线上了，发现有内存泄露，该怎么处理

### go的值传递和引用传递
golang默认都是采用值传递，即拷贝传递，如果想要使用引用传递，需要将传入的参数设置为 指针类型。如果传入的参数数据很大，建议使用指针类型，减少内存因拷贝参数而占用。

golang引用传递比一定比值传递效率高，传递指针相比值传递减少了底层拷贝，是可以提高效率，但是拷贝的数据量较小，由于指针传递会产生逃逸，可能会使用堆，也可能增加gc的负担，所以指针传递不一定是高效的。


### go内存逃逸
1. 指针逃逸,方法返回局部变量指针，就形成变量逃逸
2. 栈空间不足逃逸(空间开辟过大),实际上当栈空间不足以存放当前对象或无法判断当前切片长时会将对象分配到堆中
3. 动态类型逃逸,编译期间很难确定其参数的具体类型，也能产生逃逸度
4. 闭包引用对象逃逸
5. 跨协程引用对象逃逸,原本属于A协程的变量，通过指针传递给B协程使用，产生逃逸

### 怎么理解go的interface
1. interface 是一种类型
2. interface 变量存储的是实现者的值
3. 可以使用断言判断interface变量存储的是哪种类型
4. interface{} 是一个空的 interface 类型，认为所有的类型都实现了 interface{}

## golang new 和 make的区别
- make 返回的是值
- new 返回的是指针
- make 只能用于内置类型 这个是没有问题的
- 这要看对象所处的作用域，如果逃逸分析确定对象不会跑出方法里，是可以分配在栈上的
- 问：那map，slice 和 channel 可以用new 吗？
   不可以，有内置初始化操作
![make参数类型.png](https://upload-images.jianshu.io/upload_images/331298-01759889aa9ed5a1.png?imageMogr2/auto-orient/strip%7CimageView2/2/w/1240)
- silce、map、channel等类型属于引用类型，引用类型初始化为nil，nil是不能直接赋值的

### channel实现