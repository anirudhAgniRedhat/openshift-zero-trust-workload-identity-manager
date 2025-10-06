package spire_agent

import (
	"reflect"
	"testing"

	securityv1 "github.com/openshift/api/security/v1"
	"github.com/openshift/zero-trust-workload-identity-manager/api/v1alpha1"
	"github.com/openshift/zero-trust-workload-identity-manager/pkg/controller/utils"
	corev1 "k8s.io/api/core/v1"
)

func TestGenerateSpireAgentSCC(t *testing.T) {
	customLabels := map[string]string{
		"custom-label": "custom-value",
	}
	config := &v1alpha1.SpireAgent{
		Spec: v1alpha1.SpireAgentSpec{
			CommonConfig: v1alpha1.CommonConfig{
				Labels: customLabels,
			},
		},
	}

	scc := generateSpireAgentSCC(config)
	expectedLabels := utils.SpireAgentLabels(customLabels)

	if scc.Name != "spire-agent" {
		t.Errorf("expected SCC name to be 'spire-agent', got %s", scc.Name)
	}

	if !reflect.DeepEqual(scc.Labels, expectedLabels) {
		t.Errorf("expected labels %v, got %v", expectedLabels, scc.Labels)
	}

	if scc.Priority == nil || *scc.Priority != 10 {
		t.Errorf("expected Priority to be 10, got %v", scc.Priority)
	}

	if !scc.ReadOnlyRootFilesystem {
		t.Errorf("expected ReadOnlyRootFilesystem to be true")
	}

	if scc.RunAsUser.Type != securityv1.RunAsUserStrategyRunAsAny {
		t.Errorf("expected RunAsUser.Type to be RunAsAny, got %s", scc.RunAsUser.Type)
	}

	if scc.SELinuxContext.Type != securityv1.SELinuxStrategyMustRunAs {
		t.Errorf("expected SELinuxContext.Type to be MustRunAs, got %s", scc.SELinuxContext.Type)
	}

	if scc.SupplementalGroups.Type != securityv1.SupplementalGroupsStrategyRunAsAny {
		t.Errorf("expected SupplementalGroups.Type to be RunAsAny")
	}

	if scc.FSGroup.Type != securityv1.FSGroupStrategyRunAsAny {
		t.Errorf("expected FSGroup.Type to be RunAsAny")
	}

	expectedUser := "system:serviceaccount:zero-trust-workload-identity-manager:spire-agent"
	if len(scc.Users) != 1 || scc.Users[0] != expectedUser {
		t.Errorf("expected Users to contain %s, got %v", expectedUser, scc.Users)
	}

	expectedVolumes := []securityv1.FSType{
		securityv1.FSTypeConfigMap,
		securityv1.FSTypeHostPath,
		securityv1.FSProjected,
		securityv1.FSTypeSecret,
		securityv1.FSTypeEmptyDir,
	}
	if !reflect.DeepEqual(scc.Volumes, expectedVolumes) {
		t.Errorf("expected Volumes %v, got %v", expectedVolumes, scc.Volumes)
	}

	if !scc.AllowHostDirVolumePlugin {
		t.Errorf("expected AllowHostDirVolumePlugin to be true")
	}
	if scc.AllowHostIPC {
		t.Errorf("expected AllowHostIPC to be false")
	}
	if !scc.AllowHostNetwork {
		t.Errorf("expected AllowHostNetwork to be true")
	}
	if !scc.AllowHostPID {
		t.Errorf("expected AllowHostPID to be true")
	}
	if !scc.AllowHostPorts {
		t.Errorf("expected AllowHostPorts to be true")
	}
	if scc.AllowPrivilegeEscalation == nil || *scc.AllowPrivilegeEscalation {
		t.Errorf("expected AllowPrivilegeEscalation to be false")
	}
	if scc.AllowPrivilegedContainer {
		t.Errorf("expected AllowPrivilegedContainer to be false")
	}

	if len(scc.AllowedCapabilities) != 0 {
		t.Errorf("expected AllowedCapabilities to be empty")
	}
	if len(scc.DefaultAddCapabilities) != 0 {
		t.Errorf("expected DefaultAddCapabilities to be empty")
	}
	expectedDropCapabilities := []corev1.Capability{"ALL"}
	if !reflect.DeepEqual(scc.RequiredDropCapabilities, expectedDropCapabilities) {
		t.Errorf("expected RequiredDropCapabilities to be %v, got %v", expectedDropCapabilities, scc.RequiredDropCapabilities)
	}
	if len(scc.Groups) != 0 {
		t.Errorf("expected Groups to be empty")
	}
}
