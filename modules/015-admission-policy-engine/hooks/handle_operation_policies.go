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
	"fmt"
	"time"

	v1alpha1 "github.com/deckhouse/deckhouse/modules/015-admission-policy-engine/hooks/internal/apis"

	"github.com/flant/addon-operator/pkg/module_manager/go_hook"
	"github.com/flant/addon-operator/sdk"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var _ = sdk.RegisterFunc(&go_hook.HookConfig{
	Queue: "/modules/admission-policy-engine/operation_policies",
	Kubernetes: []go_hook.KubernetesConfig{
		{
			Name:       "templates",
			ApiVersion: "templates.gatekeeper.sh/v1",
			Kind:       "ConstraintTemplate",
			FilterFunc: filterTemplates,
		},
		{
			Name:       "operation-policies",
			ApiVersion: "deckhouse.io/v1alpha1",
			Kind:       "OperationPolicy",
			FilterFunc: filterOP,
		},
	},
	Settings: &go_hook.HookConfigSettings{
		ExecutionMinInterval: 3 * time.Second,
		ExecutionBurst:       5,
	},
}, handleOP)

func handleOP(input *go_hook.HookInput) error {
	result := make([]*operationPolicy, 0)

	bootstrapped := input.Values.Get("admissionPolicyEngine.internal.bootstrapped").Bool()
	if !bootstrapped {
		return nil
	}

	snap := input.Snapshots["templates"]
	if len(snap) == 0 {
		input.Values.Set("admissionPolicyEngine.internal.operationPolicies", result)
		return nil
	}

	kindMap := make(map[string]struct{}, len(snap))
	for _, sn := range snap {
		kind := sn.(string)
		kindMap[kind] = struct{}{}
	}

	snap = input.Snapshots["operation-policies"]

	for _, sn := range snap {
		op := sn.(*operationPolicy)
		if _, ok := kindMap[op.Kind]; !ok {
			// skip constraint if crd was not created by gatekeeper yet
			continue
		}
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

func filterTemplates(obj *unstructured.Unstructured) (go_hook.FilterResult, error) {
	kind, found, err := unstructured.NestedString(obj.Object, "spec", "crd", "spec", "names", "kind")
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("kind for ConstraintTemplate: %s not found", obj.GetName())
	}

	return kind, nil
}

type operationPolicy struct {
	Kind     string `json:"kind"`
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
	Spec v1alpha1.OperationPolicySpec `json:"spec"`
}
