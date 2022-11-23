/*
Copyright 2022 Flant JSC

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

package hooks

import (
	"k8s.io/utils/pointer"

	v1alpha1 "github.com/deckhouse/deckhouse/modules/015-admission-policy-engine/hooks/internal/apis"

	"github.com/flant/addon-operator/pkg/module_manager/go_hook"
	"github.com/flant/addon-operator/sdk"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var _ = sdk.RegisterFunc(&go_hook.HookConfig{
	OnAfterHelm: &go_hook.OrderedConfig{Order: 30},
	Kubernetes: []go_hook.KubernetesConfig{
		{
			Name:                         "operation-policies",
			ApiVersion:                   "deckhouse.io/v1alpha1",
			Kind:                         "OperationPolicy",
			ExecuteHookOnSynchronization: pointer.Bool(false),
			FilterFunc:                   filterOP,
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
