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
	"github.com/flant/addon-operator/pkg/module_manager/go_hook"
	"github.com/flant/addon-operator/pkg/module_manager/go_hook/metrics"
	"github.com/flant/addon-operator/sdk"
	"github.com/flant/shell-operator/pkg/kube_events_manager/types"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var _ = sdk.RegisterFunc(&go_hook.HookConfig{
	Queue:        "/modules/deckhouse/metrics_for_alerts",
	OnBeforeHelm: &go_hook.OrderedConfig{Order: 10},
	Kubernetes: []go_hook.KubernetesConfig{
		{
			Name:       "storageclasses",
			ApiVersion: "storage.k8s.io/v1",
			Kind:       "Storageclass",
			FilterFunc: applyStorageClasssesFilter,
		},
	},
}, storageRetentionMetricHandler)

func applyStorageClasssesFilter(obj *unstructured.Unstructured) (go_hook.FilterResult, error) {
	storageClass := &storagev1.StorageClass{}

	err := sdk.FromUnstructured(obj, storageClass)
	if err != nil {
		return nil, fmt.Errorf("cannot convert kubernetes object: %v", err)
	}

	var isDefault bool

	if storageClass.Annotations["storageclass.beta.kubernetes.io/is-default-class"] == "true" {
		isDefault = true
	}

	if storageClass.Annotations["storageclass.kubernetes.io/is-default-class"] == "true" {
		isDefault = true
	}

	return DefaultStorageClass{
		Name:      storageClass.Name,
		IsDefault: isDefault,
	}, nil
}


func storageRetentionMetricHandler(input *go_hook.HookInput) error {
	retentionDaysMain := input.Values.Get("prometheus.retentionDays")
	retentionDaysLongterm := input.Values.Get("prometheus.longtermRetentionDays")

	input.MetricsCollector.Expire("prometheus_disk_hook")

	input.MetricsCollector.Set(
		"d8_prometheus_storage_retention_days",
		retentionDaysMain.Float(),
		map[string]string{
			"prometheus": "main",
		},
		metrics.WithGroup("prometheus_disk_hook"),
	)

	input.MetricsCollector.Set(
		"d8_prometheus_storage_retention_days",
		retentionDaysLongterm.Float(),
		map[string]string{
			"prometheus": "longterm",
		},
		metrics.WithGroup("prometheus_disk_hook"),
	)

	return nil
}
