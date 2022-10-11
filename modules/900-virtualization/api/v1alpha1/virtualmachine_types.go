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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k6tv1 "kubevirt.io/api/core/v1"
)

// VirtualMachineSpec defines the desired state of VirtualMachine
type VirtualMachineSpec struct {
	Running         *bool                         `json:"running,omitempty" optional:"true"`
	StaticIPAddress string                        `json:"staticIPAddress,omitempty"`
	Resources       v1.ResourceList               `json:"resources,omitempty"`
	UserName        string                        `json:"userName,omitempty"`
	SSHPublicKey    string                        `json:"sshPublicKey,omitempty"`
	BootDisk        BootDisk                      `json:"bootDisk,omitempty"`
	CloudInit       *k6tv1.CloudInitNoCloudSource `json:"cloudInit,omitempty"`
	Disks           *[]DiskSource                 `json:"disks,omitempty"`
}

// VirtualMachineStatus defines the observed state of VirtualMachine
type VirtualMachineStatus struct {
	// Phase is a human readable, high-level representation of the status of the virtual machine
	Phase k6tv1.VirtualMachinePrintableStatus `json:"phase,omitempty"`
	// NodeName is the name where the VirtualMachineInstance is currently running.
	NodeName string `json:"nodeName,omitempty"`
	// IP address of Virtual Machine
	VMIP string `json:"vmIP,omitempty"`
}

// Represents the source of a boot disk
// Only one of its members may be specified.
type BootDisk struct {
	// Represents the source of image used to create disk
	// +optional
	Image ImageSource `json:"image,omitempty"`
	// Represents the source of existing disk
	// +optional
	Disk DiskSource `json:"disk,omitempty"`
}

// Represents the source of image used to create disk
type ImageSource struct {
	// Type represents the type for newly created disk
	Type string `json:"type,omitempty"`
	// Type represents the size for newly created disk
	Size string `json:"size"`
	// Name represents the name of the Image
	Name string `json:"name"`
	// Scope represents the source of Image
	// supported values: global, private
	Scope ImageSourceScope `json:"scope,omitempty"`
	// Type represents the type for newly created disk
	Ephemeral bool `json:"ephemeral,omitempty"`
	// Bus indicates the type of disk device to emulate.
	// supported values: virtio, sata, scsi, usb.
	Bus string `json:"bus,omitempty"`
}

// ImageSourceScope represents the source of the image.
// +enum
type ImageSourceScope string

const (
	// ImageSourceScopeGlobal indicates that disk should be
	// created from global image. This is the default mode.
	ImageSourceScopeGlobal ImageSourceScope = "global"

	// ImageSourceScopePrivate indicates that disk should be
	// created from private image from the same namespace.
	ImageSourceScopePrivate ImageSourceScope = "private"
)

// Represents the source of existing disk
type DiskSource struct {
	// Name represents the name of the Disk in the same namespace
	Name string `json:"name"`
	// Hotpluggable indicates whether the volume can be hotplugged and hotunplugged.
	// +optional
	Hotpluggable bool `json:"hotpluggable,omitempty"`
	// Bus indicates the type of disk device to emulate.
	// supported values: virtio, sata, scsi, usb.
	Bus string `json:"bus,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// VirtualMachine is the Schema for the virtualmachines API
type VirtualMachine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VirtualMachineSpec   `json:"spec,omitempty"`
	Status VirtualMachineStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// VirtualMachineList contains a list of VirtualMachine
type VirtualMachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VirtualMachine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VirtualMachine{}, &VirtualMachineList{})
}