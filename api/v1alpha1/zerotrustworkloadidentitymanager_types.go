/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ZeroTrustWorkloadIdentityManagerSpec defines the desired state of ZeroTrustWorkloadIdentityManager
type ZeroTrustWorkloadIdentityManagerSpec struct {

	// workloadIdentityManagerConfig is for configuring the workload identity manager operands behavior.
	// +kubebuilder:validation:Required
	WorkloadIdentityManagerConfig *ZeroTrustWorkloadIdentityManagerConfig `json:"workloadIdentityManagerConfig,omitempty"`

	// controllerConfig is for configuring the controller for setting up
	// defaults to enable spire.
	// +kubebuilder:validation:Optional
	ControllerConfig *OperatorControllerConfig `json:"controllerConfig,omitempty"`
}

// ZeroTrustWorkloadIdentityManagerConfig is for configuring the zero-trust-workload-identity-manager behavior.
type ZeroTrustWorkloadIdentityManagerConfig struct {
	// logLevel supports value range as per [kubernetes logging guidelines](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/logging.md#what-method-to-use).
	// +kubebuilder:default:=1
	// +kubebuilder:validation:Minimum:=1
	// +kubebuilder:validation:Maximum:=5
	// +kubebuilder:validation:Optional
	LogLevel int32 `json:"logLevel,omitempty"`

	// operatingNamespace is for restricting the zero-trust-workload-identity-manager operations to provided namespace.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:="zero-trust-workload-identity-manager"
	OperatingNamespace string `json:"operatingNamespace,omitempty"`

	// trustDomain to be used for the SPIFFE identifiers
	TrustDomain string `json:"trustDomain,omitempty"`

	// bundleConfigMap is Configmap name for Spire bundle
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=spire-bundle
	BundleConfigMap string `json:"bundleConfigMap"`

	// spireServerConfig has config for spire server.
	// +kubebuilder:validation:Optional
	SpireServerConfig *SpireServerConfig `json:"spireServerConfig,omitempty"`

	// spireAgentConfig has config for spire agents.
	// +kubebuilder:validation:Optional
	SpireAgentConfig *SpireAgentConfig `json:"spireAgentConfig,omitempty"`

	// spiffeOIDCProviderConfig has config for OIDC discovery provider
	// +kubebuilder:validation:Optional
	SpiffeOIDCProviderConfig *SpiffeOIDCProviderConfig `json:"spiffeOIDCProviderConfigMap,omitempty"`

	// spiffeCSIDriverConfig has config for spiffe csi driver.
	// +kubebuilder:validation:Optional
	SpiffeCSIDriverConfig *SpiffeCSIDriverConfig `json:"spiffeCSIDriverConfig,omitempty"`

	// resources are for defining the resource requirements.
	// Cannot be updated.
	// ref: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	// +kubebuilder:validation:Optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`

	// affinity is for setting scheduling affinity rules.
	// ref: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/
	// +kubebuilder:validation:Optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// tolerations are for setting the pod tolerations.
	// ref: https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/
	// +kubebuilder:validation:Optional
	// +listType=atomic
	Tolerations []*corev1.Toleration `json:"tolerations,omitempty"`

	// nodeSelector is for defining the scheduling criteria using node labels.
	// ref: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/
	// +kubebuilder:validation:Optional
	// +mapType=atomic
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
}

// SpireServerConfig defines the configuration for the Spire Server.
// +kubebuilder:validation:Optional
type SpireServerConfig struct {
	// Enabled specifies whether the Spire Server is enabled.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	Enabled bool `json:"enabled,omitempty"`

	// spireServerKeyManager has configs for the spire server key manager.
	// +kubebuilder:validation:Optional
	SpireServerKeyManager *SpireServerKeyManager `json:"spireServerKeyManager,omitempty"`

	// CASubject contains subject information for the Spire CA.
	// +kubebuilder:validation:Optional
	CASubject *CASubject `json:"caSubject,omitempty"`
}

// SpireServerKeyManager will contain configs for the spire server key manager
type SpireServerKeyManager struct {
	// diskEnabled is a flag to enable keyManager on disk.
	// +kubebuilder:default=true
	// +kubebuilder:validation:Optional
	DiskEnabled bool `json:"diskEnabled,omitempty"`

	// memoryEnabled is a flag to enable keyManager on memory
	// +kubebuilder:default=false
	// +kubebuilder:validation:Optional
	MemoryEnabled bool `json:"memoryEnabled,omitempty"`
}

// CASubject defines the subject information for the Spire CA.
// +kubebuilder:validation:Optional
type CASubject struct {
	// Country specifies the country for the CA.
	// +kubebuilder:validation:Optional
	Country string `json:"country,omitempty"`

	// Organization specifies the organization for the CA.
	// +kubebuilder:validation:Optional
	Organization string `json:"organization,omitempty"`

	// CommonName specifies the common name for the CA.
	// +kubebuilder:validation:Optional
	CommonName string `json:"commonName,omitempty"`
}

// SpireAgentConfig defines the configuration for the Spire Agent.
// +kubebuilder:validation:Optional
type SpireAgentConfig struct {
	// Enabled specifies whether the Spire Agent is enabled.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	Enabled bool `json:"enabled,omitempty"`

	// NodeAttestor specifies the configuration for the Node Attestor.
	// +kubebuilder:validation:Optional
	NodeAttestor *NodeAttestor `json:"nodeAttestor,omitempty"`

	// WorkloadAttestors specifies the configuration for the Workload Attestors.
	// +kubebuilder:validation:Optional
	WorkloadAttestors *WorkloadAttestors `json:"workloadAttestors,omitempty"`
}

// NodeAttestor defines the configuration for the Node Attestor.
// +kubebuilder:validation:Optional
type NodeAttestor struct {
	// k8sPSATEnabled tells if k8sPSAT configuration is enabled
	// +kubebuilder:default:=true
	K8sPSATEnabled bool `json:"k8sPSATEnabled,omitempty"`
}

// WorkloadAttestors defines the configuration for the Workload Attestors.
// +kubebuilder:validation:Optional
type WorkloadAttestors struct {

	// k8sEnabled explains if the configuration is enabled for k8s.
	// +kubebuilder:default=true
	K8sEnabled bool `json:"k8sEnabled,omitempty"`

	// workloadAttestorsVerification tells what kind of verification to do against kubelet.
	// auto will first attempt to use hostCert, and then fall back to apiServerCA.
	// Valid options are [auto, hostCert, apiServerCA, skip]
	// +kubebuilder:validation:Optional
	WorkloadAttestorsVerification *WorkloadAttestorsVerification `json:"workloadAttestorsVerification,omitempty"`

	// DisableContainerSelectors specifies whether to disable container selectors in the Kubernetes workload attestor.
	// Set to true if using holdApplicationUntilProxyStarts in Istio
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	DisableContainerSelectors bool `json:"disableContainerSelectors,omitempty"`

	// UseNewContainerLocator enables the new container locator algorithm that has support for cgroups v2.
	// Defaults to true
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	UseNewContainerLocator bool `json:"useNewContainerLocator,omitempty"`
}

type WorkloadAttestorsVerification struct {
	// Type specifies the type of verification to be used.
	// +kubebuilder:default="skip"
	Type string `json:"type,omitempty"`

	// hostCertBasePath specifies the base Path where kubelet places its certificates.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:"/var/lib/kubelet/pki"
	HostCertBasePath string `json:"hostCertBasePath,omitempty"`

	// hostCertFileName specifies the file name for the host certificate.
	// +kubebuilder:validation:Optional
	HostCertFileName string `json:"hostCertFileName,omitempty"`
}

// SpiffeCSIDriverConfig defines the configuration for the Spiffe CSI Driver.
// +kubebuilder:validation:Optional
type SpiffeCSIDriverConfig struct {
	// Enabled specifies whether the Spiffe CSI Driver is enabled.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=true
	Enabled bool `json:"enabled,omitempty"`

	// AgentSocket is the path to the agent socket.
	// +kubebuilder:validation:Optional
	AgentSocket string `json:"agentSocketPath,omitempty"`

	// PluginName defines the name of the CSI plugin.
	// +kubebuilder:validation:Optional
	PluginName string `json:"pluginName,omitempty"`
}

// SpiffeOIDCProviderConfig defines the configuration for the Spiffe OIDC Provider.
// +kubebuilder:validation:Optional
type SpiffeOIDCProviderConfig struct {
	// Enabled specifies whether the OIDC provider is enabled.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=true
	Enabled bool `json:"enabled,omitempty"`

	// AgentSocket is the name of the agent socket.
	// +kubebuilder:validation:Optional
	AgentSocket string `json:"agentSocketName,omitempty"`

	// ReplicaCount is the number of replicas for the OIDC provider.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=1
	ReplicaCount int `json:"replicaCount,omitempty"`

	// JwtIssuer specifies the JWT issuer.
	// +kubebuilder:validation:Optional
	JwtIssuer string `json:"jwtIssuer,omitempty"`
}

// OperatorControllerConfig is for configuring the operator for setting up
// defaults to install zero-trust-workload-identity-manager.
type OperatorControllerConfig struct {
	// namespace is for configuring the namespace to install the spire operand.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:="zero-trust-workload-identity-manager"
	Namespace string `json:"namespace,omitempty"`

	// labels to apply to all resources created for operands deployment.
	// +mapType=granular
	// +kubebuilder:validation:Optional
	Labels map[string]string `json:"labels,omitempty"`
}

// ZeroTrustWorkloadIdentityManagerStatus defines the observed state of ZeroTrustWorkloadIdentityManager
type ZeroTrustWorkloadIdentityManagerStatus struct {
	// conditions holds information of the current state of the external-secrets deployment.
	ConditionalStatus `json:",inline,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ZeroTrustWorkloadIdentityManager is the Schema for the zerotrustworkloadidentitymanagers API
type ZeroTrustWorkloadIdentityManager struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ZeroTrustWorkloadIdentityManagerSpec   `json:"spec,omitempty"`
	Status ZeroTrustWorkloadIdentityManagerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ZeroTrustWorkloadIdentityManagerList contains a list of ZeroTrustWorkloadIdentityManager
type ZeroTrustWorkloadIdentityManagerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ZeroTrustWorkloadIdentityManager `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ZeroTrustWorkloadIdentityManager{}, &ZeroTrustWorkloadIdentityManagerList{})
}
