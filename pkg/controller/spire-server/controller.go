package spire_server

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
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
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// SpireServerReconciler reconciles a SpireServer object
type SpireServerReconciler struct {
	ctrlClient    customClient.CustomCtrlClient
	ctx           context.Context
	eventRecorder record.EventRecorder
	log           logr.Logger
	scheme        *runtime.Scheme
}

// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete

// New returns a new Reconciler instance.
func New(mgr ctrl.Manager) (*SpireServerReconciler, error) {
	c, err := customClient.NewCustomClient(mgr)
	if err != nil {
		return nil, err
	}
	return &SpireServerReconciler{
		ctrlClient:    c,
		ctx:           context.Background(),
		eventRecorder: mgr.GetEventRecorderFor(utils.ZeroTrustWorkloadIdentityManagerSpireServerControllerName),
		log:           ctrl.Log.WithName(utils.ZeroTrustWorkloadIdentityManagerSpireServerControllerName),
		scheme:        mgr.GetScheme(),
	}, nil
}

func (r *SpireServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	var server v1alpha1.SpireServerConfig
	if err := r.ctrlClient.Get(ctx, req.NamespacedName, &server); err != nil {
		if kerrors.IsNotFound(err) {
			r.log.Info("SpireServerConfig resource not found. Ignoring since object must be deleted or not been created.")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	spireServerConfigMap, err := GenerateSpireServerConfigMap(&server.Spec)
	if err != nil {
		return ctrl.Result{}, err
	}
	// Set owner reference so GC cleans up when CR is deleted
	if err := controllerutil.SetControllerReference(&server, spireServerConfigMap, r.scheme); err != nil {
		return ctrl.Result{}, err
	}

	var existingSpireServerCM corev1.ConfigMap
	err = r.ctrlClient.Get(ctx, types.NamespacedName{Name: spireServerConfigMap.Name, Namespace: spireServerConfigMap.Namespace}, &existingSpireServerCM)
	if err != nil && kerrors.IsNotFound(err) {
		if err := r.ctrlClient.Create(ctx, spireServerConfigMap); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to create ConfigMap: %w", err)
		}
		r.log.Info("Created spire sever ConfigMap")
	} else if err == nil && existingSpireServerCM.Data["server.conf"] != spireServerConfigMap.Data["server.conf"] {
		existingSpireServerCM.Data = spireServerConfigMap.Data
		if err := r.ctrlClient.Update(ctx, &existingSpireServerCM); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to update ConfigMap: %w", err)
		}
		r.log.Info("Updated ConfigMap with new config")
	} else if err != nil {
		return ctrl.Result{}, err
	}

	spireServerConfJSON, err := marshalToJSON(generateServerConfMap(&server.Spec))
	if err != nil {
		return ctrl.Result{}, err
	}

	spireServerConfigMapHash := generateConfigHash(spireServerConfJSON)

	spireControllerManagerConfig := generateSpireControllerManagerConfigYaml(&server.Spec)
	spireControllerManagerConfigMap := generateControllerManagerConfigMap(spireControllerManagerConfig)
	// Set owner reference so GC cleans up when CR is deleted
	if err := controllerutil.SetControllerReference(&server, spireControllerManagerConfigMap, r.scheme); err != nil {
		return ctrl.Result{}, err
	}

	var existingSpireControllerManagerCM corev1.ConfigMap

	if err := controllerutil.SetControllerReference(&server, spireControllerManagerConfigMap, r.scheme); err != nil {
		return ctrl.Result{}, err
	}
	err = r.ctrlClient.Get(ctx, types.NamespacedName{Name: spireControllerManagerConfigMap.Name, Namespace: spireControllerManagerConfigMap.Namespace}, &existingSpireControllerManagerCM)
	if err != nil && kerrors.IsNotFound(err) {
		if err := r.ctrlClient.Create(ctx, spireControllerManagerConfigMap); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to create ConfigMap: %w", err)
		}
		r.log.Info("Created spire controller manager ConfigMap")
	} else if err == nil && existingSpireControllerManagerCM.Data["controller-manager-config.yaml"] != existingSpireControllerManagerCM.Data["controller-manager-config.yaml"] {
		existingSpireControllerManagerCM.Data = spireControllerManagerConfigMap.Data
		if err := r.ctrlClient.Update(ctx, &existingSpireControllerManagerCM); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to update ConfigMap: %w", err)
		}
		r.log.Info("Updated ConfigMap with new config")
	} else if err != nil {
		return ctrl.Result{}, err
	}

	spireControllerManagerConfigMapHash := generateConfigHashFromString(spireControllerManagerConfig)

	spireBundleCM := generateSpireBundleConfigMap()
	if err := controllerutil.SetControllerReference(&server, spireControllerManagerConfigMap, r.scheme); err != nil {
		return ctrl.Result{}, err
	}
	err = r.ctrlClient.Create(ctx, spireBundleCM)
	if err != nil && !kerrors.IsAlreadyExists(err) {
		return ctrl.Result{}, fmt.Errorf("failed to create spire-bundle ConfigMap: %w", err)
	}

	sts := GenerateSpireServerStatefulSet(&server.Spec, spireServerConfigMapHash, spireControllerManagerConfigMapHash)
	if err := controllerutil.SetControllerReference(&server, sts, r.scheme); err != nil {
		return ctrl.Result{}, err
	}

	// 5. Create or Update StatefulSet
	var existingSTS appsv1.StatefulSet
	err = r.ctrlClient.Get(ctx, types.NamespacedName{Name: sts.Name, Namespace: sts.Namespace}, &existingSTS)
	if err != nil && kerrors.IsNotFound(err) {
		if err := r.ctrlClient.Create(ctx, sts); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to create StatefulSet: %w", err)
		}
		r.log.Info("Created spire sever StatefulSet")
	} else if err == nil && needsUpdate(existingSTS, *sts) {
		existingSTS.Spec = sts.Spec
		if err := r.ctrlClient.Update(ctx, &existingSTS); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to update StatefulSet: %w", err)
		}
		r.log.Info("Updated spire sever StatefulSet")
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

func (r *SpireServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
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
		For(&v1alpha1.SpireServerConfig{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Named(utils.ZeroTrustWorkloadIdentityManagerSpireServerControllerName).
		Watches(&appsv1.StatefulSet{}, handler.EnqueueRequestsFromMapFunc(mapFunc), controllerManagedResourcePredicates).
		Watches(&corev1.ConfigMap{}, handler.EnqueueRequestsFromMapFunc(mapFunc), controllerManagedResourcePredicates).
		Complete(r)
	if err != nil {
		return err
	}
	return nil
}

// needsUpdate returns true if StatefulSet needs to be updated based on config checksum
func needsUpdate(current, desired appsv1.StatefulSet) bool {
	if current.Spec.Template.Annotations[spireServerStatefulSetSpireServerConfigHashAnnotationKey] != desired.Spec.Template.Annotations[spireServerStatefulSetSpireServerConfigHashAnnotationKey] {
		return true
	} else if current.Spec.Template.Annotations[spireServerStatefulSetSpireControllerMangerConfigHashAnnotationKey] != desired.Spec.Template.Annotations[spireServerStatefulSetSpireControllerMangerConfigHashAnnotationKey] {
		return true
	}
	return false
}
