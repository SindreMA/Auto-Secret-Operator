package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AutoSecretSpec defines the desired state of AutoSecret
type AutoSecretSpec struct {
	// Username for database authentication
	Username string `json:"username"`

	// Database name
	DBName string `json:"dbname"`

	// Database host/cluster name
	DBHost string `json:"dbhost"`

	// Port (optional, defaults to 5432)
	// +optional
	Port int `json:"port,omitempty"`
}

// AutoSecretStatus defines the observed state of AutoSecret
type AutoSecretStatus struct {
	// Name of the created basic-auth secret
	BasicAuthSecretName string `json:"basicAuthSecretName,omitempty"`

	// Name of the created DB URI secret
	DBURISecretName string `json:"dbURISecretName,omitempty"`

	// Conditions represent the latest available observations
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=as

// AutoSecret is the Schema for the autosecrets API
type AutoSecret struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AutoSecretSpec   `json:"spec,omitempty"`
	Status AutoSecretStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AutoSecretList contains a list of AutoSecret
type AutoSecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AutoSecret `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AutoSecret{}, &AutoSecretList{})
}
