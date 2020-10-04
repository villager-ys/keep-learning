## namespace资源隔离
linux内核提拱了6种namespace隔离的系统调用，但是真正的容器还需要处理许多其他工作。
- UTS： 主机名或域名
- IPC： 信号量、消息队列和共享内存
- PID： 进程编号
- Network： 网络设备、网络战、端口等
- Mount： 挂载点（文件系统）
- User： 用户组和用户组

实际上，linux内核实现namespace的主要目的，就是为了实现轻量级虚拟化技术服务。在同一个namespace下的进程合一感知彼此的变化，而对外界的进程一无所知。这样就可以让容器中的进程产生错觉，仿佛自己置身一个独立的系统环境中，以达到隔离的目的。

## 1.进行namespace API操作的4种方式
namespace的API包括clone()、setns()以及unshare()，还有/proc下的本分文件。为了确定隔离的到底是哪6项namespace，在使用这些API时需要指定一下6个参数中的一个或多个，通过|（位或）操作实现。这6个参数分别是CLONE_NEWUTS、CLONE_NEWIPC、CLONE_NEWPID、CLONE_NEWNET、CLONE_NEWNS、CLONE_NEWUSER。

- 通过clone()在创建新进程的同时创建namespace

使用clone()来创建一个独立的namespace，是最常见的用法，也是docker使用namespace最基本用法。

```int clone(int (*child_func)(void *),void *child_stack,int flags,void *arg);```

clone()实际上是linux系统调用fork()的一种更通用的实现方式，他可以通过flag来控制使用多少功能。一共有二十多种CLONE_*的flag(标志位)参数来控制clone进程的方方面面。

- 通过setns()加入一个已经存在的namespace

在进程都结束的情况下，也可以通过挂载的形式把namespace保留下来，保留namespace的目的是为以后有进程加入作准备。在docker中，使用docker exec命令在意境运行的容器中执行新的命令，就需要用到该方法。通过setns()系统调用，进程从原先的namespace加入到某个已经存在的namespace，通常为了使新加入的pid namespace生效，会在setns()函数执行后使用clone()创建子进程继续执行新命令，让原先的进程结束运行。

- 通过unshare()在原先进程上进行namespace隔离

最后要说明的系统调用是unshare()，它与clone()很像，不同的是，unshare()运行在原先的进程上，不需要启动一个新进程。
调用unshare()的主要作用就是，不启动新进程就可以起到隔离的作用，相当于跳出原先的namespace进行操作，这样，就可以在原进程进行了一些需要隔离的操作。linux中自带unshare命令，就是通过unshare()系统调用实现的，docker目前并没有使用这个系统调用。

- fork()系统调用

系统调用函数fork()并不属于namespace的API，当程序调用fork()函数时，系统会创建新的进程，为其分配资源，例如存储数据和代码的空间，然后把原来进程的所有值复制到新的进程中，只有少量数值与原来的进程不同，相当于复制了本身。那么程序的后续代码逻辑要如何区分是子进程还是父进程呢？
fork()的神奇之处在于它被调用一次，却能返回两次(父进程与子进程各返回一次)，通过返回值的不同就可以区分父进程与子进程。他可能有以下3种不同的返回值：

1. 在父进程中，fork()返回新创建子进程的进程id；
2. 在子进程中，fork()返回0；
3. 如果出现问题，fork()返回一个负值。

使用fork()后，父进程有义务监控子进程的运行状态，并在子进程推出后才能正常退出，否则子进程就会成为“孤儿”进程

下面将根据docker内部对namespace资源隔离使用方式分别对6种namespace进行解析。

### 2.UTS namespace
UTS(UNIX Time-sharing System)namespace提供了主机名与域名的隔离，这样每个docke容器就可以拥有独立的主机名和域名了，在网络上可以被视为一个独立的节点，而非宿主机上的一个进程。docker中，每个镜像基本都以自身所提供的服务名称来命名镜像的hostname，且不会对宿主机产生任何影响，其原理就是使用了UTS namespace

### 3.IPC namespace
进程间通信(Inter-Process Communication，IPC)涉及的IPC资源包括常见的信号量、消息队列和共享内存。申请IPC资源就申请了一个全局唯一的32位ID，所以IPC namespace中实际上包含了系统IPC标识符以及实现POSIX消息队列的文件系统。在同一个IPC namespace下的进程彼此可见，不同IPC namespace下的进程则互相不可见。
目前使用IPC namespace机制的系统不多，其中比较有名的有PostgreSQL。Docker当前也使用IPC namespace实现了容器与宿主机、容器与容器之间的IPC隔离。

### 4.PID namespace
PID namespace隔离非常实用，它对进程PID重新标号，即两个不同namespace下的进程可以有相同的PID。每个PID namespace都有自己的计数程序。内核为所有的PID namespace维护了一个树状结构，最顶层的是系统初始时创建的，被称为root namespace，它创建的心PID namespace被称为child namespace(树的子节点)，洱源县的PID namespace就是新创建的PID namespace的parent namespace(树的父节点)。通过这种方式，不同的PID namespace会形成一个层级体系。所属的父节点可以看到子节点中的进程，并可以通过信号等方式对子节点中的进程产生影响。反过来，子节点却不能看到父节点PID namespace中的任何内容，由此产生如下结论。

- 每个PID namespace中的第一个进程“PID 1”，都会像全通Linux中的init进程一样拥有特权，其特殊作用。
- 一个namespace中的进程，不可能通过kill或ptrace影响父节点或者兄弟节点中的进程，因为其他几点的PID在这个namespace没有任何意义。
- 如果你在新的PID namespace中重新挂载/proc文件系统，会发现其下只显示同属一个PID namespace中的其他进程。
- 在root namespace中看到所有的进程，并且递归包含所有子节点中的进程。到这里，读者可能已经联想到了一种在Docker外部监控运行程序的方法了，就是监控Docker daemon所在的PID namespace下的所有进程及子进程，在进行筛选即可。

**PID namespace中的init进程**

在传统的Unix系统中，PID为1的进程时init，地位非常特殊。它作为所有进程的父进程，维护一张进程表，不断检查进程状态，一旦某个子进程因为父进程错误成为了“孤儿”进程，init就会负责收养这个子进程并最终回收资源，结束进程。所以在要实现的容器中，启动的第一个进程也需要实现类似init的功能，维护所有后续启动进程的状态。
当系统中存在的树状嵌套结构的PID namespace时，若某个子进程成为了孤儿进程，收养孩子进程的责任就交给了孩子进程所属的PID namespace中的init进程。
至此，读者可以明白内核设计的良苦用心。PID namespace维护这样一个树状结构，有利于系统的资源的控制与回收。因此，如果确实需要在一个Docker容器中运行多个进程，最先启动的命令进程应该是具有资源监控与回收等管理能力的，如bash。

**信号与init进程**
内核还为PID namespace中的init进程赋予了其他特权—信号屏蔽。如果init中没有编写处理某个信号的代码逻辑，那么与init在同一个PID namespace下的进程(即使有超级权限)发送非他的信号都会屏蔽。这个功能主要作用就是防止init进程被误杀。
那么，父节点PID namespace中的进程发送同样的信号给子节点中的init的进程，这会被忽略吗？父节点中的进程发送的信号，如果不是SIGKILL(销毁进程)或SIGSTOPO(暂停进程)也会诶忽略。但如果发送SIGKILL或SIGSTOP，子节点的init会强制执行(无法通过代码捕捉进行特殊处理)，也就是说父节点中的进程有权终止子进程。
一旦init进程被销毁，同一PID namespace中的其他进程也所致接收到SIGKIKLL信号而被销毁。理论上，该PID namespace也不复存在了。但如果/proc/[pid]/ns/pid处于被挂载或打开的状态，namespace就会被保留下来。然而，被保留下来的namespace无法通过setns()或者fork()创建进程，所以实际上并没有什么作用。
当一个容器内存在多个进程时，容器内的init进程可以对信号进行捕获，当SIGTERM或SIGINT等信号到来时，对其子进程做信息保存、资源回收等处理工作。在Docker daemon的源码中也可以看到类似的处理方式，当结束信号来临时，结束容器进程并回收相应资源。

**挂载proc文件系统**
前文提到，如果在新的PID namespace中使用使用ps命令查看，看到的还是所有的进程，因为与PID直接相关的/proc文件系统(procfs)没有挂载到一个与原/proc不同的位置。如果只想看到PID namespace本身应该看到的进程，需要重新挂载/proc，命令如下。

```$ mount -t proc proc /proc
$ ps a
```
- unshare()和setns()
本文开头就谈到了unshare()和setns()这两个API，在PID namespace中使用，也有一些特别之处需要注意。
unshare()允许用户在原有进程中建立命名空间进行隔离。但创建了PID namespace后，原先unshare()调用者进程并不进入新的PID namespace，接下来创建的子进程才会进入新的namespace，这个子进程也就随之成为新的namespace中的init进程。
类似地，调用setns()创建新PID namespace时，调用者进程也不进入新的PID namespace，而是随后创建的子进程进入。
为什么创建其他namespace时unshare()和setns()会直接进入新的namespace，二唯独PID namespace例外呢？因为调用getpid()函数得到的PID是根据调用者所在的PID namespace而决定返回哪个PID，进入新的PID namespace会导致PID产生变化。而对用户态的程序和库函数来说，他们都认为进程的PID是一个常量，PID的变化会引起这些进程崩溃。
换句话说，一旦程序进程创建以后，那么它的PID namespace的关系就确定下来了，进程不会变更它们对应的PID namespace。在Docker中，docker exec会使用setns()函数加入已经存在的命名空间，但是最终还是会调用clone()函数，原因就在于此。

mount namespace
mount namespace通过隔离文件系统挂载点对隔离文件系统提供支持，它是历史上第一个Linux namespace，所以标示位比较特殊，就是CLONE_NEWNS。隔离后，不同的mount namespace中的文件结构发生变化也互不影响。也可以通过/proc/[pid]/mounts查看到所有挂载在当前namespace中的文件系统，还可以通过/proc/[pid]/mountstats看到mount namespace中文件设备的统计信息，包括挂载文件的名字、文件系统的类型、挂载位置等。
进程在创建mount namespace时，会把当前的文件结构复制给新的namespace。新namespace中的所有mount操作都只影响自身的文件系统，对外界不会产生任何影响。这种做法非常严格的实现了隔离，但对某些状况可能并不适用。比如父节点namespace中的进程挂载了一张CD-ROM，这时子节点namespace复制的目录结构是无法自动挂载上这张CD-ROM的，因为这种操作会影响到父节点的文件系统。

一个挂载状态可能为以下一种：

共享挂载
从属挂载
共享/从属挂载
私有挂载
不可绑定挂载
传播事件的挂载对象称为共享挂载；接收传播事件的挂载对象称为从属挂载；同时兼有前述两者特征的挂载对象为共享/从属挂载；既不传播也不接受事件的挂载对象称为私有挂载；另一种特殊的挂载对象称为不可绑定挂载，它们与私有挂载相似，但不允许执行绑定挂载，即创建mount namespace时这块文件对象不可被复制。

## 6.netword namespace
network namespace主要提供了关于网络资源的隔离，包括网络设备、IPv4和IPv6协议栈、IP路由表、防火墙、/proc/net目录、/sys/class/net目录、socket等。一个物理的网络设备最多存在于一个network namespace中，可以通过创建veth pair(虚拟网络设备对：有两端，类似管道，如果数据从一端传入另一端也能接受，反之亦然)在不同的network namespace间创建通道，以达到通信目的。
也许你会好奇，在建立起veth pair之前，新旧namespace该如何通信呢？答案是pipe(管道)。以Docker daemon启动容器的过程为例，假设容器内初始化的进程称为init。Docker daemon在宿主机上负责创建这个veth pair，把一段绑定到docker0网桥上，另一端介入新建的network namespace进程中。这个过程执行期间，Docker daemon和init就通过pipe进行通信。具体来说，就是在Docker deamon完成veth pair的创建之前，init在管道的另一端循环等待，直到管道另一端传来Docker daemon关于veth设备的信息，并关闭管道。init才结束等待的过程，并把它的“eth0”启动起来。
与其他namespace类似，对network namespace的使用其实就是在创建的时候添加CLONE_NEWNET标识符位。

## 7.user namespace
user namespace主要隔离了安全相关的标识符(identifier)和属性(attribute)，包括用户ID、用户组ID、root目录、key(指密钥)以及特殊权限。通俗地讲，一个普通用户的进程通过clone()创建的新进程在新user namespace中可以拥有不同的用户和用户组。这意味着一个进程在容器外属于一个没有特权的普通用户，但是它创建的容器进程却属于拥有所有权限的超级用户，这个技术为容器提供了极大的自由。
user namespace时目前的6个namespace中最后一个支持的，并且直到linux内核3.8版本的时候还未完全实现(还有部分文件系统不支持)。user namespace实际上并不算完全成熟，很多发行版担心安全问题，在编译内核的时候并未开启USER_NS。Docker在1.10版本中对user namespace进行了支持。只要用户在启动Docker daemon的时候制定了–user-remap，那么当用户运行容器时，容器内部的root用户并不等于宿主机的root用户，而是映射到宿主机上的普通用户。
Docker不仅使用了user namespace，还使用了在user namespace中涉及的Capability机制。从内核2.2版本开始，Linux把原来和超级用户相关的高级权限分为不同的单元，称为Capability。这样管理员就可以独立的对特定的Capability进行使用或禁止。Docker同时使用namespace和Capability，这很大程度上加强了容器的安全性。