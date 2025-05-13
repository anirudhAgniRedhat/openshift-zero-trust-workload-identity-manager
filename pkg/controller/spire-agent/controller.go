package spire_agent

import (
	"context"
	"fmt"
	securityv1 "github.com/openshift/api/security/v1"
	customClient "github.com/openshift/zero-trust-workload-identity-manager/pkg/client"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/client-go/tools/record"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/go-logr/logr"

	"github.com/openshift/zero-trust-workload-identity-manager/api/v1alpha1"
	"github.com/openshift/zero-trust-workload-identity-manager/pkg/controller/utils"
)

const spireAgentDaemonSetSpireAgentConfigHashAnnotationKey = "ztwim.openshift.io/spire-agent-config-hash"

// SpireAgentReconciler reconciles a SpireAgent object
type SpireAgentReconciler struct {
	ctrlClient    customClient.CustomCtrlClient
	ctx           context.Context
	eventRecorder record.EventRecorder
	log           logr.Logger
	scheme        *runtime.Scheme
}

// +kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=security.openshift.io,resources=securitycontextconstraints,verbs=get;list;watch;create;update;patch;delete

// New returns a new Reconciler instance.
func New(mgr ctrl.Manager) (*SpireAgentReconciler, error) {
	c, err := customClient.NewCustomClient(mgr)
	if err != nil {
		return nil, err
	}
	return &SpireAgentReconciler{
		ctrlClient:    c,
		ctx:           context.Background(),
		eventRecorder: mgr.GetEventRecorderFor(utils.ZeroTrustWorkloadIdentityManagerSpireAgentControllerName),
		log:           ctrl.Log.WithName(utils.ZeroTrustWorkloadIdentityManagerSpireAgentControllerName),
		scheme:        mgr.GetScheme(),
	}, nil
}

func (r *SpireAgentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var agent v1alpha1.SpireAgentConfig
	if err := r.ctrlClient.Get(ctx, req.NamespacedName, &agent); err != nil {
		if kerrors.IsNotFound(err) {
			r.log.Info("SpireAgentConfig resource not found. Ignoring since object must be deleted or not been created.")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	spireAgentSCC := generateSpireAgentSCC(&agent)
	if err := controllerutil.SetControllerReference(&agent, spireAgentSCC, r.scheme); err != nil {
		return ctrl.Result{}, err
	}
	err := r.ctrlClient.Create(ctx, spireAgentSCC)
	if err != nil && !kerrors.IsAlreadyExists(err) {
		r.log.Error(err, "Failed to create SpireAgentSCC")
		return ctrl.Result{}, err
	}
	spireAgentConfigMap, spireAgentConfigHash, err := GenerateSpireAgentConfigMap(&agent)
	if err != nil {
		return ctrl.Result{}, err
	}
	// Set owner reference so GC cleans up when CR is deleted
	if err := controllerutil.SetControllerReference(&agent, spireAgentConfigMap, r.scheme); err != nil {
		return ctrl.Result{}, err
	}

	var existingSpireAgentCM corev1.ConfigMap
	err = r.ctrlClient.Get(ctx, types.NamespacedName{Name: spireAgentConfigMap.Name, Namespace: spireAgentConfigMap.Namespace}, &existingSpireAgentCM)
	if err != nil && kerrors.IsNotFound(err) {
		if err := r.ctrlClient.Create(ctx, spireAgentConfigMap); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to create ConfigMap: %w", err)
		}
		r.log.Info("Created spire sever ConfigMap")
	} else if err == nil && existingSpireAgentCM.Data["agent.conf"] != spireAgentConfigMap.Data["agent.conf"] {
		existingSpireAgentCM.Data = spireAgentConfigMap.Data
		if err := r.ctrlClient.Update(ctx, &existingSpireAgentCM); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to update ConfigMap: %w", err)
		}
		r.log.Info("Updated ConfigMap with new config")
	} else if err != nil {
		return ctrl.Result{}, err
	}

	spireAgentDaemonset := generateSpireAgentDaemonSet(spireAgentConfigHash)
	if err := controllerutil.SetControllerReference(&agent, spireAgentDaemonset, r.scheme); err != nil {
		return ctrl.Result{}, err
	}

	// 5. Create or Update DaemonSet
	var existingSpireAgentDaemonSet appsv1.DaemonSet
	err = r.ctrlClient.Get(ctx, types.NamespacedName{Name: spireAgentDaemonset.Name, Namespace: spireAgentDaemonset.Namespace}, &existingSpireAgentDaemonSet)
	if err != nil && kerrors.IsNotFound(err) {
		if err := r.ctrlClient.Create(ctx, spireAgentDaemonset); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to create DaemonSet: %w", err)
		}
		r.log.Info("Created spire sever DaemonSet")
	} else if err == nil && needsUpdate(existingSpireAgentDaemonSet, *spireAgentDaemonset) {
		existingSpireAgentDaemonSet.Spec = spireAgentDaemonset.Spec
		if err := r.ctrlClient.Update(ctx, &existingSpireAgentDaemonSet); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to update DaemonSet: %w", err)
		}
		r.log.Info("Updated spire sever DaemonSet")
	} else if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func hasControllerManagedLabel(obj client.Object) bool {
	val, ok := obj.GetLabels()[utils.AppManagedByLabelKey]
	return ok && val == utils.AppManagedByLabelValue
}

// controllerManagedResources filters resources that have a specific label indicating they are managed
var controllerManagedResources = predicate.Funcs{
	UpdateFunc: func(e event.UpdateEvent) bool {
		return hasControllerManagedLabel(e.ObjectNew)
	},
	CreateFunc: func(e event.CreateEvent) bool {
		return hasControllerManagedLabel(e.Object)
	},
	DeleteFunc: func(e event.DeleteEvent) bool {
		return hasControllerManagedLabel(e.Object)
	},
}

func (r *SpireAgentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Always enqueue the "cluster" CR for reconciliation
	mapFunc := func(ctx context.Context, _ client.Object) []reconcile.Request {
		return []reconcile.Request{
			{
				NamespacedName: types.NamespacedName{
					Name: "cluster",
				},
			},
		}
	}

	controllerManagedResourcePredicates := builder.WithPredicates(controllerManagedResources)

	err := ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.SpireAgentConfig{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Named(utils.ZeroTrustWorkloadIdentityManagerSpireAgentControllerName).
		Watches(&appsv1.DaemonSet{}, handler.EnqueueRequestsFromMapFunc(mapFunc), controllerManagedResourcePredicates).
		Watches(&corev1.ConfigMap{}, handler.EnqueueRequestsFromMapFunc(mapFunc), controllerManagedResourcePredicates).
		Watches(&securityv1.SecurityContextConstraints{}, handler.EnqueueRequestsFromMapFunc(mapFunc), controllerManagedResourcePredicates).
		Complete(r)
	if err != nil {
		return err
	}
	return nil
}

// needsUpdate returns true if DaemonSet needs to be updated based on config checksum
func needsUpdate(current, desired appsv1.DaemonSet) bool {
	if current.Spec.Template.Annotations[spireAgentDaemonSetSpireAgentConfigHashAnnotationKey] != desired.Spec.Template.Annotations[spireAgentDaemonSetSpireAgentConfigHashAnnotationKey] {
		return true
	}
	return false
}
