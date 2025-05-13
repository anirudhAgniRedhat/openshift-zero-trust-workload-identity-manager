package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:validation:XValidation:rule="self.metadata.name == 'cluster'",message="SpireConfig is a singleton, .metadata.name must be 'cluster'"
// +operator-sdk:csv:customresourcedefinitions:displayName="SpireConfig"

// SpireConfig defines the detailed configuration of the SPIRE components
// managed by the ZeroTrustWorkloadIdentityManager.
//
// This CRD captures the user-facing configuration for the following SPIRE operands:
// - spire-server: trust domain, datastore, plugins, etc.
// - spire-agent: socket path, node attestation, workload attestors
// - spiffe-csi-driver: configuration for projected identity delivery
// - oidc-discovery-provider: optional OIDC endpoint configuration
//
// This CRD should be created and updated by cluster administrators or automation tools
// that require fine-grained control over the SPIRE deployment.
type SpireConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SpireConfigSpec   `json:"spec,omitempty"`
	Status            SpireConfigStatus `json:"status,omitempty"`
}

// SpireConfigStatus defines the observed state of spire-operand reconcilation
type SpireConfigStatus struct {
	// conditions holds information of the current state of the external-secrets deployment.
	ConditionalStatus `json:",inline,omitempty"`
}

// SpireConfigSpec is for configuring the spire-operands behavior.
type SpireConfigSpec struct {

	// trustDomain to be used for the SPIFFE identifiers
	// +kubebuilder:validation:Required
	TrustDomain string `json:"trustDomain,omitempty"`

	// clusterName will have the cluster name required to configure spire operands.
	// +kubebuilder:validation:Required
	ClusterName string `json:"clusterName,omitempty"`

	// bundleConfigMap is Configmap name for Spire bundle
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=spire-bundle
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

	// spireServerKeyManager has configs for the spire server key manager.
	// +kubebuilder:validation:Optional
	SpireServerKeyManager *SpireServerKeyManager `json:"spireServerKeyManager,omitempty"`

	// CASubject contains subject information for the Spire CA.
	// +kubebuilder:validation:Optional
	CASubject *CASubject `json:"caSubject,omitempty"`

	// resources are for defining the resource requirements.
	// Cannot be updated.
	// ref: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	// +kubebuilder:validation:Optional
	SpireServerResources *corev1.ResourceRequirements `json:"spireServerResources,omitempty"`

	// persistence has config for spire server volume related configs
	Persistence *Persistence `json:"persistence,omitempty"`

	SpireSQLConfig *SpireSQLConfig `json:"spireSQLConfig,omitempty"`
}

// Persistence defines volume-related settings.
type Persistence struct {
	// Type of volume to use for persistence.
	// +kubebuilder:validation:Enum=pvc;hostPath;emptyDir
	// +kubebuilder:default=pvc
	Type string `json:"type"`

	// Size of the persistent volume (e.g., 1Gi).
	// +kubebuilder:validation:Pattern=^[1-9][0-9]*Gi$
	// +kubebuilder:default="1Gi"
	Size string `json:"size"`

	// Access mode for the volume.
	// +kubebuilder:validation:Enum=ReadWriteOnce;ReadWriteOncePod;ReadWriteMany
	// +kubebuilder:default=ReadWriteOnce
	AccessMode string `json:"accessMode"`

	// StorageClass to be used for the PVC.
	// +kubebuilder:validation:optional
	// +kubebuilder:default:=null
	StorageClass *string `json:"storageClass,omitempty"`

	// Host path to be used when type is hostPath.
	// +kubebuilder:validation:optional
	// +kubebuilder:default=""
	HostPath string `json:"hostPath,omitempty"`
}

// SQLExternalSecret configures usage of external secrets for SQL credentials.
type SQLExternalSecret struct {
	// Whether to enable the external secret.
	// +kubebuilder:default=false
	Enabled bool `json:"enabled,omitempty"`

	// Secret name in the same namespace.
	// +kubebuilder:default=""
	Name string `json:"name,omitempty"`

	// Secret key whose value is the password.
	// +kubebuilder:default=""
	Key string `json:"key,omitempty"`
}

// SQLReadOnlyConfig allows configuring a read-only replica DB.
type SQLReadOnlyConfig struct {
	// Enable read-only SQL connection.
	// +kubebuilder:default=false
	Enabled bool `json:"enabled,omitempty"`

	// Host of the read-only DB.
	// +kubebuilder:default=""
	Host string `json:"host,omitempty"`

	// Port for the DB. If 0, defaults apply (5432/3306).
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=0
	Port int `json:"port,omitempty"`

	// Username to connect to DB.
	// +kubebuilder:default=spire
	Username string `json:"username,omitempty"`

	// Password for the username.
	// +kubebuilder:default=""
	Password string `json:"password,omitempty"`

	// DB options as key=value strings.
	// +kubebuilder:validation:optional
	// +kubebuilder:default={}
	Options []string `json:"options,omitempty"`

	// Optional external secret for read-only DB.
	ExternalSecret SQLExternalSecret `json:"externalSecret"`
}

// SpireSQLConfig configures the Spire SQL datastore backend.
type SpireSQLConfig struct {
	// Type of database to use.
	// +kubebuilder:validation:Enum=sqlite3;postgres;mysql;aws_postgresql;aws_mysql
	// +kubebuilder:default=sqlite3
	DatabaseType string `json:"databaseType"`

	// Only used if databaseType != sqlite3.
	// +kubebuilder:default=spire
	DatabaseName string `json:"databaseName,omitempty"`

	// Host of the database.
	// +kubebuilder:default=""
	Host string `json:"host,omitempty"`

	// Port for DB connection (defaults depend on type).
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=0
	Port int `json:"port,omitempty"`

	// Username for database connection.
	// +kubebuilder:default=spire
	Username string `json:"username,omitempty"`

	// Password for DB connection.
	// +kubebuilder:default=""
	Password string `json:"password,omitempty"`

	// Extra DB options.
	// +kubebuilder:validation:optional
	// +kubebuilder:default={}
	Options []string `json:"options,omitempty"`

	// MySQL TLS options.
	// +kubebuilder:default=""
	RootCAPath     string `json:"rootCAPath,omitempty"`
	ClientCertPath string `json:"clientCertPath,omitempty"`
	ClientKeyPath  string `json:"clientKeyPath,omitempty"`

	// External secret for main DB credentials.
	ExternalSecret *SQLExternalSecret `json:"externalSecret"`

	// DB pool config
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=100
	MaxOpenConns int `json:"maxOpenConns"`

	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=2
	MaxIdleConns int `json:"maxIdleConns"`

	// Max time (in seconds) a connection may live.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=0
	ConnMaxLifetime int `json:"connMaxLifetime"`

	// If true, disables DB auto-migration.
	// +kubebuilder:default=false
	DisableMigration bool `json:"disableMigration"`

	// AWS region, if using aws_* types.
	// +kubebuilder:default=""
	Region string `json:"region,omitempty"`

	// Optional read-only DB config.
	ReadOnly *SQLReadOnlyConfig `json:"readOnly"`
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
	// +kubebuilder: default="skip"
	Type string `json:"type,omitempty"`

	// hostCertBasePath specifies the base Path where kubelet places its certificates.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default="/var/lib/kubelet/pki"
	HostCertBasePath string `json:"hostCertBasePath,omitempty"`

	// hostCertFileName specifies the file name for the host certificate.
	// +kubebuilder:validation:Optional
	HostCertFileName string `json:"hostCertFileName,omitempty"`
}

// SpiffeCSIDriverConfig defines the configuration for the Spiffe CSI Driver.
// +kubebuilder:validation:Optional
type SpiffeCSIDriverConfig struct {

	// AgentSocket is the path to the agent socket.
	// +kubebuilder:default="/run/spire/agent-sockets/spire-agent.sock"
	AgentSocket string `json:"agentSocketPath,omitempty"`

	// PluginName defines the name of the CSI plugin.
	// +kubebuilder:default="csi.spiffe.io"
	PluginName string `json:"pluginName,omitempty"`
}

// SpiffeOIDCProviderConfig defines the configuration for the Spiffe OIDC Provider.
// +kubebuilder:validation:Optional
type SpiffeOIDCProviderConfig struct {

	// AgentSocket is the name of the agent socket.
	// +kubebuilder:default="spire-agent.sock"
	AgentSocketName string `json:"agentSocketName,omitempty"`

	// jwtIssuerPath to JWT issuer. Defaults to oidc-discovery.$trustDomain if unset
	// +kubebuilder:validation:Optional
	JwtIssuer string `json:"jwtIssuer,omitempty"`

	// spireOidcTlsConfig has the tls config specification for OIDC discovery provider
	// +kubebuilder:validation:Optional
	SpireOidcTlsConfig *TLSConfig `json:"spireOidcTlsConfig,omitempty"`

	// ReplicaCount is the number of replicas for the OIDC provider.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=1
	ReplicaCount int `json:"replicaCount,omitempty"`
}

// TLSConfig defines TLS configuration for securing components.
type TLSConfig struct {
	// Enable SPIRE integration to secure the oidc-discovery-provider.
	// +kubebuilder:default=true
	SpireEnabled bool `json:"spireEnabled"`

	// Configure external secret support for custom TLS certificate/key.
	ExternalSecret *ExternalSecretTLSConfig `json:"externalSecret"`

	// Configure cert-manager integration for automated certificate management.
	CertManager *CertManagerTLSConfig `json:"certManager"`
}

// ExternalSecretTLSConfig configures TLS via an externally provided Kubernetes Secret.
type ExternalSecretTLSConfig struct {
	// Enable usage of external Kubernetes TLS Secret.
	// +kubebuilder:default=false
	Enabled bool `json:"enabled"`

	// SecretName is the name of the TLS Secret to use.
	// +kubebuilder:default=""
	// +kubebuilder:validation:Optional
	SecretName string `json:"secretName,omitempty"`
}

// CertManagerTLSConfig configures TLS via cert-manager.
type CertManagerTLSConfig struct {
	// Enable TLS provisioning via cert-manager.
	// +kubebuilder:default=false
	Enabled bool `json:"enabled"`

	// Issuer contains cert-manager issuer configuration.
	Issuer *CertManagerIssuerConfig `json:"issuer"`

	// Certificate contains certificate-specific options.
	Certificate *CertManagerCertificateConfig `json:"certificate"`
}

// CertManagerIssuerConfig contains issuer-related configuration.
type CertManagerIssuerConfig struct {
	// Create indicators whether to create the issuer resource.
	// +kubebuilder:default=true
	Create bool `json:"create"`

	// ACME contains configuration for Let's Encrypt / ACME issuer.
	ACME *CertManagerACMEConfig `json:"acme"`
}

// CertManagerACMEConfig contains ACME-specific issuer settings.
type CertManagerACMEConfig struct {
	// Email address for Let's Encrypt registration (mandatory for ACME).
	// +kubebuilder:default=""
	// +kubebuilder:validation:Optional
	Email string `json:"email,omitempty"`

	// Server is the ACME server URL (production or staging).
	// +kubebuilder:default="https://acme-v02.api.letsencrypt.org/directory"
	// +kubebuilder:validation:Optional
	Server string `json:"server,omitempty"`
}

// CertManagerCertificateConfig configures certificate request details.
type CertManagerCertificateConfig struct {
	// DNSNames to include in the certificate request.
	// +kubebuilder:default={ }
	// +kubebuilder:validation:Optional
	DNSNames []string `json:"dnsNames,omitempty"`

	// IssuerRef configures which cert-manager Issuer/ClusterIssuer to use.
	IssuerRef *CertManagerIssuerRef `json:"issuerRef"`
}

// CertManagerIssuerRef is a reference to the cert-manager Issuer or ClusterIssuer.
type CertManagerIssuerRef struct {
	// Group is the API group of the issuer (e.g., "cert-manager.io").
	// +kubebuilder:default=""
	// +kubebuilder:validation:Optional
	Group string `json:"group,omitempty"`

	// Kind of the issuer resource (Issuer or ClusterIssuer).
	// +kubebuilder:default="Issuer"
	// +kubebuilder:validation:Optional
	Kind string `json:"kind,omitempty"`

	// Name of the issuer to be used.
	// +kubebuilder:default=""
	// +kubebuilder:validation:Optional
	Name string `json:"name,omitempty"`
}

// +kubebuilder:object:root=true

// SpireConfigLists contains a list of SpireConfig
type SpireConfigLists struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SpireConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SpireConfig{}, &SpireConfigLists{})
}
