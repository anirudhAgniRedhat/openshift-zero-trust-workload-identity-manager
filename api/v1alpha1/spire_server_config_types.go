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
// +kubebuilder:validation:XValidation:rule="self.metadata.name == 'cluster'",message="SpireServerConfig is a singleton, .metadata.name must be 'cluster'"
// +operator-sdk:csv:customresourcedefinitions:displayName="SpireServerConfig"

// SpireServerConfig defines the configuration for the SPIRE Server managed by zero trust workload identity manager.
// This includes details related to trust domain, data storage, plugins
// and other configs required for workload authentication.
type SpireServerConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SpireServerConfigSpec   `json:"spec,omitempty"`
	Status            SpireServerConfigStatus `json:"status,omitempty"`
}

// SpireServerConfigSpec will have specifications for configuration related to the spire server.
type SpireServerConfigSpec struct {

	// trustDomain to be used for the SPIFFE identifiers
	// +kubebuilder:validation:Required
	TrustDomain string `json:"trustDomain,omitempty"`

	// clusterName will have the cluster name required to configure spire server.
	// +kubebuilder:validation:Required
	ClusterName string `json:"clusterName,omitempty"`

	// bundleConfigMap is Configmap name for Spire bundle, it sets the trust domain to be used for the SPIFFE identifiers
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=spire-bundle
	BundleConfigMap string `json:"bundleConfigMap"`

	// JwtIssuer is the JWT issuer domain. Defaults to oidc-discovery.$trustDomain if unset
	// +kubebuilder:validation:Optional
	JwtIssuer string `json:"jwtIssuer,omitempty"`

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
	// +kubebuilder:validation:Optional
	Persistence *Persistence `json:"persistence,omitempty"`

	// spireSQLConfig has the config required for the spire server SQL DataStore.
	// +kubebuilder:validation:Optional
	SpireSQLConfig *SpireSQLConfig `json:"spireSQLConfig,omitempty"`

	// labels to apply to all resources created for operator deployment.
	// +mapType=granular
	// +kubebuilder:validation:Optional
	Labels map[string]string `json:"labels,omitempty"`

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

// SpireServerConfigStatus defines the observed state of spire-server related reconciliation made by operator
type SpireServerConfigStatus struct {
	// conditions holds information of the current state of the spire-server resources.
	ConditionalStatus `json:",inline,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SpireServerConfigLists contain the list of SpireServerConfig
type SpireServerConfigLists struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SpireServerConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SpireServerConfig{}, &SpireServerConfigLists{})
}
