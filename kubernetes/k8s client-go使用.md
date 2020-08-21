### 一，代码示例解析
```
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
	"time"

	//
	// Uncomment to load all auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	//
	//Or uncomment to load specific auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	_ "k8s.io/client-go/plugin/pkg/client/auth/openstack"
)

func main() {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"),
			"(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	pods, err := clientset.CoreV1().Pods("devops-department-qa").List(metav1.ListOptions{LabelSelector: "app=workflow-api"})
	if err != nil {
		panic(err.Error())
	}
	podlist := pods.Items
	data, _ := json.Marshal(podlist[0])
	fmt.Println(string(data))
	var x v1.PodLogOptions = v1.PodLogOptions{Container: "cdn-adapter-api"}

	log := clientset.CoreV1().Pods("devops-department-qa").GetLogs("cdn-adapter-api-849bf764b6-kf4s2",
		&x)
	result := log.Timeout(5 * time.Second).Do()
	bytes, _ := result.Raw()
	fmt.Println("--------")
	fmt.Println(string(bytes))
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
```
解析：
1. kubeconfig
```
kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"),
			"(optional) absolute path to the kubeconfig file")
```
获取kubernetes配置文件kubeconfig的绝对路径。一般路径为$HOME/.kube/config。该文件主要用来配置本地连接的kubernetes集群。
2. 获取rest.config
```
config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
```
BuildConfigFromFlags函数源码
```
// BuildConfigFromFlags is a helper function that builds configs from a master
// url or a kubeconfig filepath. These are passed in as command line flags for cluster
// components. Warnings should reflect this usage. If neither masterUrl or kubeconfigPath
// are passed in we fallback to inClusterConfig. If inClusterConfig fails, we fallback
// to the default config.
func BuildConfigFromFlags(masterUrl, kubeconfigPath string) (*restclient.Config, error) {
	if kubeconfigPath == "" && masterUrl == "" {
		klog.Warningf("Neither --kubeconfig nor --master was specified.  Using the inClusterConfig.  This might not work.")
		kubeconfig, err := restclient.InClusterConfig()
		if err == nil {
			return kubeconfig, nil
		}
		klog.Warning("error creating inClusterConfig, falling back to default config: ", err)
	}
	return NewNonInteractiveDeferredLoadingClientConfig(
		&ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&ConfigOverrides{ClusterInfo: clientcmdapi.Cluster{Server: masterUrl}}).ClientConfig()
}
```
3. clientset
```
clientset, err := kubernetes.NewForConfig(config)
```
通过*rest.Config参数和NewForConfig方法来获取clientset对象，clientset是多个client的集合，每个client可能包含不同版本的方法调用。

NewForConfig函数就是初始化clientset中的每个client。
```
func NewForConfig(c *rest.Config) (*Clientset, error) {
	configShallowCopy := *c
	if configShallowCopy.RateLimiter == nil && configShallowCopy.QPS > 0 {
		configShallowCopy.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(configShallowCopy.QPS, configShallowCopy.Burst)
	}
	var cs Clientset
	var err error
	cs.admissionregistrationV1beta1, err = admissionregistrationv1beta1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.appsV1, err = appsv1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.appsV1beta1, err = appsv1beta1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.appsV1beta2, err = appsv1beta2.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.auditregistrationV1alpha1, err = auditregistrationv1alpha1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.authenticationV1, err = authenticationv1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.authenticationV1beta1, err = authenticationv1beta1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.authorizationV1, err = authorizationv1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.authorizationV1beta1, err = authorizationv1beta1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.autoscalingV1, err = autoscalingv1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.autoscalingV2beta1, err = autoscalingv2beta1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.autoscalingV2beta2, err = autoscalingv2beta2.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.batchV1, err = batchv1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.batchV1beta1, err = batchv1beta1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.batchV2alpha1, err = batchv2alpha1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.certificatesV1beta1, err = certificatesv1beta1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.coordinationV1beta1, err = coordinationv1beta1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.coordinationV1, err = coordinationv1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.coreV1, err = corev1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.eventsV1beta1, err = eventsv1beta1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.extensionsV1beta1, err = extensionsv1beta1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.networkingV1, err = networkingv1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.networkingV1beta1, err = networkingv1beta1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.nodeV1alpha1, err = nodev1alpha1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.nodeV1beta1, err = nodev1beta1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.policyV1beta1, err = policyv1beta1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.rbacV1, err = rbacv1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.rbacV1beta1, err = rbacv1beta1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.rbacV1alpha1, err = rbacv1alpha1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.schedulingV1alpha1, err = schedulingv1alpha1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.schedulingV1beta1, err = schedulingv1beta1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.schedulingV1, err = schedulingv1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.settingsV1alpha1, err = settingsv1alpha1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.storageV1beta1, err = storagev1beta1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.storageV1, err = storagev1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.storageV1alpha1, err = storagev1alpha1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}

	cs.DiscoveryClient, err = discovery.NewDiscoveryClientForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	return &cs, nil
}
```
4. clientset.Interface
```
pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
```
clientset接口
```
type Interface interface {
	Discovery() discovery.DiscoveryInterface
	AdmissionregistrationV1beta1() admissionregistrationv1beta1.AdmissionregistrationV1beta1Interface
	AppsV1() appsv1.AppsV1Interface
	AppsV1beta1() appsv1beta1.AppsV1beta1Interface
	AppsV1beta2() appsv1beta2.AppsV1beta2Interface
	AuditregistrationV1alpha1() auditregistrationv1alpha1.AuditregistrationV1alpha1Interface
	AuthenticationV1() authenticationv1.AuthenticationV1Interface
	AuthenticationV1beta1() authenticationv1beta1.AuthenticationV1beta1Interface
	AuthorizationV1() authorizationv1.AuthorizationV1Interface
	AuthorizationV1beta1() authorizationv1beta1.AuthorizationV1beta1Interface
	AutoscalingV1() autoscalingv1.AutoscalingV1Interface
	AutoscalingV2beta1() autoscalingv2beta1.AutoscalingV2beta1Interface
	AutoscalingV2beta2() autoscalingv2beta2.AutoscalingV2beta2Interface
	BatchV1() batchv1.BatchV1Interface
	BatchV1beta1() batchv1beta1.BatchV1beta1Interface
	BatchV2alpha1() batchv2alpha1.BatchV2alpha1Interface
	CertificatesV1beta1() certificatesv1beta1.CertificatesV1beta1Interface
	CoordinationV1beta1() coordinationv1beta1.CoordinationV1beta1Interface
	CoordinationV1() coordinationv1.CoordinationV1Interface
	CoreV1() corev1.CoreV1Interface
	EventsV1beta1() eventsv1beta1.EventsV1beta1Interface
	ExtensionsV1beta1() extensionsv1beta1.ExtensionsV1beta1Interface
	NetworkingV1() networkingv1.NetworkingV1Interface
	NetworkingV1beta1() networkingv1beta1.NetworkingV1beta1Interface
	NodeV1alpha1() nodev1alpha1.NodeV1alpha1Interface
	NodeV1beta1() nodev1beta1.NodeV1beta1Interface
	PolicyV1beta1() policyv1beta1.PolicyV1beta1Interface
	RbacV1() rbacv1.RbacV1Interface
	RbacV1beta1() rbacv1beta1.RbacV1beta1Interface
	RbacV1alpha1() rbacv1alpha1.RbacV1alpha1Interface
	SchedulingV1alpha1() schedulingv1alpha1.SchedulingV1alpha1Interface
	SchedulingV1beta1() schedulingv1beta1.SchedulingV1beta1Interface
	SchedulingV1() schedulingv1.SchedulingV1Interface
	SettingsV1alpha1() settingsv1alpha1.SettingsV1alpha1Interface
	StorageV1beta1() storagev1beta1.StorageV1beta1Interface
	StorageV1() storagev1.StorageV1Interface
	StorageV1alpha1() storagev1alpha1.StorageV1alpha1Interface
}
```

```
// List takes label and field selectors, and returns the list of Pods that match those selectors.
func (c *pods) List(opts metav1.ListOptions) (result *v1.PodList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1.PodList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("pods").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do().
		Into(result)
	return
}
```
5. RESTClient

RESTClient对象的创建同样是依赖传入的config信息。
```
client, err := rest.RESTClientFor(&config)
```
rest.RESTClientFor
```
// RESTClientFor returns a RESTClient that satisfies the requested attributes on a client Config
// object. Note that a RESTClient may require fields that are optional when initializing a Client.
// A RESTClient created by this method is generic - it expects to operate on an API that follows
// the Kubernetes conventions, but may not be the Kubernetes API.
func RESTClientFor(config *Config) (*RESTClient, error) {
    ...
    qps := config.QPS
    ...
    burst := config.Burst
    ...
    baseURL, versionedAPIPath, err := defaultServerUrlFor(config)
    ...
    transport, err := TransportFor(config)
    ...
    var httpClient *http.Client
    if transport != http.DefaultTransport {
        httpClient = &http.Client{Transport: transport}
        if config.Timeout > 0 {
            httpClient.Timeout = config.Timeout
        }
    }

    return NewRESTClient(baseURL, versionedAPIPath, config.ContentConfig, qps, burst, config.RateLimiter, httpClient)
}
```
RESTClientFor函数调用了NewRESTClient的初始化函数。
```
// NewRESTClient creates a new RESTClient. This client performs generic REST functions
// such as Get, Put, Post, and Delete on specified paths.  Codec controls encoding and
// decoding of responses from the server.
func NewRESTClient(baseURL *url.URL, versionedAPIPath string, config ContentConfig, maxQPS float32, maxBurst int, rateLimiter flowcontrol.RateLimiter, client *http.Client) (*RESTClient, error) {
    base := *baseURL
    ...
    serializers, err := createSerializers(config)
    ...
    return &RESTClient{
        base:             &base,
        versionedAPIPath: versionedAPIPath,
        contentConfig:    config,
        serializers:      *serializers,
        createBackoffMgr: readExpBackoffConfig,
        Throttle:         throttle,
        Client:           client,
    }, nil
}
```
RESTClient结构体:

RESTClient结构体中包含了http.Client，即本质上RESTClient就是一个http.Client的封装实现。
```
// RESTClient imposes common Kubernetes API conventions on a set of resource paths.
// The baseURL is expected to point to an HTTP or HTTPS path that is the parent
// of one or more resources.  The server should return a decodable API resource
// object, or an api.Status object which contains information about the reason for
// any failure.
//
// Most consumers should use client.New() to get a Kubernetes API client.
type RESTClient struct {
    // base is the root URL for all invocations of the client
    base *url.URL
    // versionedAPIPath is a path segment connecting the base URL to the resource root
    versionedAPIPath string

    // contentConfig is the information used to communicate with the server.
    contentConfig ContentConfig

    // serializers contain all serializers for underlying content type.
    serializers Serializers

    // creates BackoffManager that is passed to requests.
    createBackoffMgr func() BackoffManager

    // TODO extract this into a wrapper interface via the RESTClient interface in kubectl.
    Throttle flowcontrol.RateLimiter

    // Set specific behavior of the client.  If not set http.DefaultClient will be used.
    Client *http.Client
}
```
RESTClient实现了以下的接口方法：
```
// Interface captures the set of operations for generically interacting with Kubernetes REST apis.
type Interface interface {
    GetRateLimiter() flowcontrol.RateLimiter
    Verb(verb string) *Request
    Post() *Request
    Put() *Request
    Patch(pt types.PatchType) *Request
    Get() *Request
    Delete() *Request
    APIVersion() schema.GroupVersion
}
```
在调用HTTP方法（Post()，Put()，Get()，Delete() ）时，实际上调用了Verb(verb string)函数。
```
// Verb begins a request with a verb (GET, POST, PUT, DELETE).
//
// Example usage of RESTClient's request building interface:
// c, err := NewRESTClient(...)
// if err != nil { ... }
// resp, err := c.Verb("GET").
//  Path("pods").
//  SelectorParam("labels", "area=staging").
//  Timeout(10*time.Second).
//  Do()
// if err != nil { ... }
// list, ok := resp.(*api.PodList)
//
func (c *RESTClient) Verb(verb string) *Request {
    backoff := c.createBackoffMgr()

    if c.Client == nil {
        return NewRequest(nil, verb, c.base, c.versionedAPIPath, c.contentConfig, c.serializers, backoff, c.Throttle)
    }
    return NewRequest(c.Client, verb, c.base, c.versionedAPIPath, c.contentConfig, c.serializers, backoff, c.Throttle)
}
```
Verb函数调用了NewRequest方法，最后调用Do()方法实现一个HTTP请求获取Result。

6. 总结

client-go对kubernetes资源对象的调用，需要先获取kubernetes的配置信息，即$HOME/.kube/config。

整个调用的过程如下：

kubeconfig→rest.config→clientset→具体的client(CoreV1Client)→具体的资源对象(pod)→RESTClient→http.Client→HTTP请求的发送及响应

通过clientset中不同的client和client中不同资源对象的方法实现对kubernetes中资源对象的增删改查等操作，常用的client有CoreV1Client、AppsV1beta1Client、ExtensionsV1beta1Client等。

### 二，client-go对k8s资源的调用
- deployment
```
//声明deployment对象
var deployment *v1beta1.Deployment
//构造deployment对象
//创建deployment
deployment, err := clientset.AppsV1beta1().Deployments(<namespace>).Create(<deployment>)
//更新deployment
deployment, err := clientset.AppsV1beta1().Deployments(<namespace>).Update(<deployment>)
//删除deployment
err := clientset.AppsV1beta1().Deployments(<namespace>).Delete(<deployment.Name>, &meta_v1.DeleteOptions{})
//查询deployment
deployment, err := clientset.AppsV1beta1().Deployments(<namespace>).Get(<deployment.Name>, meta_v1.GetOptions{})
//列出deployment
deploymentList, err := clientset.AppsV1beta1().Deployments(<namespace>).List(&meta_v1.ListOptions{})
//watch deployment
watchInterface, err := clientset.AppsV1beta1().Deployments(<namespace>).Watch(&meta_v1.ListOptions{})
```
- service
```
//声明service对象
var service *v1.Service
//构造service对象
//创建service
service, err := clientset.CoreV1().Services(<namespace>).Create(<service>)
//更新service
service, err := clientset.CoreV1().Services(<namespace>).Update(<service>)
//删除service
err := clientset.CoreV1().Services(<namespace>).Delete(<service.Name>, &meta_v1.DeleteOptions{})
//查询service
service, err := clientset.CoreV1().Services(<namespace>).Get(<service.Name>, meta_v1.GetOptions{})
//列出service
serviceList, err := clientset.CoreV1().Services(<namespace>).List(&meta_v1.ListOptions{})
//watch service
watchInterface, err := clientset.CoreV1().Services(<namespace>).Watch(&meta_v1.ListOptions{})
```
- ingress
```
//声明ingress对象
var ingress *v1beta1.Ingress
//构造ingress对象
//创建ingress
ingress, err := clientset.ExtensionsV1beta1().Ingresses(<namespace>).Create(<ingress>)
//更新ingress
ingress, err := clientset.ExtensionsV1beta1().Ingresses(<namespace>).Update(<ingress>)
//删除ingress
err := clientset.ExtensionsV1beta1().Ingresses(<namespace>).Delete(<ingress.Name>, &meta_v1.DeleteOptions{})
//查询ingress
ingress, err := clientset.ExtensionsV1beta1().Ingresses(<namespace>).Get(<ingress.Name>, meta_v1.GetOptions{})
//列出ingress
ingressList, err := clientset.ExtensionsV1beta1().Ingresses(<namespace>).List(&meta_v1.ListOptions{})
//watch ingress
watchInterface, err := clientset.ExtensionsV1beta1().Ingresses(<namespace>).Watch(&meta_v1.ListOptions{})
```
- pod
```
//声明pod对象
var pod *v1.Pod
//创建pod
pod, err := clientset.CoreV1().Pods(<namespace>).Create(<pod>)
//更新pod
pod, err := clientset.CoreV1().Pods(<namespace>).Update(<pod>)
//删除pod
err := clientset.CoreV1().Pods(<namespace>).Delete(<pod.Name>, &meta_v1.DeleteOptions{})
//查询pod
pod, err := clientset.CoreV1().Pods(<namespace>).Get(<pod.Name>, meta_v1.GetOptions{})
//列出pod
podList, err := clientset.CoreV1().Pods(<namespace>).List(&meta_v1.ListOptions{})
//watch pod
watchInterface, err := clientset.CoreV1().Pods(<namespace>).Watch(&meta_v1.ListOptions{})
```
- container日志
```
	var x v1.PodLogOptions = v1.PodLogOptions{Container: "cdn-adapter-api"}

	log := clientset.CoreV1().Pods("devops-department-qa").GetLogs("cdn-adapter-api-849bf764b6-kf4s2",
		&x)
	result := log.Timeout(5 * time.Second).Do()
	bytes, _ := result.Raw()
```
- event事件

event事件的name格式是:pod名+时间戳，所以要获取pod的事件可以获取的是整个ns的，然后过滤你需要的对应pod的事件。

```
	event, err := clientset.CoreV1().Events("wlx-dev").List(metav1.ListOptions{})
	if err != nil {
		fmt.Println(err)
		return
	}
```

- hpa
```
	get, err := clientset.AutoscalingV1().HorizontalPodAutoscalers("devops-department-qa").Get("workflow-activiti", metav1.GetOptions{})
	if err != nil {
		fmt.Println(err)
	}
	data, err := json.Marshal(get)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(data))
```