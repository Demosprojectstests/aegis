package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TenantSpec defines the desired state of Tenant.
type TenantSpec struct {
	// Plan is starter or pro. Controls quota size.
	// +kubebuilder:validation:Enum=starter;pro
	// +kubebuilder:default=starter
	Plan string `json:"plan,omitempty"`
}

// TenantStatus defines the observed state of Tenant.
type TenantStatus struct {
	Namespace string `json:"namespace,omitempty"`
	Ready     bool   `json:"ready"`
	Message   string `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:printcolumn:name="Plan",type=string,JSONPath=.spec.plan
// +kubebuilder:printcolumn:name="Namespace",type=string,JSONPath=.status.namespace
// +kubebuilder:printcolumn:name="Ready",type=boolean,JSONPath=.status.ready

type Tenant struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              TenantSpec   `json:"spec,omitempty"`
	Status            TenantStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type TenantList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Tenant `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Tenant{}, &TenantList{})
}
