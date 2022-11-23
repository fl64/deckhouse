package hooks

import (
	v1alpha1 "github.com/deckhouse/deckhouse/modules/015-admission-policy-engine/hooks/internal/apis"

	"github.com/flant/addon-operator/pkg/module_manager/go_hook"
	"github.com/flant/addon-operator/sdk"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var _ = sdk.RegisterFunc(&go_hook.HookConfig{
	OnBeforeHelm: &go_hook.OrderedConfig{Order: 30},
	Kubernetes: []go_hook.KubernetesConfig{
		{
			Name:       "operation-policies",
			ApiVersion: "deckhouse.io/v1alpha1",
			Kind:       "OperationPolicy",
			FilterFunc: filterOP,
		},
	},
}, handleOP)

func handleOP(input *go_hook.HookInput) error {
	snap := input.Snapshots["operation-policies"]

	result := make([]*operationPolicy, 0, len(snap))

	// if len(snap) == 0 {
	// 	input.Values.Set("admissionPolicyEngine.internal.operationPolicies",result )
	// 	return nil
	// }

	for _, sn := range snap {
		op := sn.(*operationPolicy)
		result = append(result, op)
	}

	input.Values.Set("admissionPolicyEngine.internal.operationPolicies", result)

	return nil
}

func filterOP(obj *unstructured.Unstructured) (go_hook.FilterResult, error) {
	var op operationPolicy

	err := sdk.FromUnstructured(obj, &op)
	if err != nil {
		return nil, err
	}

	return &op, nil
}

type operationPolicy struct {
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
	Spec v1alpha1.OperationPolicySpec `json:"spec"`
}
