package spire_oidc_discovery_provider

import (
	"context"
	"github.com/go-logr/logr"
	securityv1 "github.com/openshift/api/security/v1"
	customClient "github.com/openshift/zero-trust-workload-identity-manager/pkg/client"
	"github.com/openshift/zero-trust-workload-identity-manager/pkg/controller/utils"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/openshift/zero-trust-workload-identity-manager/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

// SpireOidcDiscoveryProviderReconciler reconciles a SpireOidcDiscoveryProvider object
type SpireOidcDiscoveryProviderReconciler struct {
	ctrlClient    customClient.CustomCtrlClient
	ctx           context.Context
	eventRecorder record.EventRecorder
	log           logr.Logger
	scheme        *runtime.Scheme
}

// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete

// New returns a new Reconciler instance.
func New(mgr ctrl.Manager) (*SpireOidcDiscoveryProviderReconciler, error) {
	c, err := customClient.NewCustomClient(mgr)
	if err != nil {
		return nil, err
	}
	return &SpireOidcDiscoveryProviderReconciler{
		ctrlClient:    c,
		ctx:           context.Background(),
		eventRecorder: mgr.GetEventRecorderFor(utils.ZeroTrustWorkloadIdentityManagerSpireOIDCDiscoveryProviderControllerName),
		log:           ctrl.Log.WithName(utils.ZeroTrustWorkloadIdentityManagerSpireOIDCDiscoveryProviderControllerName),
		scheme:        mgr.GetScheme(),
	}, nil
}

func (r *SpireOidcDiscoveryProviderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.log.Info("Reconciling SpireOIDCDiscoveryProviderConfig controller")

	var oidcDiscoveryProviderConfig v1alpha1.SpireOIDCDiscoveryProviderConfig
	if err := r.ctrlClient.Get(ctx, req.NamespacedName, &oidcDiscoveryProviderConfig); err != nil {
		if kerrors.IsNotFound(err) {
			r.log.Info("SpireOidcDiscoveryProviderConfig resource not found. Ignoring since object must be deleted or not been created.")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	spireOIDCClusterSpiffeID := generateSpireIODCDiscoveryProviderSpiffeID()
	err := r.ctrlClient.Create(ctx, spireOIDCClusterSpiffeID)
	if err != nil && !kerrors.IsAlreadyExists(err) {
		r.log.Error(err, "Failed to create SpireOidcDiscoveryProviderConfig")
		return ctrl.Result{}, err
	}
	defaultSpiffeID := generateDefaultFallbackClusterSPIFFEID()
	err = r.ctrlClient.Create(ctx, defaultSpiffeID)
	if err != nil && !kerrors.IsAlreadyExists(err) {
		r.log.Error(err, "Failed to create DefaultFallbackClusterSPIFFEID")
		return ctrl.Result{}, err
	}
	cm, err := GenerateOIDCConfigMapFromCR(&oidcDiscoveryProviderConfig)
	if err != nil {
		return ctrl.Result{}, err
	}
	err = r.ctrlClient.Create(ctx, cm)
	if err != nil && !kerrors.IsAlreadyExists(err) {
		return ctrl.Result{}, err
	}
	scc := generateSpireOIDCDiscoveryProviderSCC(&oidcDiscoveryProviderConfig)
	err = r.ctrlClient.Create(ctx, scc)
	if err != nil && !kerrors.IsAlreadyExists(err) {
		return ctrl.Result{}, err
	}
	deployment := buildDeployment(&oidcDiscoveryProviderConfig)
	err = r.ctrlClient.Create(ctx, deployment)
	if err != nil && !kerrors.IsAlreadyExists(err) {
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

func (r *SpireOidcDiscoveryProviderReconciler) SetupWithManager(mgr ctrl.Manager) error {
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
		For(&v1alpha1.SpireOIDCDiscoveryProviderConfig{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Named(utils.ZeroTrustWorkloadIdentityManagerSpireOIDCDiscoveryProviderControllerName).
		Watches(&appsv1.Deployment{}, handler.EnqueueRequestsFromMapFunc(mapFunc), controllerManagedResourcePredicates).
		Watches(&corev1.ConfigMap{}, handler.EnqueueRequestsFromMapFunc(mapFunc), controllerManagedResourcePredicates).
		Watches(&securityv1.SecurityContextConstraints{}, handler.EnqueueRequestsFromMapFunc(mapFunc), controllerManagedResourcePredicates).
		Complete(r)
	if err != nil {
		return err
	}
	return nil
}

//// needsUpdate returns true if StatefulSet needs to be updated based on config checksum
//func needsUpdate(current, desired appsv1.StatefulSet) bool {
//	if current.Spec.Template.Annotations[spireOidcDiscoveryProviderStatefulSetSpireOidcDiscoveryProviderConfigHashAnnotationKey] != desired.Spec.Template.Annotations[spireOidcDiscoveryProviderStatefulSetSpireOidcDiscoveryProviderConfigHashAnnotationKey] {
//		return true
//	} else if current.Spec.Template.Annotations[spireOidcDiscoveryProviderStatefulSetSpireControllerMangerConfigHashAnnotationKey] != desired.Spec.Template.Annotations[spireOidcDiscoveryProviderStatefulSetSpireControllerMangerConfigHashAnnotationKey] {
//		return true
//	}
//	return false
//}
