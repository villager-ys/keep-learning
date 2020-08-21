1，k8s调度的pod跑构建任务，无法使用kubectl命令问题

解决方法：挂载kubectl的.kube文件

![image](./images/kubeconfig.png)

2，调用apiserver报错
![image](./images/apiserver-error.png)

jenkins默认工作目录是/home/jenkins，.kube目录挂载错误，可以在执行kubectl命令时指定.kube文件目录，或者更换卷路径

![image](./images/kubeconf.png)

