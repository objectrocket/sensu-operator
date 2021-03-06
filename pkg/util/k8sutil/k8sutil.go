// Copyright 2016 The etcd-operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package k8sutil

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	api "github.com/objectrocket/sensu-operator/pkg/apis/objectrocket/v1beta1"
	"github.com/objectrocket/sensu-operator/pkg/util/etcdutil"
	"github.com/objectrocket/sensu-operator/pkg/util/retryutil"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	utilrand "k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // for gcp auth
	"k8s.io/client-go/rest"
)

const (
	// EtcdClientPort is the client port on client service and etcd nodes.
	EtcdClientPort = 2379

	etcdVolumeMountDir        = "/var/lib/sensu/etcd"
	stateDir                  = "/var/lib/sensu"
	dataDir                   = etcdVolumeMountDir + "/data"
	backupFile                = "/var/lib/sensu/etcd/latest.backup"
	sensuVersionAnnotationKey = "sensu.version"
	peerTLSDir                = "/etc/etcdtls/member/peer-tls"
	peerTLSVolume             = "member-peer-tls"
	serverTLSDir              = "/etc/etcdtls/member/server-tls"
	serverTLSVolume           = "member-server-tls"
	operatorEtcdTLSDir        = "/etc/etcdtls/operator/etcd-tls"
	operatorEtcdTLSVolume     = "etcd-client-tls"

	randomSuffixLength = 10
	// k8s object name has a maximum length
	maxNameLength = 63 - randomSuffixLength - 1

	defaultBusyboxImage = "busybox:1.28.0-glibc"

	// AnnotationScope annotation name for defining instance scope. Used for specifing cluster wide clusters.
	AnnotationScope = "objectrocket.com/scope"
	//AnnotationClusterWide annotation value for cluster wide clusters.
	AnnotationClusterWide = "clusterwide"

	// defaultDNSTimeout is the default maximum allowed time for the init container of the etcd pod
	// to reverse DNS lookup its IP. The default behavior is to wait forever and has a value of 0.
	defaultDNSTimeout = int64(0)
)

type label struct {
	key   string
	value string
}

const TolerateUnreadyEndpointsAnnotation = "service.alpha.kubernetes.io/tolerate-unready-endpoints"

func ptrInt64(i int64) *int64 {
	return &i
}

func ptrBool(b bool) *bool {
	return &b
}

func GetSensuVersion(pod *v1.Pod) string {
	return pod.Annotations[sensuVersionAnnotationKey]
}

func SetSensuVersion(pod *v1.Pod, version string) {
	pod.Annotations[sensuVersionAnnotationKey] = version
}

func SetPodTemplateSensuVersion(pod *v1.PodTemplateSpec, version string) {
	pod.Annotations[sensuVersionAnnotationKey] = version
}

func GetPodNames(pods []*v1.Pod) []string {
	if len(pods) == 0 {
		return nil
	}
	res := []string{}
	for _, p := range pods {
		res = append(res, p.Name)
	}
	return res
}

// PVCNameFromMember the way we get PVC name from the member name
func PVCNameFromMember(memberName string) string {
	return memberName
}

func ImageName(repo, version string) string {
	return fmt.Sprintf("%s:%v", repo, version)
}

// imageNameBusybox returns the default image for busybox init container, or the image specified in the PodPolicy
func imageNameBusybox(policy *api.PodPolicy) string {
	if policy != nil && len(policy.BusyboxImage) > 0 {
		return policy.BusyboxImage
	}
	return defaultBusyboxImage
}

func PodWithNodeSelector(p *v1.PodTemplateSpec, ns map[string]string) *v1.PodTemplateSpec {
	p.Spec.NodeSelector = ns
	return p
}

func DashboardServiceName(clusterName string) string {
	return clusterName + "-dashboard"
}

func CreateDashboardService(kubecli kubernetes.Interface, clusterName, ns string, owner metav1.OwnerReference) error {
	ports := []v1.ServicePort{{
		Name:       "dashboard",
		Port:       3000,
		TargetPort: intstr.FromInt(3000),
		Protocol:   v1.ProtocolTCP,
	}}
	return createService(kubecli, DashboardServiceName(clusterName), clusterName, ns, "", ports, owner)
}

func AgentServiceName(clusterName string) string {
	return clusterName + "-agent"
}

func CreateAgentService(kubecli kubernetes.Interface, clusterName, ns string, owner metav1.OwnerReference) error {
	ports := []v1.ServicePort{{
		Name:       "agent",
		Port:       8081,
		TargetPort: intstr.FromInt(8081),
		Protocol:   v1.ProtocolTCP,
	}}
	return createService(kubecli, AgentServiceName(clusterName), clusterName, ns, "", ports, owner)
}

func APIServiceName(clusterName string) string {
	return fmt.Sprintf("%s-api", clusterName)
}

func CreateAPIService(kubecli kubernetes.Interface, clusterName, ns string, owner metav1.OwnerReference) error {
	ports := []v1.ServicePort{{
		Name:       "api",
		Port:       8080,
		TargetPort: intstr.FromInt(8080),
		Protocol:   v1.ProtocolTCP,
	}}
	return createService(kubecli, APIServiceName(clusterName), clusterName, ns, "", ports, owner)
}

func CreatePeerService(kubecli kubernetes.Interface, clusterName, ns string, owner metav1.OwnerReference) error {
	ports := []v1.ServicePort{{
		Name:       "client",
		Port:       EtcdClientPort,
		TargetPort: intstr.FromInt(EtcdClientPort),
		Protocol:   v1.ProtocolTCP,
	}, {
		Name:       "peer",
		Port:       2380,
		TargetPort: intstr.FromInt(2380),
		Protocol:   v1.ProtocolTCP,
	}}

	return createService(kubecli, clusterName, clusterName, ns, v1.ClusterIPNone, ports, owner)
}

func createService(kubecli kubernetes.Interface, svcName, clusterName, ns, clusterIP string, ports []v1.ServicePort, owner metav1.OwnerReference) error {
	svc := newSensuServiceManifest(svcName, clusterName, clusterIP, ports)
	addOwnerRefToObject(svc.GetObjectMeta(), owner)
	_, err := kubecli.CoreV1().Services(ns).Create(svc)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

// CreateAndWaitPod creates a pod and waits until it is running
func CreateAndWaitPod(kubecli kubernetes.Interface, ns string, pod *v1.Pod, timeout time.Duration) (*v1.Pod, error) {
	_, err := kubecli.CoreV1().Pods(ns).Create(pod)
	if err != nil {
		return nil, err
	}

	interval := 5 * time.Second
	var retPod *v1.Pod
	err = retryutil.Retry(interval, int(timeout/(interval)), func() (bool, error) {
		retPod, err = kubecli.CoreV1().Pods(ns).Get(pod.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		switch retPod.Status.Phase {
		case v1.PodRunning:
			return true, nil
		case v1.PodPending:
			return false, nil
		default:
			return false, fmt.Errorf("unexpected pod status.phase: %v", retPod.Status.Phase)
		}
	})

	if err != nil {
		if retryutil.IsRetryFailure(err) {
			return nil, fmt.Errorf("failed to wait pod running, it is still pending: %v", err)
		}
		return nil, fmt.Errorf("failed to wait pod running: %v", err)
	}

	return retPod, nil
}

// CreateAndWaitDeployment creates a deployment and waits until the defined
// number of replicas is reached
func CreateAndWaitDeployment(kubecli kubernetes.Interface, ns string, deployment *appsv1.Deployment, timeout time.Duration) (*appsv1.Deployment, error) {
	deployment, err := kubecli.AppsV1().Deployments(ns).Create(deployment)
	if err != nil {
		return nil, err
	}

	interval := 5 * time.Second
	var retDeployment *appsv1.Deployment
	err = retryutil.Retry(interval, int(timeout/(interval)), func() (bool, error) {
		retDeployment, err = kubecli.AppsV1().Deployments(ns).Get(deployment.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		switch {
		case retDeployment.Status.ReadyReplicas == *deployment.Spec.Replicas:
			return true, nil
		case retDeployment.Status.ReadyReplicas < *deployment.Spec.Replicas:
			return false, nil
		default:
			return false, fmt.Errorf("unexpected retDeployment.Status.ReadyReplicas: %v", retDeployment.Status.ReadyReplicas)
		}
	})

	if err != nil {
		if retryutil.IsRetryFailure(err) {
			return nil, fmt.Errorf("failed to wait deployment running, it is still pending: %v", err)
		}
		return nil, fmt.Errorf("failed to wait deployment running: %v", err)
	}

	return retDeployment, nil
}

func newSensuServiceManifest(svcName, clusterName, clusterIP string, ports []v1.ServicePort) *v1.Service {
	var extraLabels = []label{{
		key:   "service",
		value: svcName,
	}}

	labels := LabelsForCluster(clusterName, extraLabels...)

	// Create a copy of the labels map to use as the selector, removing the 'service' key
	selectorLabels := map[string]string{}
	for k, v := range labels {
		selectorLabels[k] = v
	}
	delete(selectorLabels, "service")

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:   svcName,
			Labels: labels,
			Annotations: map[string]string{
				TolerateUnreadyEndpointsAnnotation: "true",
			},
		},
		Spec: v1.ServiceSpec{
			Ports:     ports,
			Selector:  selectorLabels,
			ClusterIP: clusterIP,
		},
	}
	return svc
}

// AddEtcdVolumeToPod abstract the process of appending volume spec to pod spec
func AddEtcdVolumeToPod(pod *v1.PodTemplateSpec, pvc *v1.PersistentVolumeClaim) {
	vol := v1.Volume{Name: etcdVolumeName}
	if pvc != nil {
		vol.VolumeSource = v1.VolumeSource{
			PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{ClaimName: pvc.Name},
		}
	} else {
		vol.VolumeSource = v1.VolumeSource{EmptyDir: &v1.EmptyDirVolumeSource{}}
	}
	pod.Spec.Volumes = append(pod.Spec.Volumes, vol)
}

func addOwnerRefToObject(o metav1.Object, r metav1.OwnerReference) {
	o.SetOwnerReferences(append(o.GetOwnerReferences(), r))
}

// CreateNetPolicy creates a NetworkPolicy for a Sensu cluster
func CreateNetPolicy(kubecli kubernetes.Interface, clusterName, namespace string, owner metav1.OwnerReference) error {
	labels := map[string]string{
		"app":           "sensu",
		"sensu_cluster": clusterName,
	}

	netCases := []networkingv1.NetworkPolicy{
		{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "sensu-block-all-",
				Labels:       labels,
				Namespace:    metav1.NamespaceDefault,
			},
			Spec: networkingv1.NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{
					MatchLabels: labels,
				},
				Ingress: []networkingv1.NetworkPolicyIngressRule{},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "sensu-api-pods-",
				Labels:       labels,
				Namespace:    metav1.NamespaceDefault,
			},
			Spec: networkingv1.NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{
					MatchLabels: labels,
				},
				Ingress: []networkingv1.NetworkPolicyIngressRule{
					{
						Ports: []networkingv1.NetworkPolicyPort{
							{
								Port: &intstr.IntOrString{Type: intstr.Int, IntVal: 3000},
							},
							{
								Port: &intstr.IntOrString{Type: intstr.Int, IntVal: 8080},
							},
							{
								Port: &intstr.IntOrString{Type: intstr.Int, IntVal: 8081},
							},
						},
						From: []networkingv1.NetworkPolicyPeer{},
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "sensu-operator-pods-",
				Labels:       labels,
				Namespace:    metav1.NamespaceDefault,
			},
			Spec: networkingv1.NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{
					MatchLabels: labels,
				},
				Ingress: []networkingv1.NetworkPolicyIngressRule{
					{
						Ports: []networkingv1.NetworkPolicyPort{
							{
								Port: &intstr.IntOrString{Type: intstr.Int, IntVal: 2379},
							},
							{
								Port: &intstr.IntOrString{Type: intstr.Int, IntVal: 2380},
							},
						},
						From: []networkingv1.NetworkPolicyPeer{
							{
								PodSelector: &metav1.LabelSelector{
									MatchLabels: map[string]string{
										"name": "sensu-operator",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "sensu-cluster-pods-",
				Labels:       labels,
				Namespace:    metav1.NamespaceDefault,
			},
			Spec: networkingv1.NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{
					MatchLabels: labels,
				},
				Ingress: []networkingv1.NetworkPolicyIngressRule{
					{
						Ports: []networkingv1.NetworkPolicyPort{
							{
								Port: &intstr.IntOrString{Type: intstr.Int, IntVal: 2379},
							},
							{
								Port: &intstr.IntOrString{Type: intstr.Int, IntVal: 2380},
							},
						},
						From: []networkingv1.NetworkPolicyPeer{
							{
								PodSelector: &metav1.LabelSelector{
									MatchLabels: labels,
								},
							},
						},
					},
				},
			},
		},
	}

	for _, net := range netCases {
		addOwnerRefToObject(net.GetObjectMeta(), owner)
		if _, err := kubecli.NetworkingV1().NetworkPolicies(namespace).Create(&net); err != nil {
			return err
		}
	}
	return nil
}

// NewSensuPodPVC create PVC object from etcd pod's PVC spec
func NewSensuPodPVC(m *etcdutil.MemberConfig, pvcSpec v1.PersistentVolumeClaimSpec, clusterName, namespace string, owner metav1.OwnerReference) *v1.PersistentVolumeClaim {
	pvc := &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "etcd-data",
			Namespace: namespace,
			Labels:    LabelsForCluster(clusterName),
		},
		Spec: pvcSpec,
	}
	addOwnerRefToObject(pvc.GetObjectMeta(), owner)
	return pvc
}

func newSensuPodTemplate(m *etcdutil.MemberConfig, clusterName, token string, cs api.ClusterSpec) *v1.PodTemplateSpec {
	commands := "/usr/local/bin/sensu-backend start -c /etc/sensu/backend.yml"
	options := fmt.Sprintf(`log-level: info
state-dir: %s
etcd-name: "$HOSTNAME"
etcd-initial-advertise-peer-urls: "http://${LOCAL_HOSTNAME}:2380"
etcd-listen-peer-urls: "%s"
etcd-listen-client-urls: "%s"
etcd-advertise-client-urls: "http://${LOCAL_HOSTNAME}:2379"
etcd-initial-cluster: "${INITIAL_CLUSTER}"
etcd-initial-cluster-state: "${STATE}"
`, stateDir, m.ListenPeerURL(), m.ListenClientURL())

	if m.SecurePeer {
		options += fmt.Sprintf(`peer-client-cert-auth: true
etcd-peer-trusted-ca-file: %[1]s/peer-ca.crt
etcd-peer-cert-file: %[1]s/peer.crt
etcd-peer-key-file: %[1]s/peer.key
`, peerTLSDir)
	}

	if m.SecureClient {
		options += fmt.Sprintf(`etcd-lient-cert-auth: true
etcd-trusted-ca-file: %[1]s/server-ca.crt
etcd-cert-file=%[1]s/server.crt
etcd-key-file: %[1]s/server.key
`, serverTLSDir)
	}

	labels := map[string]string{
		"app":           "sensu",
		"sensu_cluster": clusterName,
	}

	livenessProbe := newSensuProbe()
	readinessProbe := newSensuProbe()
	readinessProbe.InitialDelaySeconds = 1
	readinessProbe.TimeoutSeconds = 5
	readinessProbe.PeriodSeconds = 5
	readinessProbe.FailureThreshold = 3

	configVolumeMount := v1.VolumeMount{
		Name:      "etcsensu",
		MountPath: "/etc/sensu",
	}
	container := containerWithProbes(
		sensuContainer(strings.Split(commands, " "), cs.Repository, cs.Version, cs.ClusterAdminUsername, cs.ClusterAdminPassword),
		livenessProbe,
		readinessProbe)
	container.VolumeMounts = append(container.VolumeMounts, configVolumeMount)

	volumes := []v1.Volume{}

	if m.SecurePeer {
		container.VolumeMounts = append(container.VolumeMounts, v1.VolumeMount{
			MountPath: peerTLSDir,
			Name:      peerTLSVolume,
		})
		volumes = append(volumes, v1.Volume{Name: peerTLSVolume, VolumeSource: v1.VolumeSource{
			Secret: &v1.SecretVolumeSource{SecretName: cs.TLS.Static.Member.PeerSecret},
		}})
	}
	if m.SecureClient {
		container.VolumeMounts = append(container.VolumeMounts, v1.VolumeMount{
			MountPath: serverTLSDir,
			Name:      serverTLSVolume,
		}, v1.VolumeMount{
			MountPath: operatorEtcdTLSDir,
			Name:      operatorEtcdTLSVolume,
		})
		volumes = append(volumes, v1.Volume{Name: serverTLSVolume, VolumeSource: v1.VolumeSource{
			Secret: &v1.SecretVolumeSource{SecretName: cs.TLS.Static.Member.ServerSecret},
		}}, v1.Volume{Name: operatorEtcdTLSVolume, VolumeSource: v1.VolumeSource{
			Secret: &v1.SecretVolumeSource{SecretName: cs.TLS.Static.OperatorSecret},
		}})
	}
	volumes = append(volumes, v1.Volume{
		Name: "etcsensu",
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{},
		},
	})

	DNSTimeout := defaultDNSTimeout
	if cs.Pod != nil {
		DNSTimeout = cs.Pod.DNSTimeoutInSecond
	}
	podTemplateSpec := v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:        clusterName,
			Labels:      labels,
			Annotations: map[string]string{},
		},
		Spec: v1.PodSpec{
			InitContainers: []v1.Container{
				{
					// busybox:latest uses uclibc which contains a bug that sometimes prevents name resolution
					// More info: https://github.com/docker-library/busybox/issues/27
					//Image default: "busybox:1.28.0-glibc",
					Image: imageNameBusybox(cs.Pod),
					Name:  "check-dns",
					// In etcd 3.2, TLS listener will do a reverse-DNS lookup for pod IP -> hostname.
					// If DNS entry is not warmed up, it will return empty result and peer connection will be rejected.
					// In some cases the DNS is not created correctly so we need to time out after a given period.
					Command: []string{"/bin/sh", "-c", fmt.Sprintf(`
					TIMEOUT_READY=%d
					SUBDOMAIN=%s
					NAMESPACE=%s
					LOCAL_HOSTNAME=$(hostname).${SUBDOMAIN}.${NAMESPACE}.svc
					while ( ! nslookup $LOCAL_HOSTNAME )
					do
						# If TIMEOUT_READY is 0 we should never time out and exit
						TIMEOUT_READY=$(( TIMEOUT_READY-1 ))
                        if [ $TIMEOUT_READY -eq 0 ];
				        then
				            echo "Timed out waiting for DNS entry"
				            exit 1
				        fi
						sleep 1
					done`, DNSTimeout, clusterName, m.Namespace)},
					SecurityContext: &v1.SecurityContext{
						RunAsUser:                ptrInt64(65534),
						RunAsGroup:               ptrInt64(65534),
						AllowPrivilegeEscalation: ptrBool(false),
					},
				},
				{
					Image: imageNameBusybox(cs.Pod),
					Name:  "make-sensu-config",
					SecurityContext: &v1.SecurityContext{
						RunAsUser:                ptrInt64(65534),
						RunAsGroup:               ptrInt64(65534),
						AllowPrivilegeEscalation: ptrBool(false),
					},
					Command: []string{"/bin/sh", "-c", fmt.Sprintf(`HOSTNAME=$(hostname)
ORDINAL=${HOSTNAME##*-}
TOKEN=%s
SUBDOMAIN=%s
NAMESPACE=%s
LOCAL_HOSTNAME=${HOSTNAME}.${SUBDOMAIN}.${NAMESPACE}.svc
SEED_NAME=${SUBDOMAIN}-0
SEED_HOSTNAME=${SEED_NAME}.${SUBDOMAIN}.${NAMESPACE}.svc
INITIAL_CLUSTER="${SEED_NAME}=http://${SEED_HOSTNAME}:2380"
STATE="new"
if [[ "$ORDINAL" == "0" ]]
then
	STATE="new"
else
	STATE="existing"
	for i in $(seq 1 $ORDINAL)
	do
		INITIAL_CLUSTER=${INITIAL_CLUSTER},${SUBDOMAIN}-${i}=http://${SUBDOMAIN}-${i}.${SUBDOMAIN}.${NAMESPACE}.svc:2380
	done
fi
if [[ "${STATE}" == "new" ]]
then
    echo "etcd-initial-cluster-token: ${TOKEN}" >> /etc/sensu/backend.yml
fi
cat >> /etc/sensu/backend.yml <<EOL
%s
EOL
cat /etc/sensu/backend.yml
`, token, clusterName, m.Namespace, options)},
					VolumeMounts: []v1.VolumeMount{configVolumeMount},
				},
			},
			Containers:    []v1.Container{container},
			RestartPolicy: v1.RestartPolicyAlways,
			Volumes:       volumes,
			// This is a hack, this likely should be exposed in the sensucluster rbac specification
			ServiceAccountName: "sensu-operator",
			// DNS A record: `[m.Name].[clusterName].Namespace.svc`
			// For example, etcd-795649v9kq in default namesapce will have DNS name
			// `etcd-795649v9kq.etcd.default.svc`.
			Subdomain:                    clusterName,
			AutomountServiceAccountToken: func(b bool) *bool { return &b }(false),
			SecurityContext:              podSecurityContext(cs.Pod),
		},
	}

	SetPodTemplateSensuVersion(&podTemplateSpec, cs.Version)
	return &podTemplateSpec
}

func podSecurityContext(podPolicy *api.PodPolicy) *v1.PodSecurityContext {
	if podPolicy == nil {
		return nil
	}
	return podPolicy.SecurityContext
}

// NewSensuStatefulSet creates a new StatefulSet for a Sensu cluster
func NewSensuStatefulSet(m *etcdutil.MemberConfig, clusterName, token string, cs api.ClusterSpec, owner metav1.OwnerReference) *appsv1.StatefulSet {
	podTemplate := newSensuPodTemplate(m, clusterName, token, cs)
	applyPodPolicy(clusterName, podTemplate, cs.Pod)
	addOwnerRefToObject(podTemplate.GetObjectMeta(), owner)
	statefulSet := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterName,
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: podTemplate.ObjectMeta.Labels,
			},
			Template:            *podTemplate,
			ServiceName:         clusterName,
			Replicas:            newInt32(1),
			PodManagementPolicy: appsv1.ParallelPodManagement,
		},
	}
	return statefulSet
}

// MustNewKubeClient creates a new Kubernetes client with an in cluster config or panics
func MustNewKubeClient() kubernetes.Interface {
	cfg, err := InClusterConfig()
	if err != nil {
		panic(err)
	}
	return kubernetes.NewForConfigOrDie(cfg)
}

func InClusterConfig() (*rest.Config, error) {
	// Work around https://github.com/kubernetes/kubernetes/issues/40973
	// See https://github.com/sensu/sensu-operator/issues/731#issuecomment-283804819
	if len(os.Getenv("KUBERNETES_SERVICE_HOST")) == 0 {
		addrs, err := net.LookupHost("kubernetes.default.svc")
		if err != nil {
			panic(err)
		}
		os.Setenv("KUBERNETES_SERVICE_HOST", addrs[0])
	}
	if len(os.Getenv("KUBERNETES_SERVICE_PORT")) == 0 {
		os.Setenv("KUBERNETES_SERVICE_PORT", "443")
	}
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func IsKubernetesResourceAlreadyExistError(err error) bool {
	return apierrors.IsAlreadyExists(err)
}

func IsKubernetesResourceNotFoundError(err error) bool {
	return apierrors.IsNotFound(err)
}

// ClusterListOpt returns the ListOptions for selecting a Cluster
func ClusterListOpt(clusterName string) metav1.ListOptions {
	return metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(LabelsForCluster(clusterName)).String(),
	}
}

func LabelsForCluster(clusterName string, extraLabels ...label) map[string]string {
	labels := map[string]string{
		"sensu_cluster": clusterName,
		"app":           "sensu",
	}

	for _, v := range extraLabels {
		labels[v.key] = v.value
	}

	return labels
}

func CreatePatch(o, n, datastruct interface{}) ([]byte, error) {
	oldData, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}
	newData, err := json.Marshal(n)
	if err != nil {
		return nil, err
	}
	return strategicpatch.CreateTwoWayMergePatch(oldData, newData, datastruct)
}

// mergeLabels merges l2 into l1. Conflicting label will be skipped.
func mergeLabels(l1, l2 map[string]string) {
	for k, v := range l2 {
		if _, ok := l1[k]; ok {
			continue
		}
		l1[k] = v
	}
}

func UniqueMemberName(clusterName string) string {
	suffix := utilrand.String(randomSuffixLength)
	if len(clusterName) > maxNameLength {
		clusterName = clusterName[:maxNameLength]
	}
	return clusterName + "-" + suffix
}

func newInt32(i int) *int32 {
	var newI int32 = int32(i)
	return &newI
}
