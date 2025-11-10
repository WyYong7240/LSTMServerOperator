/*
Copyright 2025 wuyong7240.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LSTMPredictAppSpec defines the desired state of LSTMPredictApp
type LSTMPredictAppSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// The following markers will use OpenAPI v3 schema to validate the value
	// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html

	// foo is an example field of LSTMPredictApp. Edit lstmpredictapp_types.go to remove/update
	// +optional
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Schemaless
	AppImage      string `json:"appImage"`
	ContainerPort int32  `json:"containerPort"`
	// 必填项，但是用户可以不提供，由Webhook进行默认注入
	BackendAppReplicas *int32                      `json:"backendAppReplicas,omitempty"`
	ResourcesLimit     corev1.ResourceRequirements `json:"resourceLimit,omitempty"`
	// 必填项，但是用户可以不提供，由Webhook进行默认注入
	ServicePort int32 `json:"servicePort,omitempty"`
	// 必填项，但是用户可以不提供，由Webhook进行默认注入
	ServiceType corev1.ServiceType `json:"serviceType,omitempty"`
}

// LSTMPredictAppStatus defines the observed state of LSTMPredictApp.
type LSTMPredictAppStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// For Kubernetes API conventions, see:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	// conditions represent the current state of the LSTMPredictApp resource.
	// Each condition has a unique type and reflects the status of a specific aspect of the resource.
	//
	// Standard condition types include:
	// - "Available": the resource is fully functional
	// - "Progressing": the resource is being created or updated
	// - "Degraded": the resource failed to reach or maintain its desired state
	//
	// The status of each condition is one of True, False, or Unknown.
	// +optional
	ReadyReplicas   int32       `json:"readyReplicas,omitempty"`
	ServiceEndPoint string      `json:"serviceEndPoint,omitempty"`
	Phase           string      `json:"phase,omitempty"`
	LastUpdateTime  metav1.Time `json:"lastUpdateTime,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=lstmpredictapps,singular=lstmpredictapp,scope=Namespaced,shortName=lstmpa

// LSTMPredictApp is the Schema for the lstmpredictapps API
type LSTMPredictApp struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of LSTMPredictApp
	// +required
	Spec LSTMPredictAppSpec `json:"spec"`

	// status defines the observed state of LSTMPredictApp
	// +optional
	Status LSTMPredictAppStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// LSTMPredictAppList contains a list of LSTMPredictApp
type LSTMPredictAppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LSTMPredictApp `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LSTMPredictApp{}, &LSTMPredictAppList{})
}
