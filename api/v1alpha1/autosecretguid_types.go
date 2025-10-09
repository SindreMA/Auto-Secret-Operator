package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AutoSecretGuidSpec defines the desired state of AutoSecretGuid
type AutoSecretGuidSpec struct {
	// GUID format (optional, defaults to "uuidv4")
	// Options: "uuidv4", "short-uuid", "uuidv7"
	// +optional
	// +kubebuilder:default="uuidv4"
	// +kubebuilder:validation:Enum=uuidv4;short-uuid;uuidv7
	Format string `json:"format,omitempty"`

	// Custom secret name (optional, defaults to metadata.name)
	// +optional
	SecretName string `json:"secretName,omitempty"`
}

// AutoSecretGuidStatus defines the observed state of AutoSecretGuid
type AutoSecretGuidStatus struct {
	// Name of the created secret
	SecretName string `json:"secretName,omitempty"`

	// The generated GUID
	GUID string `json:"guid,omitempty"`

	// Conditions represent the latest available observations
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=asg

// AutoSecretGuid is the Schema for the autosecretguids API
type AutoSecretGuid struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AutoSecretGuidSpec   `json:"spec,omitempty"`
	Status AutoSecretGuidStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AutoSecretGuidList contains a list of AutoSecretGuid
type AutoSecretGuidList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AutoSecretGuid `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AutoSecretGuid{}, &AutoSecretGuidList{})
}
