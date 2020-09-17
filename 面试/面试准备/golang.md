### go调度(gpm)
- G goroutine 我们代码写的go func(){ }
- M 内核线程
- P M调度G的上下文, P中存储了很多G,M通过调用P来获取并执行G。
- schedt - 全局调度器，主要存储了一些空闲的G、M、P,runable的G

go程序启动的大致流程：
1. 创建g0
2. 创建m0
3. m.g0 = g0
4. g0.m = m0
5. 命令行初始化，OS初始化
6. schedinit 调度器初始化，主要做：空间申请、M的最大数量设置、P的数量设置、初始化参数和环境等
7. newproc为main.main创建一个主goroutine
8. mstart,运行主goroutine-->执行main.main

go func(){}我们写的协程，创建goroutine,调用newproc方法，newproc方法首先会获取func和参数，判断当前P本地是否有空闲G或是全局有没有空闲的G，如果都没有，则创建一个新的G,并设置成GDead状态，加入全局的G列表;如果本地没有空闲G而全局有，那么就全局空闲G列表移动32个到本地空闲G列表，从本地空闲G列表中pop一个G，初始化栈，把参数复制到栈，清除G的运行现场，因为G有可能是从P中获取的，清除原有的数据，状态设置成runable，调用runqput函数，加入当前P的runable的队列，如果本地runable队列已满，移动一半到全局，再次尝试加入，最后在判断有空闲的P且没有自旋的M，调用wakep()，wakep()获取全局空闲M和空闲P绑定，如果全局不存在空闲M新生成M，接下来唤醒M执行mstart，然后调用schedule来调度任务，schedule这个函数主要是找到一个runnable的g，然后调用execute来启动g，先从全局队列或是本地（P）的队列获取一个runable g，如果本地和全局都没有runable的g则调用findrunnable，findrunnable这个函数会阻塞知道找到一个可运行的g，调用execute执行找到的g，execute执行完后执行那个goexit,将G设置为GDead,G与M解绑，将G放入本地空闲列表，如果本地空闲个数大于64则移动一半到全局，继续调用schedule,不断的去获取G，执行G

存放G的P怎么来的？

在程序启动的时候有一个环节是schedinit，会调用procresize生成对应个数的P，这里我们就可以修改变量GOMAXPROCS来动态修改P的个数，所以在procresize中会对P数组进行调整，或新增P或减少P。被减少的P会将自身的runable、runnext、gfee移到全局去。
除了当前P外，所有P都设为Pidle,也就是不和M关联，如果P中没有runable的G,则将P加入全局空闲P,否则获取全局空闲M和P绑定

抢占式调度实现(sysmon线程):

runtime.main会创建一个额外的M运行sysmon函数，抢占式调度就是在sysmon中实现的。

sysmon中有netpool(获取fd事件)，retake(抢占)，forcegc(按时间强制执行gc),scavenge heap(释放自由列表中多余的项减少内存占用)等处理。

retake函数负责处理抢占，流程是:

- 枚举所有的P：

如果P在系统调用中(_Psyscall), 且经过了一次sysmon循环(20us~10ms), 则抢占这个P

调用handoffp解除M和P之间的关联

如果P在运行中(_Prunning), 且经过了一次sysmon循环并且G运行时间超过forcePreemptNS(10ms), 则抢占这个P

调用preemptone函数

设置g.preempt = true

设置g.stackguard0 = stackPreempt

stackguard0这个值用于检查当前栈空间是否需要扩张栈，stackPreempt是一个特殊的常量, 它的值会比任何的栈地址都要大, 检查时一定会触发栈扩张.栈扩张是会保存G的状态到g.sched, 切换到g0和g0的栈空间, 然后调用newstack函数。
newstack函数判断g.stackguard0等于stackPreempt, 就知道这是抢占触发的, 这时会再检查一遍是否要抢占：
- 如果M被锁定(函数的本地变量中有P), 则跳过这一次的抢占并调用gogo函数继续运行G
- 如果M正在分配内存, 则跳过这一次的抢占并调用gogo函数继续运行G
- 如果M设置了当前不能抢占, 则跳过这一次的抢占并调用gogo函数继续运行G
- 如果M的状态不是运行中, 则跳过这一次的抢占并调用gogo函数继续运行G

即使这一次抢占失败, 因为g.preempt等于true, runtime中的一些代码会重新设置stackPreempt以重试下一次的抢占。
如果判断可以抢占, 则继续判断是否GC引起的, 如果是则对G的栈空间执行标记处理(扫描根对象)然后继续运行。
如果不是GC引起的则调用gopreempt_m函数完成抢占。
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

slice struct:array指针，len 元素个数,cap slice第一个元素到底层数组的最后一个元素的长度,扩容是以这个值为标准

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
1. Stack scan: 收集根对象（全局变量和 goroutine 栈上的变量），收集根对象（全局变量，和G stack），开启写屏障。全局变量开启写屏障需要STW，G stack只需要停止该G就好，时间比较少。。
2. Mark: 标记对象，直到标记完所有根对象和根对象可达对象。此时写屏障会记录所有指针的更改(通过 mutator)。
3. Mark Termination: 重新扫描部分全局变量和发生更改的栈变量，完成标记，该阶段会STW(Stop The World)，也是 gc 时造成 go 程序停顿的主要阶段。
4. Sweep: 并发的清除未标记的对象。

目前整个GC流程会进行两次STW(Stop The World), 第一次是Stack scan阶段, 第二次是Mark Termination阶段.

从1.8以后的golang将第一步的stop the world 也取消了，这又是一次优化； 1.9开始, 写屏障的实现使用了Hybrid Write Barrier, 大幅减少了第二次STW的时间.

什么是三色标记?

三色标记法，是传统标记-清除算法的一种优化，主要思想是增加了一种中间状态，即灰色对象，以减少 STW 时间。
三色标记将对象分为黑色、白色、灰色三种：
- 黑色：已标记的对象，表示对象是根对象可达的。
- 白色：未标记对象，gc开始时所有对象为白色，当gc结束时，如果仍为白色，说明对象不可达，在 sweep 阶段会被清除。
- 灰色：被黑色对象引用到的对象，但其引用的自对象还未被扫描，灰色为标记过程的中间状态，当灰色对象全部被标记完成代表本次标记阶段结束。

什么时候触发:
- 阈值：默认内存扩大一倍，启动gc
- 定期：默认2min触发一次gc，src/runtime/proc.go:forcegcperiod
- 手动：runtime.gc()

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
channel底层是一个hchan的结构，由环形数据缓冲队列、类型信息、goroutine等待队列组成

make chan实际上就是初始化hchan结构的过程

从一个channel读数据简单过程如下：
1. 先获取channel全局锁
2. 尝试从sendq等待队列中获取等待的goroutine，
3. 如有等待的goroutine，没有缓冲区，取出goroutine并读取数据，然后唤醒这个goroutine，结束读取释放锁。
4. 如有等待的goroutine，且有缓冲区（此时缓冲区已满），从缓冲区队首取出数据，再从sendq取出一个goroutine，将goroutine中的数据存入buf队尾，结束读取释放锁。
5. 如没有等待的goroutine，且缓冲区有数据，直接读取缓冲区数据，结束读取释放锁。
6. 如没有等待的goroutine，且没有缓冲区或缓冲区为空，将当前的goroutine加入recvq排队，进入睡眠，等待被写goroutine唤醒。结束读取释放锁。

向一个channel中写数据简单过程如下：
1. 锁定整个通道结构。
2. 确定写入。尝试从recvq等待队列中取一个g，然后将元素直接写入goroutine。
3. 如果recvq为Empty，则确定缓冲区是否可用。如果可用，从当前goroutine复制数据到缓冲区。
4. 如果缓冲区已满，则要写入的元素将保存在当前正在执行的goroutine的结构中，并且当前goroutine加入sendq中并挂起，等待被唤醒。
5. 写入完成释放锁。

关闭chan:

关闭channel时会把recvq中的G全部唤醒，本该写入G的数据位置为nil。把sendq中的G全部唤醒，但这些G会panic。

chansend、chanrecv、closechan发现里面都加了锁，所以chan是线程安全的

#### go内存分配
Go语言内置运行时（就是runtime），抛弃了传统的内存分配方式,而是在程序启动时，会先向操作系统申请一块内存，自己进行管理。

申请到的内存块被分配了三个区域，arena(堆区)，bitmap区(arena区域哪些地址保存了对象),spans(存放mspan的指针，每个指针对应一页)。mspan也叫内存管理单元是一个包含起始地址、mspan规格、页的数量等内容的双端链表。

内存管理组件分为:
- mcache:每个线程M会绑定给一个处理器P，而每个P都会绑定一个上面说的本地缓存mcache，这样就可以直接给Goroutine分配，因为不存在多个Goroutine竞争的情况，所以不会消耗锁资源。mcache在初始化的时候是没有任何mspan资源的，在使用过程中会动态地从mcentral申请，之后会缓存下来。
- mcentral: 为所有mcache提供切分好的mspan资源。每个central保存一种特定大小的全局mspan列表。当工作goroutine的P中mcache没有合适的mspan时就会从mcentral获取。mcentral被所有的工作线程共同享有，存在多个Goroutine竞争的情况，因此会消耗锁资源。
- mheap:代表Go程序持有的所有堆空间。当mcentral没有空闲的mspan时，会向mheap申请。而mheap没有资源时，会向操作系统申请新内存。mheap主要用于大对象的内存分配，以及管理未切割的mspan，用于给mcentral切割成小对象。

分配流程：

Go的内存分配器在分配对象时，根据对象的大小，分成三类：小对象（小于等于16B）、一般对象（大于16B，小于等于32KB）、大对象（大于32KB）。

大体上的分配流程：
- 大于32KB 的对象，直接从mheap上分配；
- <=16B 的对象使用mcache的tiny分配器分配；
- (16B,32KB] 的对象，首先计算对象的规格大小，然后使用mcache中相应规格大小的mspan分配；
- 如果mcache没有相应规格大小的mspan，则向mcentral申请
- 如果mcentral没有相应规格大小的mspan，则向mheap申请
- 如果mheap中也没有合适大小的mspan，则向操作系统申请



#### sync.Once

#### 内存对齐
Go编译器可能会在结构体的相邻字段之间填充一些字节。 这使得一个结构体类型的尺寸并非等于它的各个字段类型尺寸的简单相加之和,这种现象称为内存对齐

#### 为什么要内存对齐
- 平台问题：并不是所有的硬件平台都能访问任意地址上的任意数据。
- 性能问题：访问未对齐内存需要cpu进行两次访问，对齐后只需要一次。

#### sync.Pool
整个设计充分利用了go.runtime的调度器优势：一个P下goroutine竞争的无锁化；
     
一个goroutine固定在一个局部调度器P上，从当前 P 对应的 poolLocal 取值， 若取不到，则从对应的 shared 数组上取，若还是取不到；则尝试从其他 P 的 shared 中偷。 若偷不到，则调用 New 创建一个新的对象。池中所有临时对象在一次 GC 后会被全部清空。