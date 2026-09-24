//+kubebuilder:rbac:groups=platform.aegis.dev,resources=tenants,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=platform.aegis.dev,resources=tenants/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=platform.aegis.dev,resources=tenants/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=resourcequotas,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch;create;update;patch;delete
package controller

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	platformv1 "github.com/Demosprojectstests/aegis/operator/api/v1"
)

const finalizerName = "platform.aegis.dev/tenant"

type TenantReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *TenantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	var tenant platformv1.Tenant
	if err := r.Get(ctx, req.NamespacedName, &tenant); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	nsName := "tenant-" + tenant.Name

	if !tenant.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(&tenant, finalizerName) {
			_ = r.Delete(ctx, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: nsName}})
			controllerutil.RemoveFinalizer(&tenant, finalizerName)
			if err := r.Update(ctx, &tenant); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	if !controllerutil.ContainsFinalizer(&tenant, finalizerName) {
		controllerutil.AddFinalizer(&tenant, finalizerName)
		if err := r.Update(ctx, &tenant); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: nsName}}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, ns, func() error {
		if ns.Labels == nil {
			ns.Labels = map[string]string{}
		}
		ns.Labels["aegis.dev/tenant"] = tenant.Name
		return nil
	}); err != nil {
		return r.fail(ctx, &tenant, nsName, err)
	}

	if err := r.Get(ctx, client.ObjectKey{Name: nsName}, ns); err != nil {
		return r.fail(ctx, &tenant, nsName, err)
	}
	owner := metav1.OwnerReference{
		APIVersion: "v1",
		Kind:       "Namespace",
		Name:       ns.Name,
		UID:        ns.UID,
	}

	cpu, mem := "1", "1Gi"
	if tenant.Spec.Plan == "pro" {
		cpu, mem = "4", "4Gi"
	}

	rq := &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{Name: "tenant-quota", Namespace: nsName},
	}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, rq, func() error {
		rq.OwnerReferences = []metav1.OwnerReference{owner}
		rq.Spec.Hard = corev1.ResourceList{
			corev1.ResourceRequestsCPU:    resource.MustParse(cpu),
						   corev1.ResourceRequestsMemory: resource.MustParse(mem),
						   corev1.ResourcePods:           resource.MustParse("20"),
		}
		return nil
	}); err != nil {
		return r.fail(ctx, &tenant, nsName, err)
	}

	udp := corev1.ProtocolUDP
	dnsPort := intstr.FromInt(53)
	np := &netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: "tenant-isolate", Namespace: nsName},
	}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, np, func() error {
		np.OwnerReferences = []metav1.OwnerReference{owner}
		np.Spec.PodSelector = metav1.LabelSelector{}
		np.Spec.PolicyTypes = []netv1.PolicyType{netv1.PolicyTypeIngress, netv1.PolicyTypeEgress}
		np.Spec.Ingress = []netv1.NetworkPolicyIngressRule{{
			From: []netv1.NetworkPolicyPeer{{PodSelector: &metav1.LabelSelector{}}},
		}}
		np.Spec.Egress = []netv1.NetworkPolicyEgressRule{
			{To: []netv1.NetworkPolicyPeer{{PodSelector: &metav1.LabelSelector{}}}},
			{
				To: []netv1.NetworkPolicyPeer{{
					NamespaceSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"kubernetes.io/metadata.name": "kube-system"},
					},
				}},
				Ports: []netv1.NetworkPolicyPort{{Protocol: &udp, Port: &dnsPort}},
			},
		}
		return nil
	}); err != nil {
		return r.fail(ctx, &tenant, nsName, err)
	}

	tenant.Status.Namespace = nsName
	tenant.Status.Ready = true
	tenant.Status.Message = fmt.Sprintf("namespace %s ready", nsName)
	if err := r.Status().Update(ctx, &tenant); err != nil {
		return ctrl.Result{}, err
	}
	log.Info("tenant ready", "namespace", nsName)
	return ctrl.Result{}, nil
}

func (r *TenantReconciler) fail(ctx context.Context, tenant *platformv1.Tenant, ns string, err error) (ctrl.Result, error) {
	tenant.Status.Namespace = ns
	tenant.Status.Ready = false
	tenant.Status.Message = err.Error()
	_ = r.Status().Update(ctx, tenant)
	return ctrl.Result{}, err
}

func (r *TenantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
	For(&platformv1.Tenant{}).
	Named("tenant").
	Complete(r)
}
