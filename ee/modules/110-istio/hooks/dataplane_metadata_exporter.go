/*
Copyright 2022 Flant JSC
Licensed under the Deckhouse Platform Enterprise Edition (EE) license. See https://github.com/deckhouse/deckhouse/blob/main/ee/LICENSE
*/

package hooks

import (
	"encoding/json"
	"fmt"

	"github.com/flant/addon-operator/pkg/module_manager/go_hook"
	"github.com/flant/addon-operator/pkg/module_manager/go_hook/metrics"
	"github.com/flant/addon-operator/sdk"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/deckhouse/deckhouse/ee/modules/110-istio/hooks/internal"
	"github.com/deckhouse/deckhouse/ee/modules/110-istio/hooks/internal/istio_versions"
)

const (
	istioRevsionAbsent           = "absent"
	istioVersionAbsent           = "absent"
	istioVersionUnknown          = "unknown"
	istioPodMetadataMetricName   = "d8_istio_dataplane_metadata"
	metadataExporterMetricsGroup = "metadata"
	autoUpgradeLabelName         = "istio.deckhouse.io/auto-upgrade"
	patchTemplate                = `{ "spec": { "template": { "metadata": { "annotations": { "istio.deckhouse.io/version": "%s" } } } } }`
)

var _ = sdk.RegisterFunc(&go_hook.HookConfig{
	Queue: internal.Queue("dataplane-metadata"),
	Kubernetes: []go_hook.KubernetesConfig{
		{
			Name:       "namespaces_global_revision",
			ApiVersion: "v1",
			Kind:       "Namespace",
			FilterFunc: applyNamespaceFilter, // from revisions_discovery.go
			LabelSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"istio-injection": "enabled"},
			},
		},
		{
			Name:       "namespaces_definite_revision",
			ApiVersion: "v1",
			Kind:       "Namespace",
			FilterFunc: applyNamespaceFilter, // from revisions_discovery.go
			LabelSelector: &metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Key:      "istio.io/rev",
						Operator: "Exists",
					},
				},
			},
		},
		{
			Name:       "istio_pod",
			ApiVersion: "v1",
			Kind:       "Pod",
			FilterFunc: applyIstioPodFilter,
			LabelSelector: &metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Key:      "job-name",
						Operator: "DoesNotExist",
					},
					{
						Key:      "heritage",
						Operator: "NotIn",
						Values:   []string{"upmeter"},
					},
					{
						Key:      "sidecar.istio.io/inject",
						Operator: "NotIn",
						Values:   []string{"false"},
					},
				},
			},
		},
		{
			Name:       "deployment",
			ApiVersion: "apps/v1",
			Kind:       "Deployment",
			FilterFunc: applyIstioDeploymentFilter,
		},
		{
			Name:       "daemonset",
			ApiVersion: "apps/v1",
			Kind:       "DaemonSet",
			FilterFunc: applyIstioDaemonSetFilter,
		},
		{
			Name:       "statefulset",
			ApiVersion: "apps/v1",
			Kind:       "StatefulSet",
			FilterFunc: applyIstioStatefulSetFilter,
		},
		{
			Name:       "replicaset",
			ApiVersion: "apps/v1",
			Kind:       "ReplicaSet",
			FilterFunc: applyIstioReplicaSetFilter,
		},
	},
}, dataplaneController)

type IstioNamespace struct {
	Revision               string
	AutoUpgradeLabelExists bool
}

// Needed to extend v1.Pod with our methods
type IstioDrivenPod v1.Pod

// Current istio revision is located in `sidecar.istio.io/status` annotation
type IstioPodStatus struct {
	Revision string `json:"revision"`
	// ... we aren't interested in the other fields
}

func (p *IstioDrivenPod) getIstioCurrentRevision() string {
	var istioStatusJSON string
	var istioPodStatus IstioPodStatus
	var revision string
	var ok bool

	if istioStatusJSON, ok = p.Annotations["sidecar.istio.io/status"]; ok {
		_ = json.Unmarshal([]byte(istioStatusJSON), &istioPodStatus)

		if istioPodStatus.Revision != "" {
			revision = istioPodStatus.Revision
		} else {
			revision = istioRevsionAbsent
		}
	} else {
		revision = istioRevsionAbsent
	}
	return revision
}

func (p *IstioDrivenPod) injectAnnotation() bool {
	NeedInject := true
	if inject, ok := p.Annotations["sidecar.istio.io/inject"]; ok {
		if inject == "false" {
			NeedInject = false
		}
	}
	return NeedInject
}

func (p *IstioDrivenPod) injectLabel() bool {
	NeedInject := false
	if inject, ok := p.Labels["sidecar.istio.io/inject"]; ok {
		if inject == "true" {
			NeedInject = true
		}
	}
	return NeedInject
}

func (p *IstioDrivenPod) getIstioSpecificRevision() string {
	if specificPodRevision, ok := p.Labels["istio.io/rev"]; ok {
		return specificPodRevision
	}
	return ""
}

func (p *IstioDrivenPod) getIstioFullVersion() string {
	if istioVersion, ok := p.Annotations["istio.deckhouse.io/version"]; ok {
		return istioVersion
	} else if _, ok := p.Annotations["sidecar.istio.io/status"]; ok {
		return istioVersionUnknown
	}
	return istioVersionAbsent
}

type Owner struct {
	Name string
	Kind string
}

type IstioPodInfo struct {
	DesiredFullVersion string
	Owner              Owner
}

type IstioPodResult struct {
	Name             string
	Namespace        string
	FullVersion      string // istio dataplane version (i.e. "1.15.6")
	Revision         string // istio dataplane revision (i.e. "v1x15")
	SpecificRevision string // istio.io/rev: vXxYZ label if it is
	InjectAnnotation bool   // sidecar.istio.io/inject annotation if it is
	InjectLabel      bool   // sidecar.istio.io/inject label if it is
	Owner            Owner
}

func applyIstioPodFilter(obj *unstructured.Unstructured) (go_hook.FilterResult, error) {
	pod := v1.Pod{}
	err := sdk.FromUnstructured(obj, &pod)
	if err != nil {
		return nil, fmt.Errorf("cannot convert pod object to pod: %v", err)
	}
	istioPod := IstioDrivenPod(pod)

	result := IstioPodResult{
		Name:             istioPod.Name,
		Namespace:        istioPod.Namespace,
		FullVersion:      istioPod.getIstioFullVersion(),
		Revision:         istioPod.getIstioCurrentRevision(),
		SpecificRevision: istioPod.getIstioSpecificRevision(),
		InjectAnnotation: istioPod.injectAnnotation(),
		InjectLabel:      istioPod.injectLabel(),
	}

	if len(pod.OwnerReferences) == 1 {
		result.Owner.Name = pod.OwnerReferences[0].Name
		result.Owner.Kind = pod.OwnerReferences[0].Kind
	}
	return result, nil
}

type IstioResourceResult struct {
	Name                   string
	Kind                   string
	Namespace              string
	AvailableForUpgrade    bool
	AutoUpgradeLabelExists bool
	Owner                  Owner
}

func applyIstioDeploymentFilter(obj *unstructured.Unstructured) (go_hook.FilterResult, error) {
	deploy := appsv1.Deployment{}
	err := sdk.FromUnstructured(obj, &deploy)
	if err != nil {
		return nil, fmt.Errorf("cannot convert deployment object to deployment: %v", err)
	}

	result := IstioResourceResult{
		Name:                deploy.Name,
		Kind:                deploy.Kind,
		Namespace:           deploy.Namespace,
		AvailableForUpgrade: deploy.Status.UnavailableReplicas == 0,
	}

	if _, ok := deploy.Labels[autoUpgradeLabelName]; ok {
		result.AutoUpgradeLabelExists = deploy.Labels[autoUpgradeLabelName] == "true"
	}

	return result, nil
}

func applyIstioStatefulSetFilter(obj *unstructured.Unstructured) (go_hook.FilterResult, error) {
	sts := appsv1.StatefulSet{}
	err := sdk.FromUnstructured(obj, &sts)
	if err != nil {
		return nil, fmt.Errorf("cannot convert statefulset object to statefulset: %v", err)
	}

	result := IstioResourceResult{
		Name:                sts.Name,
		Kind:                sts.Kind,
		Namespace:           sts.Namespace,
		AvailableForUpgrade: sts.Status.Replicas == sts.Status.ReadyReplicas,
	}

	if _, ok := sts.Labels[autoUpgradeLabelName]; ok {
		result.AutoUpgradeLabelExists = sts.Labels[autoUpgradeLabelName] == "true"
	}

	return result, nil
}

func applyIstioDaemonSetFilter(obj *unstructured.Unstructured) (go_hook.FilterResult, error) {
	ds := appsv1.DaemonSet{}
	err := sdk.FromUnstructured(obj, &ds)
	if err != nil {
		return nil, fmt.Errorf("cannot convert deployment object to deployment: %v", err)
	}

	result := IstioResourceResult{
		Name:                ds.Name,
		Kind:                ds.Kind,
		Namespace:           ds.Namespace,
		AvailableForUpgrade: ds.Status.NumberUnavailable == 0,
	}

	if _, ok := ds.Labels[autoUpgradeLabelName]; ok {
		result.AutoUpgradeLabelExists = ds.Labels[autoUpgradeLabelName] == "true"
	}

	return result, nil
}

func applyIstioReplicaSetFilter(obj *unstructured.Unstructured) (go_hook.FilterResult, error) {
	rs := appsv1.ReplicaSet{}
	err := sdk.FromUnstructured(obj, &rs)
	if err != nil {
		return nil, fmt.Errorf("cannot convert replicaset object to replicaset: %v", err)
	}

	result := IstioResourceResult{
		Name:                rs.Name,
		Namespace:           rs.Namespace,
		AvailableForUpgrade: rs.Status.Replicas == rs.Status.ReadyReplicas,
	}

	if len(rs.OwnerReferences) == 1 {
		result.Owner.Name = rs.OwnerReferences[0].Name
		result.Owner.Kind = rs.OwnerReferences[0].Kind
	}

	return result, nil
}

func dataplaneController(input *go_hook.HookInput) error {
	if !input.Values.Get("istio.internal.globalVersion").Exists() {
		return nil
	}

	versionMap := istio_versions.VersionMapJSONToVersionMap(input.Values.Get("istio.internal.versionMap").String())

	globalRevision := versionMap[input.Values.Get("istio.internal.globalVersion").String()].Revision

	input.MetricsCollector.Expire(metadataExporterMetricsGroup)

	istioNamespaceMap := make(map[string]IstioNamespace)
	for _, ns := range append(input.Snapshots["namespaces_definite_revision"], input.Snapshots["namespaces_global_revision"]...) {
		nsInfo := ns.(IstioNamespaceResult)
		if nsInfo.Revision == "global" {
			istioNamespaceMap[nsInfo.Name] = IstioNamespace{Revision: globalRevision, AutoUpgradeLabelExists: nsInfo.AutoUpgradeLabelExists}
		} else {
			istioNamespaceMap[nsInfo.Name] = IstioNamespace{Revision: nsInfo.Revision, AutoUpgradeLabelExists: nsInfo.AutoUpgradeLabelExists}
		}
	}

	// istioPodsMapToUpgrade[namespace][pod name]<pod info>
	istioPodsMapToUpgrade := make(map[string]map[string]IstioPodInfo)

	// istioResources[kind][namespace][name]desiredFullVersion
	istioResources := make(map[string]map[string]map[string]string)

	// istioReplicaSets[namespace][replicaset-name]owner
	istioReplicaSets := make(map[string]map[string]Owner)

	resources := make([]go_hook.FilterResult, len(input.Snapshots["deployment"])+len(input.Snapshots["statefulset"])+len(input.Snapshots["daemonset"]))
	resources = append(resources, input.Snapshots["deployment"]...)
	resources = append(resources, input.Snapshots["statefulset"]...)
	resources = append(resources, input.Snapshots["daemonset"]...)

	for _, resRaw := range resources {
		res := resRaw.(IstioResourceResult)

		// check if AutoUpgrade Label Exists on namespace
		var NamespaceAutoUpgradeLabelExists bool
		if deployNS, ok := istioNamespaceMap[res.Namespace]; ok {
			NamespaceAutoUpgradeLabelExists = deployNS.AutoUpgradeLabelExists
		}

		// if an istio.deckhouse.io/auto-upgrade Label exists in the namespace or in the deployment
		// and the resource is available for upgrade -> add to deployments map
		if (NamespaceAutoUpgradeLabelExists || res.AutoUpgradeLabelExists) && res.AvailableForUpgrade {
			if _, ok := istioResources[res.Kind]; !ok {
				istioResources[res.Kind] = make(map[string]map[string]string)
			}
			if _, ok := istioResources[res.Namespace]; !ok {
				istioResources[res.Kind][res.Namespace] = make(map[string]string)
			}
			istioResources[res.Kind][res.Namespace][res.Name] = ""
		}
	}

	// create a map of the replica sets depending on the deployments
	for _, rs := range input.Snapshots["replicaset"] {
		rsInfo := rs.(IstioResourceResult)
		if rsInfo.Owner.Kind == "Deployment" {
			if _, ok := istioResources["Deployment"][rsInfo.Namespace][rsInfo.Owner.Name]; ok {
				if _, ok := istioReplicaSets[rsInfo.Namespace]; !ok {
					istioReplicaSets[rsInfo.Namespace] = make(map[string]Owner)
				}
				istioReplicaSets[rsInfo.Namespace][rsInfo.Name] = Owner{
					Kind: rsInfo.Owner.Kind,
					Name: rsInfo.Owner.Name,
				}
			}
		}
	}

	for _, pod := range input.Snapshots["istio_pod"] {
		istioPodInfo := pod.(IstioPodResult)

		// sidecar.istio.io/inject=false annotation set -> ignore
		if !istioPodInfo.InjectAnnotation {
			continue
		}

		desiredRevision := istioRevsionAbsent

		// if label sidecar.istio.io/inject=true -> use global revision
		if istioPodInfo.InjectLabel {
			desiredRevision = globalRevision
		}
		// override if injection labels on namespace
		if desiredRevisionNS, ok := istioNamespaceMap[istioPodInfo.Namespace]; ok {
			desiredRevision = desiredRevisionNS.Revision
		}
		// override if label istio.io/rev with specific revision exists
		if istioPodInfo.SpecificRevision != "" {
			desiredRevision = istioPodInfo.SpecificRevision
		}

		// we don't need metrics for pod without desired revision and without istio sidecar
		if desiredRevision == istioRevsionAbsent && istioPodInfo.Revision == istioRevsionAbsent {
			continue
		}

		desiredFullVersion := versionMap.GetFullVersionByRevision(desiredRevision)
		if desiredFullVersion == "" {
			desiredFullVersion = istioVersionUnknown
		}
		desiredVersion := versionMap.GetVersionByRevision(desiredRevision)
		if desiredVersion == "" {
			desiredVersion = istioVersionUnknown
		}
		var podVersion string
		if istioPodInfo.FullVersion == istioVersionAbsent {
			podVersion = istioVersionAbsent
		} else {
			podVersion = versionMap.GetVersionByFullVersion(istioPodInfo.FullVersion)
			if podVersion == "" {
				podVersion = istioVersionUnknown
			}
		}

		labels := map[string]string{
			"namespace":            istioPodInfo.Namespace,
			"dataplane_pod":        istioPodInfo.Name,
			"desired_revision":     desiredRevision,
			"revision":             istioPodInfo.Revision,
			"full_version":         istioPodInfo.FullVersion,
			"desired_full_version": desiredFullVersion,
			"version":              podVersion,
			"desired_version":      desiredVersion,
		}

		input.MetricsCollector.Set(istioPodMetadataMetricName, 1, labels, metrics.WithGroup(metadataExporterMetricsGroup))

		// create a map of pods needed for upgrading
		if istioPodInfo.FullVersion != desiredFullVersion {
			if _, ok := istioPodsMapToUpgrade[istioPodInfo.Namespace]; !ok {
				istioPodsMapToUpgrade[istioPodInfo.Namespace] = make(map[string]IstioPodInfo)
			}
			istioPodsMapToUpgrade[istioPodInfo.Namespace][istioPodInfo.Name] = IstioPodInfo{
				Owner:              istioPodInfo.Owner,
				DesiredFullVersion: desiredFullVersion,
			}
		}
	}

	// search for resources that require a sidecar update
	for ns, pods := range istioPodsMapToUpgrade {
		for _, pod := range pods {
			switch pod.Owner.Kind {
			case "ReplicaSet":
				if rs, ok := istioReplicaSets[ns][pod.Owner.Name]; ok {
					if _, ok := istioResources[rs.Kind][ns][rs.Name]; ok {
						istioResources[rs.Kind][ns][rs.Name] = pod.DesiredFullVersion
					}
				}
			case "StatefulSet", "DaemonSet":
				if _, ok := istioResources[pod.Owner.Kind][ns][pod.Owner.Name]; ok {
					istioResources[pod.Owner.Kind][ns][pod.Owner.Name] = pod.DesiredFullVersion
				}
			}
		}
	}

	// update all resources that require a sidecar update
	for kind, namespaces := range istioResources {
		for namespace, resources := range namespaces {
			for name, desiredFullVersion := range resources {
				if desiredFullVersion != "" {
					patch := fmt.Sprintf(patchTemplate, desiredFullVersion)
					input.PatchCollector.MergePatch(patch, "apps/v1", kind, namespace, name)
				}
			}
		}
	}

	return nil
}
