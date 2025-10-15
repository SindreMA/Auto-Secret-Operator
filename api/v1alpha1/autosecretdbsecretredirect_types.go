package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AutoSecretDbSecretRedirectSpec defines the desired state of AutoSecretDbSecretRedirect
type AutoSecretDbSecretRedirectSpec struct {
	// SecretName is the name of the source secret to watch
	SecretName string `json:"secretname"`

	// TargetSecretName is the optional name for the created secret
	// If not specified, defaults to <secretname>-redirect
	// +optional
	TargetSecretName string `json:"targetSecretName,omitempty"`
}

// AutoSecretDbSecretRedirectStatus defines the observed state of AutoSecretDbSecretRedirect
type AutoSecretDbSecretRedirectStatus struct {
	// TargetSecretName is the name of the created secret
	TargetSecretName string `json:"targetSecretName,omitempty"`

	// SourceSecretResourceVersion tracks the last processed version of the source secret
	SourceSecretResourceVersion string `json:"sourceSecretResourceVersion,omitempty"`

	// Conditions represent the latest available observations
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=asdbsr

// AutoSecretDbSecretRedirect is the Schema for the autosecretdbsecretredirects API
type AutoSecretDbSecretRedirect struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AutoSecretDbSecretRedirectSpec   `json:"spec,omitempty"`
	Status AutoSecretDbSecretRedirectStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AutoSecretDbSecretRedirectList contains a list of AutoSecretDbSecretRedirect
type AutoSecretDbSecretRedirectList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AutoSecretDbSecretRedirect `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AutoSecretDbSecretRedirect{}, &AutoSecretDbSecretRedirectList{})
}
