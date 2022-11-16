/*
Copyright 2022.

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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// VirtualMachineDiskSpec defines the desired state of VirtualMachineDisk
type VirtualMachineDiskSpec struct {
	StorageClassName string                            `json:"storageClassName,omitempty"`
	Size             resource.Quantity                 `json:"size,omitempty"`
	Source           *corev1.TypedLocalObjectReference `json:"source,omitempty"`
	VMName           string                            `json:"vmName,omitempty"`
	Ephemeral        bool                              `json:"ephemeral,omitempty"`
}

// VirtualMachineDiskStatus defines the observed state of VirtualMachineDisk
type VirtualMachineDiskStatus struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName={"vmd","vmdisk","vmdisks"}
//+kubebuilder:printcolumn:JSONPath=".spec.ephemeral",name=Ephemeral,type=string
//+kubebuilder:printcolumn:JSONPath=".spec.vmName",name=VM,type=string

// VirtualMachineDisk is the Schema for the virtualmachinedisks API
type VirtualMachineDisk struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VirtualMachineDiskSpec   `json:"spec,omitempty"`
	Status VirtualMachineDiskStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// VirtualMachineDiskList contains a list of VirtualMachineDisk
type VirtualMachineDiskList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VirtualMachineDisk `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VirtualMachineDisk{}, &VirtualMachineDiskList{})
}
