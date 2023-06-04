/*
Copyright 2023.

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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NodeQuotaConfigSpec defines the desired state of NodeQuotaConfig
type NodeQuotaConfigSpec struct {
	ReservedHoursTolive int         `json:"reservedHoursToLive"`
	NodeGroupList       []NodeGroup `json:"nodeGroups"`
}

type ReservedResources struct {
	Resources corev1.ResourceList `json:"resources"`
	NodeGroup string              `json:"nodeGroup"`
	Timestamp metav1.Time         `json:"Timestamp,omitempty" protobuf:"bytes,8,opt,name=Timestamp"`
}

type NodeGroup struct {
	LabelSelector      map[string]string  `json:"labelSelector"`
	Name               string             `json:"name"`
	ResourceMultiplier map[string]float64 `json:"multipliers"`
}

// NodeQuotaConfigStatus defines the observed state of NodeQuotaConfig
type NodeQuotaConfigStatus struct {
	Conditions        []metav1.Condition             `json:"conditions"`
	ReservedResources map[string][]ReservedResources `json:"reservedResources"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// NodeQuotaConfig is the Schema for the nodequotaconfigs API
type NodeQuotaConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodeQuotaConfigSpec   `json:"spec,omitempty"`
	Status NodeQuotaConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NodeQuotaConfigList contains a list of NodeQuotaConfig
type NodeQuotaConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodeQuotaConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NodeQuotaConfig{}, &NodeQuotaConfigList{})
}
