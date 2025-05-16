package static_resource_controller

import (
	"context"
	"fmt"
	"reflect"

	"k8s.io/client-go/util/retry"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"

	"github.com/openshift/zero-trust-workload-identity-manager/api/v1alpha1"
	"github.com/openshift/zero-trust-workload-identity-manager/pkg/controller/utils"
)

type staticResourceCtrlClientImpl struct {
	client.Client
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate -o fakes . CtrlClient
type StaticResourceCtrlClient interface {
	Get(context.Context, client.ObjectKey, client.Object) error
	List(context.Context, client.ObjectList, ...client.ListOption) error
	StatusUpdate(context.Context, client.Object, ...client.SubResourceUpdateOption) error
	Update(context.Context, client.Object, ...client.UpdateOption) error
	UpdateWithRetry(context.Context, client.Object, ...client.UpdateOption) error
	Create(context.Context, client.Object, ...client.CreateOption) error
	Delete(context.Context, client.Object, ...client.DeleteOption) error
	Patch(context.Context, client.Object, client.Patch, ...client.PatchOption) error
	Exists(context.Context, client.ObjectKey, client.Object) (bool, error)
	CreateOrUpdateObject(ctx context.Context, obj client.Object) error
}

func NewStaticResourceControllerClient(m manager.Manager) (StaticResourceCtrlClient, error) {
	c, err := BuildCustomStaticResourceClient(m)
	if err != nil {
		return nil, fmt.Errorf("failed to build custom client: %w", err)
	}
	return &staticResourceCtrlClientImpl{
		Client: c,
	}, nil
}

func (c *staticResourceCtrlClientImpl) Get(
	ctx context.Context, key client.ObjectKey, obj client.Object,
) error {
	return c.Client.Get(ctx, key, obj)
}

func (c *staticResourceCtrlClientImpl) List(
	ctx context.Context, list client.ObjectList, opts ...client.ListOption,
) error {
	return c.Client.List(ctx, list, opts...)
}

func (c *staticResourceCtrlClientImpl) Create(
	ctx context.Context, obj client.Object, opts ...client.CreateOption,
) error {
	return c.Client.Create(ctx, obj, opts...)
}

func (c *staticResourceCtrlClientImpl) Delete(
	ctx context.Context, obj client.Object, opts ...client.DeleteOption,
) error {
	return c.Client.Delete(ctx, obj, opts...)
}

func (c *staticResourceCtrlClientImpl) Update(
	ctx context.Context, obj client.Object, opts ...client.UpdateOption,
) error {
	return c.Client.Update(ctx, obj, opts...)
}

func (c *staticResourceCtrlClientImpl) UpdateWithRetry(
	ctx context.Context, obj client.Object, opts ...client.UpdateOption,
) error {
	key := types.NamespacedName{Name: obj.GetName(), Namespace: obj.GetNamespace()}
	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		current := reflect.New(reflect.TypeOf(obj).Elem()).Interface().(client.Object)
		if err := c.Client.Get(ctx, key, current); err != nil {
			return fmt.Errorf("failed to fetch latest %q for update: %w", key, err)
		}
		obj.SetResourceVersion(current.GetResourceVersion())
		if err := c.Client.Update(ctx, obj, opts...); err != nil {
			return fmt.Errorf("failed to update %q resource: %w", key, err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (c *staticResourceCtrlClientImpl) StatusUpdate(
	ctx context.Context, obj client.Object, opts ...client.SubResourceUpdateOption,
) error {
	return c.Client.Status().Update(ctx, obj, opts...)
}

func (c *staticResourceCtrlClientImpl) Patch(
	ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption,
) error {
	return c.Client.Patch(ctx, obj, patch, opts...)
}

func (c *staticResourceCtrlClientImpl) Exists(ctx context.Context, key client.ObjectKey, obj client.Object) (bool, error) {
	if err := c.Client.Get(ctx, key, obj); err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// CreateOrUpdateObject tries to create the object, updates if already exists
func (c *staticResourceCtrlClientImpl) CreateOrUpdateObject(ctx context.Context, obj client.Object) error {
	err := c.Create(ctx, obj)
	if err != nil && errors.IsAlreadyExists(err) {
		return c.Update(ctx, obj)
	}
	return err
}

func BuildCustomStaticResourceClient(mgr ctrl.Manager) (client.Client, error) {
	spireServerManagedResourceAppManagedReq, err := labels.NewRequirement(utils.SpireSeverAppManagedByLabelKey, selection.Equals, []string{utils.SpireSeverAppManagedByLabelValue})
	if err != nil {
		return nil, err
	}
	managedResourceLabelReqSelector := labels.NewSelector().Add(*spireServerManagedResourceAppManagedReq)

	customCacheOpts := cache.Options{
		HTTPClient: mgr.GetHTTPClient(),
		Scheme:     mgr.GetScheme(),
		Mapper:     mgr.GetRESTMapper(),
		ByObject: map[client.Object]cache.ByObject{
			&rbacv1.ClusterRole{}: {
				Label: managedResourceLabelReqSelector,
			},
			&rbacv1.ClusterRoleBinding{}: {
				Label: managedResourceLabelReqSelector,
			},
			&rbacv1.Role{}: {
				Label: managedResourceLabelReqSelector,
			},
			&rbacv1.RoleBinding{}: {
				Label: managedResourceLabelReqSelector,
			},
			&corev1.Service{}: {
				Label: managedResourceLabelReqSelector,
			},
			&corev1.ServiceAccount{}: {
				Label: managedResourceLabelReqSelector,
			},
			&storagev1.CSIDriver{}: {
				Label: managedResourceLabelReqSelector,
			},
			&admissionregistrationv1.ValidatingWebhookConfiguration{}: {
				Label: managedResourceLabelReqSelector,
			},
		},
		ReaderFailOnMissingInformer: true,
	}
	customCache, err := cache.New(mgr.GetConfig(), customCacheOpts)
	if err != nil {
		return nil, err
	}
	if _, err = customCache.GetInformer(context.Background(), &v1alpha1.ZeroTrustWorkloadIdentityManager{}); err != nil {
		return nil, err
	}
	if _, err = customCache.GetInformer(context.Background(), &corev1.ConfigMap{}); err != nil {
		return nil, err
	}
	if _, err = customCache.GetInformer(context.Background(), &rbacv1.ClusterRole{}); err != nil {
		return nil, err
	}
	if _, err = customCache.GetInformer(context.Background(), &rbacv1.ClusterRoleBinding{}); err != nil {
		return nil, err
	}
	if _, err = customCache.GetInformer(context.Background(), &rbacv1.Role{}); err != nil {
		return nil, err
	}
	if _, err = customCache.GetInformer(context.Background(), &rbacv1.RoleBinding{}); err != nil {
		return nil, err
	}
	if _, err = customCache.GetInformer(context.Background(), &corev1.Service{}); err != nil {
		return nil, err
	}
	if _, err = customCache.GetInformer(context.Background(), &corev1.ServiceAccount{}); err != nil {
		return nil, err
	}
	if _, err = customCache.GetInformer(context.Background(), &storagev1.CSIDriver{}); err != nil {
		return nil, err
	}
	if _, err = customCache.GetInformer(context.Background(), &admissionregistrationv1.ValidatingWebhookConfiguration{}); err != nil {
		return nil, err
	}

	err = mgr.Add(customCache)
	if err != nil {
		return nil, err
	}

	customClient, err := client.New(mgr.GetConfig(), client.Options{
		HTTPClient: mgr.GetHTTPClient(),
		Scheme:     mgr.GetScheme(),
		Mapper:     mgr.GetRESTMapper(),
		Cache: &client.CacheOptions{
			Reader: customCache,
		},
	})
	if err != nil {
		return nil, err
	}
	return customClient, nil
}
