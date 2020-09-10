### go调度(gpm)

### go struct能不能比较

可以能，也可以不能。因为go存在不能使用==判断类型：map、slice，如果struct包含这些类型的字段，则不能比较。这两种类型也不能作为map的key。

### go defer

类似栈操作，后进先出。

因为go的return是一个非原子性操作，比如语句 return i，实际上分两步进行，即将i值存入栈中作为返回值，然后执行跳转，而defer的执行时机正是跳转前，所以说defer执行时还是有机会操作返回值的。

### select

### context
goroutine管理、信息传递。context的意思是上下文，在线程、协程中都有这个概念，它指的是程序单元的一个运行状态、现场、快照，包含。context在多个goroutine中是并发安全的。

### client如何实现长链接


### 主协程如何等其余协程完再操作
- waitgroup
- channel

### slice，len，cap，共享，扩容

len：切片的长度，访问时间复杂度为O(1)，go的slice底层是对数组的引用。

cap：切片的容量，扩容是以这个值为标准。默认扩容是2倍，当达到1024的长度后，按1.25倍。

扩容：每次扩容slice底层都将先分配新的容量的内存空间，再将老的数组拷贝到新的内存空间，因为这个操作不是并发安全的。所以并发进行append操作，读到内存中的老数组可能为同一个，最终导致append的数据丢失。

共享：slice的底层是对数组的引用，因此如果两个切片引用了同一个数组片段，就会形成共享底层数组。当sliec发生内存的重新分配（如扩容）时，会对共享进行隔断。

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
