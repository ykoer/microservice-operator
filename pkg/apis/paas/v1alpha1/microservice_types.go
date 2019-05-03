package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MicroServiceSpec defines the desired state of MicroService
type MicroServiceSpec struct {
	GitSource GitSource `json:"source"`
	Image     string    `json:"image"`
	Hostname  string    `json:"hostname"`
	Replicas  int32     `json:"replicas"`
}

// GitSource defines a git URI and branch
type GitSource struct {
	URI string `json:"uri"`
	Ref string `json:"ref"`
}

// MicroServiceStatus defines the observed state of MicroService
type MicroServiceStatus struct {
	Status string `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MicroService is the Schema for the microservices API
// +k8s:openapi-gen=true
type MicroService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MicroServiceSpec   `json:"spec,omitempty"`
	Status MicroServiceStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MicroServiceList contains a list of MicroService
type MicroServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MicroService `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MicroService{}, &MicroServiceList{})
}
