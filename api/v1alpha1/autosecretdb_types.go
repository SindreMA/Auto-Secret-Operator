package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AutoSecretDbSpec defines the desired state of AutoSecretDb
type AutoSecretDbSpec struct {
	// Username for database authentication
	Username string `json:"username"`

	// Database name
	DBName string `json:"dbname"`

	// Database host/cluster name (FQDN)
	DBHost string `json:"dbhost"`

	// Port (optional, defaults to 5432)
	// +optional
	// +kubebuilder:default=5432
	Port int32 `json:"port,omitempty"`

	// Password length (optional, defaults to 30)
	// +optional
	// +kubebuilder:default=30
	// +kubebuilder:validation:Minimum=8
	// +kubebuilder:validation:Maximum=128
	PasswordLength int32 `json:"passwordLength,omitempty"`

	// Password charset (optional, defaults to "hex")
	// Options: "alphanumeric", "ascii-printable", "hex", "base64"
	// +optional
	// +kubebuilder:default="hex"
	// +kubebuilder:validation:Enum=alphanumeric;ascii-printable;hex;base64
	PasswordCharset string `json:"passwordCharset,omitempty"`

	// Database type (optional, defaults to "postgresql")
	// +optional
	// +kubebuilder:default="postgresql"
	DBType string `json:"dbType,omitempty"`

	// Additional connection parameters (optional)
	// +optional
	AdditionalParams string `json:"additionalParams,omitempty"`

	// Custom secret name (optional, defaults to metadata.name)
	// +optional
	SecretName string `json:"secretName,omitempty"`
}

// AutoSecretDbStatus defines the observed state of AutoSecretDb
type AutoSecretDbStatus struct {
	// Name of the created secret
	SecretName string `json:"secretName,omitempty"`

	// Conditions represent the latest available observations
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=asdb

// AutoSecretDb is the Schema for the autosecretdbs API
type AutoSecretDb struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AutoSecretDbSpec   `json:"spec,omitempty"`
	Status AutoSecretDbStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AutoSecretDbList contains a list of AutoSecretDb
type AutoSecretDbList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AutoSecretDb `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AutoSecretDb{}, &AutoSecretDbList{})
}
