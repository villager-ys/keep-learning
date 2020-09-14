#### sync包中的type
包中主要有: Locker, Cond, Map, Mutex, Once, Pool,kRWMutex, WaitGroup

- Mutex: 互斥锁
- RWMutex：读写锁
- WaitGroup：并发等待组
- Once：执行一次
- Cond：信号量
- Pool：临时对象池
- Map：自带锁的map

#### sync.Mutex
sync.Mutex称为互斥锁，常用在并发编程里面。互斥锁需要保证的是同一个时间段内不能有多个并发协程同时访问某一个资源(临界区)。

sync.Mutex有2个函数Lock和UnLock分别表示加锁和释放锁。
```
func (m *Mutex) Lock(){}
func (m *Mutex) UnLock()
```
for example: 我们经常使用网上支付购物东西，就会出现同一个银行账户在某一个时间既有支出也有收入，那银行就得保证我们余额准确，保证数据无误。我们可以简单的实现银行的支出和收入来说明Mutex的使用
```go
package back
import "sync"
import "errors"

type Bank struct {
	sync.Mutex
	balance map[string]float64
}

// In 收入
func (b *Bank) In(account string, value float64) {
	// 加锁 保证同一时间只有一个协程能访问这段代码
	b.Lock()
	defer b.Unlock()

	v, ok := b.balance[account]
	if !ok {
		b.balance[account] = 0.0
	}

	b.balance[account] = v + value
}

// Out 支出
func (b *Bank) Out(account string, value float64) error {
	// 加锁 保证同一时间只有一个协程能访问这段代码
	b.Lock()
	defer b.Unlock()

	v, ok := b.balance[account]
	if !ok || v < value {
		return errors.New("account not enough balance")
	}

	b.balance[account] -= value
	return nil
}
```
#### sync.RWMutex
sync.RWMutex称为读写锁。

RWMutex有5个函数，分别为读和写提供锁操作
```
写操作
func (rw *RWMutex) Lock()
func (rw *RWMutex) Unlock()

读操作
func (rw *RWMutex) RLock()
func (rw *RWMutex) RUnlock()

RLocker()能获取读锁，然后传递给其他协程使用。
func (rw *RWMutex) RLocker() Locker
```

#### sync.WaitGroup
sync.WaitGroup指的是等待组，在Golang并发编程里面非常常见，指的是等待一组工作完成后，再进行下一组工作

sync.WaitGroup有3个函数
```
func (wg *WaitGroup) Add(delta int)  Add添加n个并发协程
func (wg *WaitGroup) Done()  Done完成一个并发协程
func (wg *WaitGroup) Wait()  Wait等待其它并发协程结束
```
sync.WaitGroup在Golang编程里面最常用于协程池，例如
```
func main() {
     wg := &sync.WaitGroup{}
     for i := 0; i < 1000; i++ {
         wg.Add(1)
         go func() {
	     defer func() {
		wg.Done()
	     }()
	     time.Sleep(1 * time.Second)
	     fmt.Println("hello world ~")
	 }()
     }
     // 等待所有协程结束
     wg.Wait()
     fmt.Println("WaitGroup all process done ~")
}
```
sync.WaitGroup没有办法指定最大并发协程数，在一些场景下会有问题。为了能够控制最大的并发数，推荐使用github.com/remeh/sizedwaitgroup，用法和sync.WaitGroup非常类似。

#### sync.Once
sync.Once指的是只执行一次的对象实现，常用来控制某些函数只能被调用一次。sync.Once的使用场景例如单例模式、系统初始化。

sync.Once的结构如下所示，只有一个函数。使用变量done来记录函数的执行状态，使用sync.Mutex和sync.atomic来保证线程安全的读取done。
```
type Once struct {
	m    Mutex     #互斥锁
	done uint32    #执行状态
}

func (o *Once) Do(f func())
```

#### sync.Cond
sync.Cond指的是同步条件变量，一般需要与互斥锁组合使用，本质上是一些正在等待某个条件的协程的同步机制。
```
// NewCond returns a new Cond with Locker l.
func NewCond(l Locker) *Cond {
    return &Cond{L: l}
}

// A Locker represents an object that can be locked and unlocked.
type Locker interface {
    Lock()
    Unlock()
}
```
sync.Cond有3个函数Wait、Signal、Broadcast
```
// Wait 等待通知
func (c *Cond) Wait()
// Signal 单播通知
func (c *Cond) Signal()
// Broadcast 广播通知
func (c *Cond) Broadcast()
```
举个例子
```
var sharedRsc = make(map[string]interface{})
func main() {
    var wg sync.WaitGroup
    wg.Add(2)
    m := sync.Mutex{}
    c := sync.NewCond(&m)
    
    go func() {
        // this go routine wait for changes to the sharedRsc
        c.L.Lock()
        for len(sharedRsc) == 0 {
            c.Wait()
        }
        fmt.Println(sharedRsc["rsc1"])
        c.L.Unlock()
        wg.Done()
    }()

    go func() {
        // this go routine wait for changes to the sharedRsc
        c.L.Lock()
        for len(sharedRsc) == 0 {
            c.Wait()
        }
        fmt.Println(sharedRsc["rsc2"])
        c.L.Unlock()
        wg.Done()
    }()

    // this one writes changes to sharedRsc
    c.L.Lock()
    sharedRsc["rsc1"] = "foo"
    sharedRsc["rsc2"] = "bar"
    c.Broadcast()
    c.L.Unlock()
    wg.Wait()
}
```
#### sync.Pool
sync.Pool指的是临时对象池，Golang和Java具有GC机制，因此很多开发者基本上都不会考虑内存回收问题。

Gc是一把双刃剑，带来了编程的方便但同时也增加了运行时开销，使用不当可能会严重影响程序的性能，因此性能要求高的场景不能任意产生太多的垃圾。

sync.Pool正是用来解决这类问题的，Pool可以作为临时对象池来使用，不再自己单独创建对象，而是从临时对象池中获取出一个对象。

sync.Pool有2个函数Get和Put，Get负责从临时对象池中取出一个对象，Put用于结束的时候把对象放回临时对象池中。
```
func (p *Pool) Get() interface{}
func (p *Pool) Put(x interface{})
```

官方的例子
```
var bufPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func timeNow() time.Time {
    return time.Unix(1136214245, 0)
}

func Log(w io.Writer, key, val string) {
    // 获取临时对象，没有的话会自动创建
    b := bufPool.Get().(*bytes.Buffer)
    b.Reset()
    b.WriteString(timeNow().UTC().Format(time.RFC3339))
    b.WriteByte(' ')
    b.WriteString(key)
    b.WriteByte('=')
    b.WriteString(val)
    w.Write(b.Bytes())
    // 将临时对象放回到 Pool 中
    bufPool.Put(b)
}

func main() {
    Log(os.Stdout, "path", "/search?q=flowers")
}
```
从上面的例子我们可以看到创建一个Pool对象并不能指定大小，所以sync.Pool的缓存对象数量是没有限制的(只受限于内存)，那sync.Pool是如何控制缓存临时对象数的呢？

sync.Pool在init的时候注册了一个poolCleanup函数，它会清除所有的pool里面的所有缓存的对象，该函数注册进去之后会在每次Gc之前都会调用，因此sync.Pool缓存的期限只是两次Gc之间这段时间。正因Gc的时候会清掉缓存对象，所以不用担心pool会无限增大的问题。

正因为如此sync.Pool适合用于缓存临时对象，而不适合用来做持久保存的对象池(连接池等)。

#### sync.Map
Go在1.9版本之前自带的map对象是不具有并发安全的，很多时候我们都得自己封装支持并发安全的Map结构，如下所示给map加个读写锁sync.RWMutex。
```
type MapWithLock struct {
    sync.RWMutex
    M map[string]Kline
}
```
Go1.9版本新增了sync.Map它是原生支持并发安全的map，sync.Map封装了更为复杂的数据结构实现了比之前加读写锁锁map更优秀的性能。

key and value都是interface{}，但对key有要求，需要是Comparable，不支持map, slice, func，在使用sync.Map时必须自己进行类型检查

方案一

建议将sync.Map放到一个结构体中，然后为结构体提供多个方法，在方法的参数中明确参数类型，这样go编译器就可以帮助类型检测，确保类型ok

简单，但需要为每个使用到的类型定义方法，比较麻烦

方案二

也是将sync.Map放到一个结构体中，但使用reflect.Type类规定keyType和valueType，初始化结构体时指定

```
type ConcurrentMap struct {
 m         sync.Map
 keyType   reflect.Type
 valueType reflect.Type
}

func (cMap *ConcurrentMap) Load(key interface{}) (value interface{}, ok bool) {
 if reflect.TypeOf(key) != cMap.keyType {
  return
 }
 return cMap.m.Load(key)
}

func (cMap *ConcurrentMap) Store(key, value interface{}) {
 if reflect.TypeOf(key) != cMap.keyType {
  panic(fmt.Errorf("wrong key type: %v", reflect.TypeOf(key)))
 }
 if reflect.TypeOf(value) != cMap.valueType {
  panic(fmt.Errorf("wrong value type: %v", reflect.TypeOf(value)))
 }
 cMap.m.Store(key, value)
}
```
sync.Map并发原理
```
type Map struct {
	mu Mutex
	read atomic.Value // readOnly
	dirty map[interface{}]*entry
	misses int
}
```
- 有两个字典：read和dirty，其中read是atomic.Value类型，存取都是原子操作不需要锁
- 两个字典中存的key和value都是*interface{}类型，这样任何一个字典的值update都是更新指针地址，都是可以是原子操作（atomic中有相应的unsafe.Pointer操作）
- 相同的key指向相同的value（*entry {p unsafe.Pointer}）
- read中的key是只读的，value可以直接udpate，删除就是entry.p=nil
- dirty中就是一个普通的map，访问需要加锁，新增和删除和普通map操作类似

1. 访问是先从read中找，如果没有就去dirty中找，misses记录read没有命中的次数
2. 更新是先从read中找，如果没有就去dirty中找，不论在哪个中找到，直接update entry.p
3. 删除是先从read中找，如果没有就去dirty中找，如果在read中 entry.p=nil，如果read中没有但在dirty中，那就加锁然后delete(map, key)