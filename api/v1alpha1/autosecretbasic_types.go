package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AutoSecretBasicSpec defines the desired state of AutoSecretBasic
type AutoSecretBasicSpec struct {
	// Username for the secret
	Username string `json:"username"`

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

	// Custom secret name (optional, defaults to metadata.name)
	// +optional
	SecretName string `json:"secretName,omitempty"`
}

// AutoSecretBasicStatus defines the observed state of AutoSecretBasic
type AutoSecretBasicStatus struct {
	// Name of the created secret
	SecretName string `json:"secretName,omitempty"`

	// Conditions represent the latest available observations
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=asb

// AutoSecretBasic is the Schema for the autosecretbasics API
type AutoSecretBasic struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AutoSecretBasicSpec   `json:"spec,omitempty"`
	Status AutoSecretBasicStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AutoSecretBasicList contains a list of AutoSecretBasic
type AutoSecretBasicList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AutoSecretBasic `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AutoSecretBasic{}, &AutoSecretBasicList{})
}
