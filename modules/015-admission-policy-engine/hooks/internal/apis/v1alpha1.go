package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type OperationPolicy struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the behavior of a node group.
	Spec OperationPolicySpec `json:"spec"`

	// Most recently observed status of the node.
	// Populated by the system.

	Status OperationPolicyStatus `json:"status,omitempty"`
}

type OperationPolicySpec struct {
	Policies struct {
		AllowedRepos   []string `json:"allowedRepos"`
		RequiredLabels []string `json:"requiredLabels"`
	} `json:"policies"`
	Match struct {
		NamespaceSelector NamespaceSelector    `json:"namespaceSelector,omitempty"`
		LabelSelector     metav1.LabelSelector `json:"labelSelector,omitempty"`
	} `json:"match"`
}

type OperationPolicyStatus struct {
}

type NamespaceSelector struct {
	MatchNames   []string `json:"matchNames"`
	ExcludeNames []string `json:"excludeNames"`

	LabelSelector metav1.LabelSelector `json:"labelSelector,omitempty"`
}
