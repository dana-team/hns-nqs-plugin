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

// NodeQuotaConfigSpec defines the desired state of NodeQuotaConfig
type NodeQuotaConfigSpec struct {
	// ReservedHoursToLive defines how many hours the ReservedResources can live until they are removed from the cluster resources
	ReservedHoursToLive int `json:"reservedHoursToLive"`

	// ControlledResources defines which node resources are controlled
	// Possible values examples: ["cpu","memory"], ["cpu","gpu"]
	ControlledResources []string `json:"controlledResources"`

	// Roots defines the state of the cluster's secondary roots and roots
	Roots []SubnamespacesRoots `json:"subnamespacesRoots"`
}

// ReservedResources shows the resources of nodes that were deleted from the cluster but not from the subnamespace quota
type ReservedResources struct {
	// Resources defines the number of resources of the nodes
	Resources corev1.ResourceList `json:"resources,omitempty"`
	// NodeGroup defines which of the secondaryRoots the nodes that were removed was a part of
	NodeGroup string `json:"nodeGroup,omitempty"`
	// Timestamp defines when the nodes were removed
	Timestamp metav1.Time `json:"Timestamp,omitempty" protobuf:"bytes,8,opt,name=Timestamp"`
}

// SubnamespacesRoots define the root and secondary root of the cluster's hierarchy
type SubnamespacesRoots struct {
	// RootNamespace is the name of the root namespace
	RootNamespace string `json:"rootNamespace"`
	// SecondaryRoots are the subnamespaces under the root namespace
	SecondaryRoots []NodeGroup `json:"secondaryRoots"`
}

// NodeGroup defines a group of nodes that allocated to the secondary root workloads
type NodeGroup struct {
	// LabelSelector defines the label selector of the nodes and how to find them.
	// Possible values examples: {"app":"gpu-nodes"}
	LabelSelector map[string]string `json:"labelSelector"`
	// Name is the name of the secondaryRoot.
	Name string `json:"name"`
	// ResourceMultiplier defines the multiplier that will be used when calculating the resources of nodes for allowing overcommit
	// Possible values examples: {"cpu":2, "memory":3} {"cpu":3, "gpu":3}
	ResourceMultiplier map[string]string `json:"multipliers,omitempty"`
}

// NodeQuotaConfigStatus defines the observed state of NodeQuotaConfig
type NodeQuotaConfigStatus struct {
	Conditions        []metav1.Condition  `json:"conditions,omitempty"`
	ReservedResources []ReservedResources `json:"reservedResources,omitempty"`
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
