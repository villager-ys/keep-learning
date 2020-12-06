## 前言

在动手实现调度队列前，我们应该先来学习参考一下那些优秀的开源项目里是怎么实现调度队列的。Kubernetes的调度器的调度算法的设计里使用了调度队列，在调度队列的实现里，使用了两个不同的队列。

第一个队列，叫做activeQ。凡是在activeQ 里的Pod，都是下一个调度周期需要调度的对象。第二个队列，叫做unschedulableQ，专门用来存放调度失败的Pod。而这里的一个关键点就在于，当一个 unschedulableQ 里的Pod被更新之后，调度器会自动把这个Pod移动到activeQ里。所以，当你在 Kubernetes集群里新创建或者更新一个Pod 的时候，调度器会将这个Pod入队到activeQ 里面。Kubernetes的调度器会不断从activeQ队列里出队（Pop）一个Pod进行调度。

## Kubernetes的调度队列实现
我们来看一下Kubernetes的activeQ调度队列的出队和入队操作是怎么实现的。

```
func (p *PriorityQueue) Add(pod *v1.Pod) error {
    p.lock.Lock()
    defer p.lock.Unlock()
    pInfo := p.newPodInfo(pod)
    // 将Pod加入到activeQ队列
    if err := p.activeQ.Add(pInfo); err != nil {
        klog.Errorf("Error adding pod %v/%v to the scheduling queue: %v", pod.Namespace, pod.Name, err)
        return err
    }
  ......

    p.nominatedPods.add(pod, "")
    p.cond.Broadcast()

    return nil
}


func (p *PriorityQueue) Pop() (*framework.PodInfo, error) {
    p.lock.Lock()
    defer p.lock.Unlock()
    for p.activeQ.Len() == 0 {
        // 当队列为空时，调用Pop方法会阻塞调用者直到有新元素入队。
        // 另外判断队列是否已被关闭，防止调用者一直阻塞在这里
        if p.closed {
            return nil, fmt.Errorf(queueClosed)
        }
        p.cond.Wait()
    }
    obj, err := p.activeQ.Pop()
    if err != nil {
        return nil, err
    }
    pInfo := obj.(*framework.PodInfo)
    pInfo.Attempts++
    p.schedulingCycle++
    return pInfo, err
}

type PriorityQueue struct {
  ......
    lock sync.RWMutex
    cond sync.Cond
    ......
    closed bool
}
```

如果调用Pop方法出队时如果队列为空则会调用p.cond.Wait()让调用者goroutine进入阻塞休眠等待，新元素入队的Add方法则是在元素入队后会调用p.cond.Broadcast()进行通知，唤醒被使用p.cond.Wait()陷入休眠状态的调用者goroutine。通过PriorityQueue类型的定义可以看出来这个功能是依赖标准库的sync.Cond并发原语实现的

针对并发环境下可能会有多个调用者在进行等待，那么p.cond.Broadcast()在唤醒所有等待者后是怎么避免产生多个gouroutine操作Pop出队造成数据竞争的呢？我们看一下sync.Cond这个原语的源代码实现。
```
type Cond struct {
    noCopy noCopy

    // 当观察或者修改等待条件的时候需要加锁
    L Locker

    // 等待队列
    notify  notifyList
    checker copyChecker
}

func NewCond(l Locker) *Cond {
    return &Cond{L: l}
}

func (c *Cond) Wait() {
    c.checker.check()
    // 增加到等待队列中
    t := runtime_notifyListAdd(&c.notify)
    c.L.Unlock()
    // 阻塞休眠调用者goroutine，直到重新唤醒才会执行下面的Lock获取独占锁
    runtime_notifyListWait(&c.notify, t)
    c.L.Lock()
}

func (c *Cond) Signal() {
    c.checker.check()
    // 唤醒一个等待者
    runtime_notifyListNotifyOne(&c.notify)
}

func (c *Cond) Broadcast() {
    c.checker.check()
    // 唤醒队列中的所有等待者
    runtime_notifyListNotifyAll(&c.notify）
}
```
可以看到，在调用Cond的Wait()方法阻塞休眠调用者goroutine后，当通过Broadcast或者Signal唤醒调用者后，调用者会从之前休眠的地方醒来执行下面的c.L.Lock()方法获取Cond原语自带的独占锁，这样就能避免唤醒的多个调用者goroutine同时执行例如上面p.cond.Wait()方法后面activeQ出队的逻辑而造成数据竞争了。

## sync.Cond
### Cond的适用场景
可以看到Kubernetes的调度队列是通过sync.Cond实现的调度控制。Cond并发原语相对sync包里的其他并发原语来说不是那么常用，它是sync包为等待 / 通知场景下的并发问题提供的支持。我们真正使用 Cond 的场景比较少，通常一旦遇到需要使用Cond 的场景，我们更多地会使用 Channel 的方式。但是对于需要多次通知的场景，比如上面Kubernetes 的例子，每次往队列中成功增加了元素后就需要调用Broadcast通知所有的等待者，使用Cond 就再合适不过了。因为Channel关闭后无法再次打开复用所以通过关闭Channel只能实现一次通知的功能，无法达到多次通知等待者的效果。

### Cond的基本用法
标准库中的Cond并发原语初始化的时候，需要关联一个Locker接口的实例，一般我们使用Mutex或者 RWMutex。通过上面列出来的Cond原语的源代码能看到它提供的并发控制方法有三个Broadcast、Signal 和 Wait 方法。

### Signal 方法
允许调用者 Caller 唤醒一个等待此Cond的goroutine。如果此时没有等待的goroutine，显然无需通知 waiter；如果Cond等待队列中有一个或者多个等待的goroutine，则需要从等待队列中移除第一个 goroutine并把它唤醒。调用Signal方法时，不强求调用者goroutine一定要持有c.L锁。

### Broadcast 方法
允许调用者Caller唤醒所有等待此Cond 的 goroutine。如果此时没有等待的 goroutine，显然无需通知 waiter；如果Cond 等待队列中有一个或者多个等待的goroutine，则清空队列中所有等待的goroutine，并全部唤醒。同样地，调用Broadcast 方法时，也不强求你一定持有c.L 的锁。

### Wait 方法
会把调用者Caller放入Cond的等待队列中并阻塞，直到被Signal或者Broadcast的方法从等待队列中移除并唤醒。调用Wait方法时必须要持有c.L的锁。

**上面这段文字节选自极客时间《Go并发编程实战课》的Cond:条件变量的实现机制及避坑指南。**
### Cond的使用示例
下面这个例子可以比较好的说明Cond并发原语的使用方法：

```
package main

import (
    "fmt"
    "strconv"
    "sync"
    "time"
)

var queue []struct{}

func main() {
    var wg sync.WaitGroup
    wg.Add(2)
    c := sync.NewCond(&sync.Mutex{})

    for i := 0; i < 2; i ++ {
        go func(i int) {
            // this go routine wait for changes to the sharedRsc
            c.L.Lock()
            for len(queue) <= 0 {
                fmt.Println("goroutine" + strconv.Itoa(i) +" wait")
                c.Wait()
            }
            fmt.Println("goroutine" + strconv.Itoa(i), "pop data")
            queue = queue[1:]
            c.L.Unlock()
            wg.Done()
        }(i)

    }


    for i := 0; i < 2; i ++ {
        // 主goroutine延迟两秒准备好后把变量设置为true
        time.Sleep(2 * time.Second)
        c.L.Lock()
        fmt.Println("main goroutine push data")
        queue= append(queue, struct{}{})
        c.Broadcast()
        fmt.Println("main goroutine broadcast")
        c.L.Unlock()

    }

    wg.Wait()
}
```
在这个例子里我们开启两个goroutine来从queue里边取数据，一开始队列为空，为了模拟多个等待者等候通知从队列里取数据，我们每个两秒往队列中存入一个元素，存入新元素后通过Broadcast方法通知等待者。此时之前两个陷入休眠等待的goroutine都会被唤醒。通过上面Wait方法源码中的逻辑我们知道，醒来后两个goroutine会通过Cond.L.Lock()争夺队列的使用权，所以主goroutine通知他们有新元素入队后，只有一个等待者goroutien能从队列中取出数据。取出数据释放锁之后，另一个等待者获取到Cond.L锁，但是因为队列此时再次为空，另外一个只能再次调用Wait方法新休眠等待下一次通知。

### 关于Cond原语Wait方法的使用有两点需要注意：

1. 调用Cond.Wait 方法之前一定要通过Cond.L.Lock加锁。
2. 不能省略条件检查只调用一次Cond.Wait。

针对第一点，通过Cnd.Wait 方法的代码实现能知道，把当前调用者加入到通知队列后会释放锁（如果不释放锁，其他 Wait 的调用者就没有机会加入到 通知 队列中了），然后一直等待；等调用者被唤醒之后，又会去争抢这把锁。如果调用Cond.Wait之前不加锁的话，就有可能会去释放一个未加锁的Locker而造成panic。

主goroutine发送通知唤醒所有等待者后，并不意味着所有等待者都满足了等待条件，就像上面代码示例里描述的比较特殊的情况，队列为空入队一个元素后发送通知，此时只有一个等待者能够从队列中出队数据，另外的等待者则需继续等待下次通知。这也是为什么我们在上面示例代码里使用了循环结构调用Wait方法的原因。

```
for len(queue) <= 0 {
    fmt.Println("goroutine" + strconv.Itoa(i) +" wait")
    c.Wait()
}
```
既然调用Cond.Wait 方法之前一定要通过Cond.L.Lock()加锁，那么我相信一定会有人问：“那为什么Kubernetes实现的那个调度队列里没用cond.L.Lock()加锁？” Em... 其实也加了，只不过是cond.L共用了自己结构定义里的lock锁。感兴趣的朋友可以去Kubernetes调度队列源码里学习一下这个技巧。

## 实现自己的调度队列
在看完Kubernetes的队列实现后我们知道对于调度队列这种存在多次往复等待 / 通知的场景，使用sync包提供的Cond原语再合适不过了。熟知sync.Cond的实现原理以及实现方法后要自己实现一个队列也不是什么难事儿。