package spire_server

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openshift/zero-trust-workload-identity-manager/api/v1alpha1"
	"github.com/openshift/zero-trust-workload-identity-manager/pkg/controller/utils"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GenerateSpireServerConfigMap generates the spire-server ConfigMap
func GenerateSpireServerConfigMap(config *v1alpha1.SpireServerConfigSpec) (*corev1.ConfigMap, error) {
	confMap := generateServerConfMap(config)
	confJSON, err := marshalToJSON(confMap)
	if err != nil {
		return nil, err
	}
	labels := map[string]string{}
	labels[utils.AppManagedByLabelKey] = utils.AppManagedByLabelValue
	for key, value := range config.Labels {
		labels[key] = value
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "spire-server",
			Namespace: utils.OperatorNamespace,
			Labels:    labels,
		},
		Data: map[string]string{
			"server.conf": string(confJSON),
		},
	}

	return cm, nil
}

// generateServerConfMap builds the server.conf structure as a Go map
func generateServerConfMap(config *v1alpha1.SpireServerConfigSpec) map[string]interface{} {
	return map[string]interface{}{
		"health_checks": map[string]interface{}{
			"bind_address":     "0.0.0.0",
			"bind_port":        "8080",
			"listener_enabled": true,
			"live_path":        "/live",
			"ready_path":       "/ready",
		},
		"plugins": map[string]interface{}{
			"DataStore": []map[string]interface{}{
				{
					"sql": map[string]interface{}{
						"plugin_data": map[string]interface{}{
							"connection_string": config.Datastore.ConnectionString,
							"database_type":     config.Datastore.DatabaseType,
							"disable_migration": config.Datastore.DisableMigration,
							"max_idle_conns":    config.Datastore.MaxIdleConns,
							"max_open_conns":    config.Datastore.MaxOpenConns,
						},
					},
				},
			},
			"KeyManager": []map[string]interface{}{
				{
					"disk": map[string]interface{}{
						"plugin_data": map[string]interface{}{
							"keys_path": "/run/spire/data/keys.json",
						},
					},
				},
			},
			"NodeAttestor": []map[string]interface{}{
				{
					"k8s_psat": map[string]interface{}{
						"plugin_data": map[string]interface{}{
							"clusters": []map[string]interface{}{
								{
									config.ClusterName: map[string]interface{}{
										"allowed_node_label_keys": []string{},
										"allowed_pod_label_keys":  []string{},
										"audience":                []string{"spire-server"},
										"service_account_allow_list": []string{
											"zero-trust-workload-identity-manager:spire-agent",
										},
									},
								},
							},
						},
					},
				},
			},
			"Notifier": []map[string]interface{}{
				{
					"k8sbundle": map[string]interface{}{
						"plugin_data": map[string]interface{}{
							"config_map": config.BundleConfigMap,
							"namespace":  utils.OperatorNamespace,
						},
					},
				},
			},
		},
		"server": map[string]interface{}{
			"audit_log_enabled": false,
			"bind_address":      "0.0.0.0",
			"bind_port":         "8081",
			"ca_key_type":       "rsa-2048",
			"ca_subject": []map[string]interface{}{
				{
					"common_name":  config.CASubject.CommonName,
					"country":      []string{config.CASubject.Country},
					"organization": []string{config.CASubject.Organization},
				},
			},
			"ca_ttl":                "24h",
			"data_dir":              "/run/spire/data",
			"default_jwt_svid_ttl":  "1h",
			"default_x509_svid_ttl": "4h",
			"jwt_issuer":            config.JwtIssuer,
			"log_level":             "debug",
			"trust_domain":          config.TrustDomain,
		},
	}
}

// marshalToJSON marshals a map to JSON with indentation
func marshalToJSON(data map[string]interface{}) ([]byte, error) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal server.conf: %w", err)
	}
	return jsonBytes, nil
}

// generateConfigHash returns a SHA256 hex string of the trimmed input string
func generateConfigHashFromString(data string) string {
	normalized := strings.TrimSpace(data) // Removes leading/trailing whitespace and newlines
	return generateConfigHash([]byte(normalized))
}

// generateConfigHash returns a SHA256 hex string of the trimmed input bytes
func generateConfigHash(data []byte) string {
	normalized := strings.TrimSpace(string(data)) // Convert to string, trim, convert back to bytes
	hash := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(hash[:])
}

func generateSpireControllerManagerConfigYaml(config *v1alpha1.SpireServerConfigSpec) string {
	return fmt.Sprintf(`apiVersion: spire.spiffe.io/v1alpha1
kind: ControllerManagerConfig
metadata:
  name: spire-controller-manager
  namespace: zero-trust-workload-identity-manager
  labels:
    app.kubernetes.io/name: server
    app.kubernetes.io/instance: spire
    app.kubernetes.io/version: "1.12.0"
    app.kubernetes.io/managed-by: zero-trust-workload-identity-manager
metrics:
  bindAddress: 0.0.0.0:8082
health:
  healthProbeBindAddress: 0.0.0.0:8083
validatingWebhookConfigurationName: spire-controller-manager-webhook
entryIDPrefix: %s
clusterName: %s
trustDomain: %s
ignoreNamespaces:
  - kube-system
  - kube-public
  - local-path-storage
  - openshift-cluster-node-tuning-operator
  - openshift-cluster-samples-operator
  - openshift-cluster-storage-operator
  - openshift-console-operator
  - openshift-console
  - openshift-dns
  - openshift-dns-operator
  - openshift-image-registry
  - openshift-ingress
  - openshift-kube-storage-version-migrator
  - openshift-kube-storage-version-migrator-operator
  - openshift-kube-proxy
  - openshift-marketplace
  - openshift-monitoring
  - openshift-multus
  - openshift-network-diagnostics
  - openshift-network-operator
  - openshift-operator-lifecycle-manager
  - openshift-roks-metrics
  - openshift-service-ca-operator
  - openshift-service-ca
spireServerSocketPath: "/tmp/spire-server/private/api.sock"
watchClassless: false
parentIDTemplate: "spiffe://{{ .TrustDomain }}/spire/agent/k8s_psat/{{ .ClusterName }}/{{ .NodeMeta.UID }}"
reconcile:
  clusterSPIFFEIDs: true
  clusterStaticEntries: true
  clusterFederatedTrustDomains: true
`, config.ClusterName, config.ClusterName, config.TrustDomain)
}

func generateControllerManagerConfigMap(configYAML string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "spire-controller-manager",
			Namespace: utils.OperatorNamespace,
			Labels: map[string]string{
				"app":                      "spire-controller-manager",
				utils.AppManagedByLabelKey: utils.AppManagedByLabelValue,
			},
		},
		Data: map[string]string{
			"controller-manager-config.yaml": configYAML,
		},
	}
}

func generateSpireBundleConfigMap() *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "spire-bundle",
			Namespace: utils.OperatorNamespace,
			Labels: map[string]string{
				"app":                      "spire-server",
				utils.AppManagedByLabelKey: utils.AppManagedByLabelValue,
			},
		},
	}
}
