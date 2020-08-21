### 1，解决not enough arguments in call to watch.NewStreamWatcher
github issue:https://github.com/kubernetes/client-go/issues/691

原因：
1. 当我们使用client-go v11.x时，要注意同时使用kubernetes-1.14 levels of k8s.io/apimachinery and k8s.io/api
```
go get k8s.io/client-go@kubernetes-1.15.0

```
```
require (
k8s.io/api kubernetes-1.14.7
k8s.io/apimachinery kubernetes-1.14.7
k8s.io/client-go kubernetes-1.14.7
)
```
2. 详细说明
https://github.com/kubernetes/client-go/blob/master/INSTALL.md

### 2，报错Azure go-autorest causing ambiguous import
github issue:https://github.com/kubernetes/client-go/issues/628
```
build github.com/.../cmd/operator: cannot load github.com/Azure/go-autorest/autorest: ambiguous import: found github.com/Azure/go-autorest/autorest in multiple modules:
	github.com/Azure/go-autorest v11.1.2+incompatible (/.../github.com/!azure/go-autorest@v11.1.2+incompatible/autorest)
	github.com/Azure/go-autorest/autorest v0.3.0 (/.../github.com/!azure/go-autorest/autorest@v0.3.0)
```

This occurs because:

v11.1.2 -> doesn't have github.com/Azure/go-autorest/autorest{go.mod,go.sum} files.

autorest/v0.3.0 -> has github.com/Azure/go-autorest/autorest{go.mod,go.sum} files.

解决办法
```
require github.com/Azure/go-autorest v12.2.0+incompatible
```