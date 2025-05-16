package static_resource_controller

import (
	"context"
	"fmt"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"github.com/openshift/zero-trust-workload-identity-manager/pkg/controller/utils"
	"github.com/openshift/zero-trust-workload-identity-manager/pkg/operator/assets"
)

func (r *StaticResourceReconciler) ApplyOrCreateValidatingWebhookConfiguration(ctx context.Context) error {
	desired := r.GetSpireControllerManagerValidatingWebhookConfiguration()

	var existing admissionregistrationv1.ValidatingWebhookConfiguration
	err := r.Get(ctx, types.NamespacedName{Name: desired.Name}, &existing)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Not found, so create it
			if err := r.Create(ctx, desired); err != nil {
				return fmt.Errorf("failed to create ValidatingWebhookConfiguration: %w", err)
			}
			return nil
		}
		// Some other error while getting
		return fmt.Errorf("failed to get ValidatingWebhookConfiguration: %w", err)
	}

	// Object exists, set the resourceVersion to allow update
	desired.ResourceVersion = existing.ResourceVersion

	// Apply the update
	if err := r.Update(ctx, desired); err != nil {
		return fmt.Errorf("failed to update ValidatingWebhookConfiguration: %w", err)
	}
	return nil
}

func (r *StaticResourceReconciler) GetSpireControllerManagerValidatingWebhookConfiguration() *admissionregistrationv1.ValidatingWebhookConfiguration {
	spireControllerManagerValidatingWebhookConfiguration := utils.DecodeValidatingWebhookConfigurationByBytes(assets.MustAsset(utils.SpireControllerManagerValidatingWebhookConfigurationAssetName))
	return spireControllerManagerValidatingWebhookConfiguration
}
