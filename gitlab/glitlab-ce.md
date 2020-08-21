gitlab-ce chart requires:
 - Redis 
 - Postgresql

gitlab-ce 必配参数：
- entrenalUrl: https://gitlab-xx.xxx.com (指定gitlab拉代码的url,一旦配置成https需要上传crt,key到指定目录)

gitlab安装错误问题处理：
1,serviceaccount没有权限问题
```
==> /var/log/gitlab/prometheus/current <==
2018-12-24_03:06:08.88786 level=error ts=2018-12-24T03:06:08.887812767Z caller=main.go:240 component=k8s_client_runtime err="github.com/prometheus/prometheus/discovery/kubernetes/kubernetes.go:372: Failed to list *v1.Node: nodes is forbidden: User \"system:serviceaccount:default:default\" cannot list resource \"nodes\" in API group \"\" at the cluster scope"
2018-12-24_03:06:08.89075 level=error ts=2018-12-24T03:06:08.890719525Z caller=main.go:240 component=k8s_client_runtime err="github.com/prometheus/prometheus/discovery/kubernetes/kubernetes.go:320: Failed to list *v1.Pod: pods is forbidden: User \"system:serviceaccount:default:default\" cannot list resource \"pods\" in API group \"\" at the cluster scope"
```
原因：default service account 没有权限获取集群的nodes和pods。重新创建一个CluserterRole和clusterRoleBinding
```
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: prom-admin
rules:
# Just an example, feel free to change it
- apiGroups: [""]
  resources: ["pods", "nodes"]
  verbs: ["get", "watch", "list"]

---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: prom-rbac
subjects:
- kind: ServiceAccount
  name: default
roleRef:
  kind: ClusterRole
  name: prom-admin
  apiGroup: rbac.authorization.k8s.io
```

2,entrenalUrl使用https,未上传crt和key
```
=>/var/log/gitlab/nigix/error.log<=
BIO_new_file("/etc/gitlab/ssl/gitlab-test.xxx.com.crt") failed (SSL:error:02001002:system libarry:fopen:no such file or directory:fopen('/etc/gitlab/ssl/gitlab-test.xxx.com.crt','r'))
```
解决办法：把crt和key做成configMap挂载到对应目录下。(这个crt一定要是是trusted certificate,open ssl制作)

