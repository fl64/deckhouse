/*
Copyright 2022 Flant JSC
Licensed under the Deckhouse Platform Enterprise Edition (EE) license. See https://github.com/deckhouse/deckhouse/blob/main/ee/LICENSE
*/

package hooks

import (
	"strings"

	"github.com/deckhouse/deckhouse/ee/modules/110-istio/hooks/internal"
	. "github.com/deckhouse/deckhouse/testing/hooks"
	"github.com/flant/shell-operator/pkg/metric_storage/operation"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"k8s.io/utils/pointer"
)

const (
	nsName     = "ns"
	deployName = "deploy"
	stsName    = "sts"
	dsName     = "ds"
	rsName     = "rs"
	podName    = "pod"
)

type nsParams struct {
	GlobalRevision   bool
	AutoUpgrade      bool
	DefiniteRevision string
	Name             string
}

const nsTemplate = `apiVersion: v1
kind: Namespace
metadata:
  name: {{ .Name }}
  {{- if or .GlobalRevision .DefiniteRevision }}
  labels:
    {{ if .AutoUpgrade }}istio.deckhouse.io/auto-upgrade: "true"{{ end }}
    {{ if .GlobalRevision }}istio-injection: enabled{{ end }}
    {{ if .DefiniteRevision }}istio.io/rev: "{{ .DefiniteRevision }}"{{ end }}
 {{ end }}
`

func generateIstioNsYAML(ns nsParams) string {
	ns.Name = nsName
	return internal.TemplateToYAML(nsTemplate, ns)
}

type deployParams struct {
	Name                string
	Namespace           string
	Replicas            int32
	UnavailableReplicas int32
	AutoUpgrade         bool
}

const deployTemplate = `apiVersion: apps/v1
kind: Deployment
metadata:
  namespace: {{ .Namespace }}
  name: {{ .Name }}
  labels:
    app: test
    {{ if .AutoUpgrade }}istio.deckhouse.io/auto-upgrade: "true"{{ end }}
spec:
  replicas: {{ .Replicas }}
  selector:
    matchLabels:
      app: test
  template: {}
status:
  replicas: {{ .Replicas }}
  unavailableReplicas: {{ .UnavailableReplicas }}
`

func generateIstioDeploymentYAML(deploy deployParams) string {
	deploy.Namespace = nsName
	deploy.Name = deployName
	return internal.TemplateToYAML(deployTemplate, deploy)
}

type stsParams struct {
	Name          string
	Namespace     string
	Replicas      int32
	ReadyReplicas int32
	AutoUpgrade   bool
}

const stsTemplate = `apiVersion: apps/v1
kind: StatefulSet
metadata:
  namespace: {{ .Namespace }}
  name: {{ .Name }}
  labels:
    app: test
    {{ if .AutoUpgrade }}istio.deckhouse.io/auto-upgrade: "true"{{ end }}
spec:
  podManagementPolicy: OrderedReady
  replicas: {{ .Replicas }}
  selector:
    matchLabels:
      app: test
  serviceName: test
  template: {}
status:
  readyReplicas: {{ .ReadyReplicas }}
  replicas: {{ .Replicas }}
`

func generateIstioStatefulSetYAML(sts stsParams) string {
	sts.Namespace = nsName
	sts.Name = stsName
	return internal.TemplateToYAML(stsTemplate, sts)
}

type dsParams struct {
	Name              string
	Namespace         string
	NumberUnavailable int32
	AutoUpgrade       bool
}

const dsTemplate = `apiVersion: apps/v1
kind: DaemonSet
metadata:
  labels:
    app: test
    {{ if .AutoUpgrade }}istio.deckhouse.io/auto-upgrade: "true"{{ end }}
  name: {{ .Name }}
  namespace: {{ .Namespace }}
spec:
  selector:
    matchLabels:
      app: test
  template: {}
status:
  numberUnavailable: {{ .NumberUnavailable }}
`

func generateIstioDaemonSetYAML(ds dsParams) string {
	ds.Namespace = nsName
	ds.Name = dsName
	return internal.TemplateToYAML(dsTemplate, ds)
}

type rsParams struct {
	Name      string
	Namespace string
	Replicas  int32
	OwnerName string
	OwnerKind string
}

const rsTemplate = `apiVersion: apps/v1
kind: ReplicaSet
metadata:
  namespace: {{ .Namespace }}
  name: {{ .Name }}
  labels:
    app: test
    pod-template-hash: rs
  {{- if and .Name .OwnerKind }}
  ownerReferences:
    - kind: {{ .OwnerKind }}
      name: {{ .OwnerName }}
  {{- end }}
spec:
  replicas: {{ .Replicas }}
  selector:
    matchLabels:
      app: test
      pod-template-hash: rs
  template: {}
status:
  replicas: {{ .Replicas }}
`

func generateIstioReplicaSetYAML(rs rsParams) string {
	rs.Namespace = nsName
	rs.Name = rsName
	return internal.TemplateToYAML(rsTemplate, rs)
}

type podParams struct {
	InjectionLabel             bool
	InjectionLabelValue        bool
	DisableInjectionAnnotation bool
	DefiniteRevision           string
	CurrentRevision            string
	FullVersion                string
	Name                       string
	Namespace                  string
	OwnerName                  string
	OwnerKind                  string
}

const podTemplate = `apiVersion: v1
kind: Pod
metadata:
  name: {{ .Name }}
  namespace: {{ .Namespace }}
  labels:
    app: test
    pod-template-hash: rs
    service.istio.io/canonical-name: {{ .Name }}
    {{- if .InjectionLabel }}
    sidecar.istio.io/inject: "{{ .InjectionLabelValue }}"
    {{- end }}
    {{- if .DefiniteRevision }}
    istio.io/rev: {{ .DefiniteRevision }}
    {{- end }}
  annotations:
    some-annotation: some-value
    {{- if .FullVersion }}
    istio.deckhouse.io/version: '{{ .FullVersion }}'
    {{- end }}
    {{- if .CurrentRevision }}
    sidecar.istio.io/status: '{"a":"b", "revision":"{{ .CurrentRevision }}" }'
    {{- end }}
    {{- if .DisableInjectionAnnotation }}
    sidecar.istio.io/inject: "false"
    {{- end }}
  {{- if and .Name .OwnerKind }}
  ownerReferences:
    - kind: {{ .OwnerKind }}
      name: {{ .OwnerName }}
  {{- end }}
spec: {}
`

func generateIstioPodYAML(pod podParams) string {
	pod.Namespace = nsName
	if pod.Name == "" {
		pod.Name = podName
	}
	return internal.TemplateToYAML(podTemplate, pod)
}

type wantedMetric struct {
	Revision           string
	DesiredRevision    string
	Version            string
	DesiredVersion     string
	FullVersion        string
	DesiredFullVersion string
}

var hookInitValues = `
{  "istio":
  { "internal":
    { "versionMap":
      {
         "1.15": { revision: "v1x15", fullVersion: "1.15.15" },
         "1.42": { revision: "v1x42", fullVersion: "1.42.42" },
         "1.71": { revision: "v1x71", fullVersion: "1.71.71" },
         "1.77": { revision: "v1x77", fullVersion: "1.77.77" },
         "1.155": { revision: "v1x155", fullVersion: "1.155.155" }
      }
    }
  }
}
`

var _ = Describe("Istio hooks :: dataplane_controller :: metrics ::", func() {

	f := HookExecutionConfigInit(hookInitValues, "")
	Context("Empty cluster and minimal settings", func() {
		BeforeEach(func() {
			f.RunHook()
		})

		It("Hook must execute successfully", func() {
			Expect(f).To(ExecuteSuccessfully())
			Expect(string(f.LogrusOutput.Contents())).To(HaveLen(0))

			m := f.MetricsCollector.CollectedMetrics()
			Expect(m).To(HaveLen(0))

		})
	})

	DescribeTable("There are different desired and actual revisions", func(objectsYAMLs []string, want *wantedMetric) {
		f.ValuesSet("istio.internal.globalVersion", "1.42")
		yamlState := strings.Join(objectsYAMLs, "\n---\n")
		f.BindingContexts.Set(f.KubeStateSet(yamlState))

		f.RunHook()
		Expect(f).To(ExecuteSuccessfully())
		Expect(string(f.LogrusOutput.Contents())).To(HaveLen(0))
		m := f.MetricsCollector.CollectedMetrics()

		// the first action should always be "expire"
		Expect(m[0]).To(BeEquivalentTo(operation.MetricOperation{
			Group:  metadataExporterMetricsGroup,
			Action: "expire",
		}))

		// there are no istio pods or ignored pods in the cluster, hense no metrics
		if yamlState == "" || want == nil {
			Expect(m).To(HaveLen(1))
			return
		}
		Expect(m).To(HaveLen(2))
		Expect(m[1]).To(BeEquivalentTo(operation.MetricOperation{
			Name:   istioPodMetadataMetricName,
			Group:  metadataExporterMetricsGroup,
			Action: "set",
			Value:  pointer.Float64Ptr(1.0),
			Labels: map[string]string{
				"namespace":            nsName,
				"dataplane_pod":        podName,
				"desired_revision":     want.DesiredRevision,
				"revision":             want.Revision,
				"version":              want.Version,
				"desired_version":      want.DesiredVersion,
				"full_version":         want.FullVersion,
				"desired_full_version": want.DesiredFullVersion,
			},
		}))
	},

		// Checks for normal behavior, everything with revision is ok!
		Entry("Empty cluster", []string{}, nil),
		Entry("NS with global revision, Pod to ignore with inject=false label",
			[]string{
				generateIstioNsYAML(nsParams{
					GlobalRevision: true,
				}),
				generateIstioPodYAML(podParams{
					InjectionLabel:      true,
					InjectionLabelValue: false,
				}),
			}, nil),
		Entry("NS with definite revision, but revision is absent in revisionFullVersionMap",
			[]string{
				generateIstioNsYAML(nsParams{
					DefiniteRevision: "v1x00",
				}),
				generateIstioPodYAML(podParams{
					InjectionLabel:      true,
					InjectionLabelValue: true,
					CurrentRevision:     "v1x00",
					FullVersion:         "", // annotation is absent
				}),
			}, &wantedMetric{
				Revision:           "v1x00",
				DesiredRevision:    "v1x00",
				Version:            "unknown",
				DesiredVersion:     "unknown",
				FullVersion:        "unknown",
				DesiredFullVersion: "unknown",
			}),
		Entry("NS without any revisions, pod with inject=true label",
			[]string{
				generateIstioNsYAML(nsParams{
					GlobalRevision: false,
				}),
				generateIstioPodYAML(podParams{
					InjectionLabel:      true,
					InjectionLabelValue: true,
					CurrentRevision:     "v1x42",
					FullVersion:         "1.42.42",
				}),
			}, &wantedMetric{
				Revision:           "v1x42",
				DesiredRevision:    "v1x42",
				Version:            "1.42",
				DesiredVersion:     "1.42",
				FullVersion:        "1.42.42",
				DesiredFullVersion: "1.42.42",
			}),
		Entry("NS with global revision, pod with inject=true label",
			[]string{
				generateIstioNsYAML(nsParams{
					GlobalRevision: true,
				}),
				generateIstioPodYAML(podParams{
					InjectionLabel:      true,
					InjectionLabelValue: true,
					CurrentRevision:     "v1x42",
					FullVersion:         "1.42.42",
				}),
			}, &wantedMetric{
				Revision:           "v1x42",
				DesiredRevision:    "v1x42",
				Version:            "1.42",
				DesiredVersion:     "1.42",
				FullVersion:        "1.42.42",
				DesiredFullVersion: "1.42.42",
			}),
		Entry("NS with definite revision, pod with inject=true label",
			[]string{
				generateIstioNsYAML(nsParams{
					DefiniteRevision: "v1x15",
				}),
				generateIstioPodYAML(podParams{
					InjectionLabel:      true,
					InjectionLabelValue: true,
					CurrentRevision:     "v1x15",
					FullVersion:         "1.15.15",
				}),
			}, &wantedMetric{
				Revision:           "v1x15",
				DesiredRevision:    "v1x15",
				Version:            "1.15",
				DesiredVersion:     "1.15",
				FullVersion:        "1.15.15",
				DesiredFullVersion: "1.15.15",
			}),
		Entry("NS without any revisions, pod with istio.io/rev label",
			[]string{
				generateIstioNsYAML(nsParams{
					GlobalRevision: false,
				}),
				generateIstioPodYAML(podParams{
					DefiniteRevision: "v1x15",
					CurrentRevision:  "v1x15",
					FullVersion:      "1.15.15",
				}),
			}, &wantedMetric{
				Revision:           "v1x15",
				DesiredRevision:    "v1x15",
				Version:            "1.15",
				DesiredVersion:     "1.15",
				FullVersion:        "1.15.15",
				DesiredFullVersion: "1.15.15",
			}),
		Entry("NS with global revision, pod with istio.io/rev label",
			[]string{
				generateIstioNsYAML(nsParams{
					GlobalRevision: true,
				}),
				generateIstioPodYAML(podParams{
					DefiniteRevision: "v1x15",
					CurrentRevision:  "v1x15",
					FullVersion:      "1.15.15",
				}),
			}, &wantedMetric{
				Revision:           "v1x15",
				DesiredRevision:    "v1x15",
				Version:            "1.15",
				DesiredVersion:     "1.15",
				FullVersion:        "1.15.15",
				DesiredFullVersion: "1.15.15",
			}),
		Entry("NS with definite revision, pod with inject=true label",
			[]string{
				generateIstioNsYAML(nsParams{
					DefiniteRevision: "v1x15",
				}),
				generateIstioPodYAML(podParams{
					DefiniteRevision: "v1x155",
					CurrentRevision:  "v1x155",
					FullVersion:      "1.155.155",
				}),
			}, &wantedMetric{
				Revision:           "v1x155",
				DesiredRevision:    "v1x155",
				Version:            "1.155",
				DesiredVersion:     "1.155",
				FullVersion:        "1.155.155",
				DesiredFullVersion: "1.155.155",
			}),
		Entry("NS with global revision, Pod to ignore with inject=false annotation",
			[]string{
				generateIstioNsYAML(nsParams{
					GlobalRevision: true,
				}),
				generateIstioPodYAML(podParams{
					DisableInjectionAnnotation: true,
				}),
			}, nil),
		Entry("NS with definite revision, Pod to ignore with inject=false annotation",
			[]string{
				generateIstioNsYAML(nsParams{
					DefiniteRevision: "v1x15",
				}),
				generateIstioPodYAML(podParams{
					DisableInjectionAnnotation: true,
				}),
			}, nil),
		Entry("NS with global revision, Pod revision is actual",
			[]string{
				generateIstioNsYAML(nsParams{
					GlobalRevision: true,
				}),
				generateIstioPodYAML(podParams{
					CurrentRevision: "v1x42",
					FullVersion:     "1.42.42",
				}),
			}, &wantedMetric{
				Revision:           "v1x42",
				DesiredRevision:    "v1x42",
				Version:            "1.42",
				DesiredVersion:     "1.42",
				FullVersion:        "1.42.42",
				DesiredFullVersion: "1.42.42",
			}),
		Entry("Namespace with definite revision, pod revision is actual",
			[]string{
				generateIstioNsYAML(nsParams{
					DefiniteRevision: "v1x15",
				}),
				generateIstioPodYAML(podParams{
					CurrentRevision: "v1x15",
					FullVersion:     "1.15.15",
				}),
			}, &wantedMetric{
				Revision:           "v1x15",
				DesiredRevision:    "v1x15",
				Version:            "1.15",
				DesiredVersion:     "1.15",
				FullVersion:        "1.15.15",
				DesiredFullVersion: "1.15.15",
			}),

		// Checks for revision inconsistencies
		Entry("NS global revision, pod revision is not actual",
			[]string{
				generateIstioNsYAML(nsParams{
					GlobalRevision: true,
				}),
				generateIstioPodYAML(podParams{
					CurrentRevision: "v1x77",
					FullVersion:     "1.77.77",
				}),
			}, &wantedMetric{
				Revision:           "v1x77",
				DesiredRevision:    "v1x42",
				Version:            "1.77",
				DesiredVersion:     "1.42",
				FullVersion:        "1.77.77",
				DesiredFullVersion: "1.42.42",
			}),
		Entry("NS global revision, pod revision is absent (no sidecar)",
			[]string{
				generateIstioNsYAML(nsParams{
					GlobalRevision: true,
				}),
				generateIstioPodYAML(podParams{}),
			}, &wantedMetric{
				Revision:           "absent",
				DesiredRevision:    "v1x42",
				Version:            "absent",
				DesiredVersion:     "1.42",
				FullVersion:        "absent",
				DesiredFullVersion: "1.42.42",
			}),
		Entry("Namespace with definite revision, pod revision is not actual",
			[]string{
				generateIstioNsYAML(nsParams{
					DefiniteRevision: "v1x15",
				}),
				generateIstioPodYAML(podParams{
					CurrentRevision: "v1x77",
					FullVersion:     "1.77.77",
				}),
			}, &wantedMetric{
				Revision:           "v1x77",
				DesiredRevision:    "v1x15",
				Version:            "1.77",
				DesiredVersion:     "1.15",
				FullVersion:        "1.77.77",
				DesiredFullVersion: "1.15.15",
			}),
		Entry("Namespace with definite revision, pod revision is absent (no sidecar)",
			[]string{
				generateIstioNsYAML(nsParams{
					DefiniteRevision: "v1x15",
				}),
				generateIstioPodYAML(podParams{}),
			}, &wantedMetric{
				Revision:           "absent",
				DesiredRevision:    "v1x15",
				Version:            "absent",
				DesiredVersion:     "1.15",
				FullVersion:        "absent",
				DesiredFullVersion: "1.15.15",
			}),
		Entry("Namespace with definite revision and pod with definite revision is actual",
			[]string{
				generateIstioNsYAML(nsParams{
					DefiniteRevision: "v1x15",
				}),
				generateIstioPodYAML(podParams{
					DefiniteRevision: "v1x77",
					CurrentRevision:  "v1x77",
					FullVersion:      "1.77.77",
				}),
			}, &wantedMetric{
				Revision:           "v1x77",
				DesiredRevision:    "v1x77",
				Version:            "1.77",
				DesiredVersion:     "1.77",
				FullVersion:        "1.77.77",
				DesiredFullVersion: "1.77.77",
			}),
		Entry("Namespace with definite revision and pod with definite revision is not actual",
			[]string{
				generateIstioNsYAML(nsParams{
					DefiniteRevision: "v1x15",
				}),
				generateIstioPodYAML(podParams{
					DefiniteRevision: "v1x77",
					CurrentRevision:  "v1x71",
					FullVersion:      "1.71.71",
				}),
			}, &wantedMetric{
				Revision:           "v1x71",
				DesiredRevision:    "v1x77",
				Version:            "1.71",
				DesiredVersion:     "1.77",
				FullVersion:        "1.71.71",
				DesiredFullVersion: "1.77.77",
			}),
		Entry("Namespace without labels and pod with definite revision",
			[]string{
				generateIstioNsYAML(nsParams{}),
				generateIstioPodYAML(podParams{
					DefiniteRevision: "v1x77",
					CurrentRevision:  "v1x77",
					FullVersion:      "1.77.77",
				}),
			}, &wantedMetric{
				Revision:           "v1x77",
				DesiredRevision:    "v1x77",
				Version:            "1.77",
				DesiredVersion:     "1.77",
				FullVersion:        "1.77.77",
				DesiredFullVersion: "1.77.77",
			}),
		Entry("Namespace without labels and pod with definite revision but sidecar absent",
			[]string{
				generateIstioNsYAML(nsParams{}),
				generateIstioPodYAML(podParams{
					DefiniteRevision: "v1x77",
				}),
			}, &wantedMetric{
				Revision:           "absent",
				DesiredRevision:    "v1x77",
				Version:            "absent",
				DesiredVersion:     "1.77",
				FullVersion:        "absent",
				DesiredFullVersion: "1.77.77",
			}),
		Entry("Pod orphan",
			[]string{
				generateIstioNsYAML(nsParams{}),
				generateIstioPodYAML(podParams{
					CurrentRevision: "v1x77",
					FullVersion:     "1.77.77",
				}),
			}, &wantedMetric{
				Revision:           "v1x77",
				DesiredRevision:    "absent",
				Version:            "1.77",
				DesiredVersion:     "unknown",
				FullVersion:        "1.77.77",
				DesiredFullVersion: "unknown",
			}),
		Entry("Pod without current and desired revisions",
			[]string{
				generateIstioNsYAML(nsParams{}),
				generateIstioPodYAML(podParams{}),
			}, nil),
	)
})

var _ = Describe("Istio hooks :: dataplane_controller :: dataplane upgrade ::", func() {

	f := HookExecutionConfigInit(hookInitValues, "")

	istioNsYAML := generateIstioNsYAML(nsParams{
		GlobalRevision: true,
	})

	istioNsWithAutoupgradeYAML := generateIstioNsYAML(nsParams{
		AutoUpgrade:    true,
		GlobalRevision: true,
	})

	Context("Test Deployment", func() {

		istioDeployYAML := generateIstioDeploymentYAML(deployParams{
			Replicas:            2,
			UnavailableReplicas: 0,
			AutoUpgrade:         false,
		})

		istioDeployWithAutoupgradeYAML := generateIstioDeploymentYAML(deployParams{
			Replicas:            2,
			UnavailableReplicas: 0,
			AutoUpgrade:         true,
		})

		istioDeployWithUnavailableYAML := generateIstioDeploymentYAML(deployParams{
			Replicas:            2,
			UnavailableReplicas: 1,
			AutoUpgrade:         true,
		})

		istioRsYAML := generateIstioReplicaSetYAML(rsParams{
			OwnerKind: "Deployment",
			OwnerName: deployName,
			Replicas:  2,
		})

		istioRSPod0 := generateIstioPodYAML(podParams{
			Name:            "pod-0",
			CurrentRevision: "v1x42",
			FullVersion:     "1.42.00",
			OwnerName:       rsName,
			OwnerKind:       "ReplicaSet",
		})

		istioRSPod1 := generateIstioPodYAML(podParams{
			Name:            "pod-1",
			CurrentRevision: "v1x42",
			FullVersion:     "1.42.42",
			OwnerName:       rsName,
			OwnerKind:       "ReplicaSet",
		})

		istioRSPod2 := generateIstioPodYAML(podParams{
			Name:            "pod-2",
			CurrentRevision: "v1x42",
			FullVersion:     "1.42.42",
			OwnerName:       rsName,
			OwnerKind:       "ReplicaSet",
		})

		Context("Deployment with auto-upgrade label has a pod with old istio version", func() {
			BeforeEach(func() {
				f.ValuesSet("istio.internal.globalVersion", "1.42")

				clusterState := strings.Join([]string{istioNsYAML, istioDeployWithAutoupgradeYAML, istioRsYAML, istioRSPod0, istioRSPod1}, "---\n")
				f.BindingContexts.Set(f.KubeStateSet(clusterState))

				f.RunHook()
			})

			It("Hook must execute successfully", func() {
				Expect(f).To(ExecuteSuccessfully())
				Expect(string(f.LogrusOutput.Contents())).To(HaveLen(0))

				m := f.MetricsCollector.CollectedMetrics()
				Expect(m).To(HaveLen(3))

				d := f.KubernetesResource("Deployment", nsName, deployName)
				Expect(d.Exists()).Should(BeTrue())
				Expect(f.KubernetesResource("ReplicaSet", nsName, rsName).Exists()).Should(BeTrue())
				Expect(d.Field("spec.template.metadata.annotations").String()).To(MatchJSON(`{"istio.deckhouse.io/version": "1.42.42"}`))
			})
		})

		Context("Name space with auto-upgrade label. Deployment has a pod with old istio version", func() {
			BeforeEach(func() {
				f.ValuesSet("istio.internal.globalVersion", "1.42")

				clusterState := strings.Join([]string{istioNsWithAutoupgradeYAML, istioDeployYAML, istioRsYAML, istioRSPod0, istioRSPod1}, "---\n")
				f.BindingContexts.Set(f.KubeStateSet(clusterState))

				f.RunHook()
			})

			It("Hook must execute successfully", func() {
				Expect(f).To(ExecuteSuccessfully())
				Expect(string(f.LogrusOutput.Contents())).To(HaveLen(0))

				m := f.MetricsCollector.CollectedMetrics()
				Expect(m).To(HaveLen(3))

				d := f.KubernetesResource("Deployment", nsName, deployName)
				Expect(d.Exists()).Should(BeTrue())
				Expect(f.KubernetesResource("ReplicaSet", nsName, rsName).Exists()).Should(BeTrue())
				Expect(d.Field("spec.template.metadata.annotations").String()).To(MatchJSON(`{"istio.deckhouse.io/version": "1.42.42"}`))
			})
		})

		Context("Name space with auto-upgrade label. All deployment pods have actial istio version", func() {
			BeforeEach(func() {
				f.ValuesSet("istio.internal.globalVersion", "1.42")

				clusterState := strings.Join([]string{istioNsWithAutoupgradeYAML, istioDeployYAML, istioRsYAML, istioRSPod1, istioRSPod2}, "---\n")
				f.BindingContexts.Set(f.KubeStateSet(clusterState))

				f.RunHook()
			})

			It("Hook must execute successfully", func() {
				Expect(f).To(ExecuteSuccessfully())
				Expect(string(f.LogrusOutput.Contents())).To(HaveLen(0))

				m := f.MetricsCollector.CollectedMetrics()
				Expect(m).To(HaveLen(3))

				d := f.KubernetesResource("Deployment", nsName, deployName)
				Expect(d.Exists()).Should(BeTrue())
				Expect(f.KubernetesResource("ReplicaSet", nsName, rsName).Exists()).Should(BeTrue())
				Expect(d.Field("spec.template").String()).To(MatchJSON(`{}`))
			})
		})

		Context("Name space with auto-upgrade label. Deployment is not ready", func() {
			BeforeEach(func() {
				f.ValuesSet("istio.internal.globalVersion", "1.42")

				clusterState := strings.Join([]string{istioNsWithAutoupgradeYAML, istioDeployWithUnavailableYAML, istioRsYAML, istioRSPod0, istioRSPod1}, "---\n")
				f.BindingContexts.Set(f.KubeStateSet(clusterState))

				f.RunHook()
			})

			It("Hook must execute successfully", func() {
				Expect(f).To(ExecuteSuccessfully())
				Expect(string(f.LogrusOutput.Contents())).To(HaveLen(0))

				m := f.MetricsCollector.CollectedMetrics()
				Expect(m).To(HaveLen(3))

				d := f.KubernetesResource("Deployment", nsName, deployName)
				Expect(d.Exists()).Should(BeTrue())
				Expect(f.KubernetesResource("ReplicaSet", nsName, rsName).Exists()).Should(BeTrue())
				Expect(d.Field("spec.template").String()).To(MatchJSON(`{}`))
			})
		})
	})

	Context("Test DaemonSet", func() {

		istioDsYAML := generateIstioDaemonSetYAML(dsParams{
			NumberUnavailable: 0,
			AutoUpgrade:       false,
		})
		istioDsWithAutoupgradeYAML := generateIstioDaemonSetYAML(dsParams{
			NumberUnavailable: 0,
			AutoUpgrade:       true,
		})
		istioDsWithAutoupgradeNotReadyYAML := generateIstioDaemonSetYAML(dsParams{
			NumberUnavailable: 1,
		})

		istioDsPod0 := generateIstioPodYAML(podParams{
			Name:            "pod-0",
			CurrentRevision: "v1x42",
			FullVersion:     "1.42.00",
			OwnerName:       dsName,
			OwnerKind:       "DaemonSet",
		})

		istioDsPod1 := generateIstioPodYAML(podParams{
			Name:            "pod-1",
			CurrentRevision: "v1x42",
			FullVersion:     "1.42.42",
			OwnerName:       dsName,
			OwnerKind:       "DaemonSet",
		})

		istioDsPod2 := generateIstioPodYAML(podParams{
			Name:            "pod-2",
			CurrentRevision: "v1x42",
			FullVersion:     "1.42.42",
			OwnerName:       dsName,
			OwnerKind:       "DaemonSet",
		})

		Context("DaemonSet with auto-upgrade label has a pod with old istio version", func() {
			BeforeEach(func() {
				f.ValuesSet("istio.internal.globalVersion", "1.42")

				clusterState := strings.Join([]string{istioNsYAML, istioDsWithAutoupgradeYAML, istioDsPod0, istioDsPod1}, "---\n")
				f.BindingContexts.Set(f.KubeStateSet(clusterState))

				f.RunHook()
			})

			It("Hook must execute successfully", func() {
				Expect(f).To(ExecuteSuccessfully())
				Expect(string(f.LogrusOutput.Contents())).To(HaveLen(0))

				m := f.MetricsCollector.CollectedMetrics()
				Expect(m).To(HaveLen(3))

				d := f.KubernetesResource("DaemonSet", nsName, dsName)
				Expect(d.Exists()).Should(BeTrue())
				Expect(d.Field("spec.template.metadata.annotations").String()).To(MatchJSON(`{"istio.deckhouse.io/version": "1.42.42"}`))
			})
		})

		Context("Name space with auto-upgrade label. DaemonSet has a pod with old istio version", func() {
			BeforeEach(func() {
				f.ValuesSet("istio.internal.globalVersion", "1.42")

				clusterState := strings.Join([]string{istioNsWithAutoupgradeYAML, istioDsYAML, istioDsPod0, istioDsPod1}, "---\n")
				f.BindingContexts.Set(f.KubeStateSet(clusterState))

				f.RunHook()
			})

			It("Hook must execute successfully", func() {
				Expect(f).To(ExecuteSuccessfully())
				Expect(string(f.LogrusOutput.Contents())).To(HaveLen(0))

				m := f.MetricsCollector.CollectedMetrics()
				Expect(m).To(HaveLen(3))

				d := f.KubernetesResource("DaemonSet", nsName, dsName)
				Expect(d.Exists()).Should(BeTrue())
				Expect(d.Field("spec.template.metadata.annotations").String()).To(MatchJSON(`{"istio.deckhouse.io/version": "1.42.42"}`))
			})
		})

		Context("Name space with auto-upgrade label. All DaemonSet's pods have actial istio version", func() {
			BeforeEach(func() {
				f.ValuesSet("istio.internal.globalVersion", "1.42")

				clusterState := strings.Join([]string{istioNsWithAutoupgradeYAML, istioDsYAML, istioDsPod1, istioDsPod2}, "---\n")
				f.BindingContexts.Set(f.KubeStateSet(clusterState))

				f.RunHook()
			})

			It("Hook must execute successfully", func() {
				Expect(f).To(ExecuteSuccessfully())
				Expect(string(f.LogrusOutput.Contents())).To(HaveLen(0))

				m := f.MetricsCollector.CollectedMetrics()
				Expect(m).To(HaveLen(3))

				d := f.KubernetesResource("DaemonSet", nsName, dsName)
				Expect(d.Exists()).Should(BeTrue())
				Expect(d.Field("spec.template").String()).To(MatchJSON(`{}`))
			})
		})

		Context("Name space with auto-upgrade label. DaemonSet is not ready", func() {
			BeforeEach(func() {
				f.ValuesSet("istio.internal.globalVersion", "1.42")

				clusterState := strings.Join([]string{istioNsWithAutoupgradeYAML, istioDsWithAutoupgradeNotReadyYAML, istioDsPod0, istioDsPod1}, "---\n")
				f.BindingContexts.Set(f.KubeStateSet(clusterState))

				f.RunHook()
			})

			It("Hook must execute successfully", func() {
				Expect(f).To(ExecuteSuccessfully())
				Expect(string(f.LogrusOutput.Contents())).To(HaveLen(0))

				m := f.MetricsCollector.CollectedMetrics()
				Expect(m).To(HaveLen(3))

				d := f.KubernetesResource("DaemonSet", nsName, dsName)
				Expect(d.Exists()).Should(BeTrue())
				Expect(d.Field("spec.template").String()).To(MatchJSON(`{}`))
			})
		})
	})

	Context("Testing StatefulSet", func() {

		istioStsYAML := generateIstioStatefulSetYAML(stsParams{
			Replicas:      2,
			ReadyReplicas: 2,
			AutoUpgrade:   false,
		})
		istioStsWithAutoupgradeYAML := generateIstioStatefulSetYAML(stsParams{
			Replicas:      2,
			ReadyReplicas: 2,
			AutoUpgrade:   true,
		})
		istioStsWithAutoupgradeNotReadyYAML := generateIstioStatefulSetYAML(stsParams{
			Replicas:      2,
			ReadyReplicas: 1,
		})

		istioSTSPod0 := generateIstioPodYAML(podParams{
			Name:            "pod-0",
			CurrentRevision: "v1x42",
			FullVersion:     "1.42.00",
			OwnerName:       stsName,
			OwnerKind:       "StatefulSet",
		})

		istioSTSPod1 := generateIstioPodYAML(podParams{
			Name:            "pod-1",
			CurrentRevision: "v1x42",
			FullVersion:     "1.42.42",
			OwnerName:       stsName,
			OwnerKind:       "StatefulSet",
		})

		istioSTSPod2 := generateIstioPodYAML(podParams{
			Name:            "pod-2",
			CurrentRevision: "v1x42",
			FullVersion:     "1.42.42",
			OwnerName:       stsName,
			OwnerKind:       "StatefulSet",
		})

		Context("DaemonSet with auto-upgrade label has a pod with old istio version", func() {
			BeforeEach(func() {
				f.ValuesSet("istio.internal.globalVersion", "1.42")

				clusterState := strings.Join([]string{istioNsYAML, istioStsWithAutoupgradeYAML, istioSTSPod0, istioSTSPod1}, "---\n")
				f.BindingContexts.Set(f.KubeStateSet(clusterState))

				f.RunHook()
			})

			It("Hook must execute successfully", func() {
				Expect(f).To(ExecuteSuccessfully())
				Expect(string(f.LogrusOutput.Contents())).To(HaveLen(0))

				m := f.MetricsCollector.CollectedMetrics()
				Expect(m).To(HaveLen(3))

				d := f.KubernetesResource("StatefulSet", nsName, stsName)
				Expect(d.Exists()).Should(BeTrue())
				Expect(d.Field("spec.template.metadata.annotations").String()).To(MatchJSON(`{"istio.deckhouse.io/version": "1.42.42"}`))
			})
		})

		Context("Name space with auto-upgrade label. DaemonSet has a pod with old istio version", func() {
			BeforeEach(func() {
				f.ValuesSet("istio.internal.globalVersion", "1.42")

				clusterState := strings.Join([]string{istioNsWithAutoupgradeYAML, istioStsYAML, istioSTSPod0, istioSTSPod1}, "---\n")
				f.BindingContexts.Set(f.KubeStateSet(clusterState))

				f.RunHook()
			})

			It("Hook must execute successfully", func() {
				Expect(f).To(ExecuteSuccessfully())
				Expect(string(f.LogrusOutput.Contents())).To(HaveLen(0))

				m := f.MetricsCollector.CollectedMetrics()
				Expect(m).To(HaveLen(3))

				d := f.KubernetesResource("StatefulSet", nsName, stsName)
				Expect(d.Exists()).Should(BeTrue())
				Expect(d.Field("spec.template.metadata.annotations").String()).To(MatchJSON(`{"istio.deckhouse.io/version": "1.42.42"}`))
			})
		})

		Context("Name space with auto-upgrade label. All DaemonSet's pods have actial istio version", func() {
			BeforeEach(func() {
				f.ValuesSet("istio.internal.globalVersion", "1.42")

				clusterState := strings.Join([]string{istioNsWithAutoupgradeYAML, istioStsYAML, istioSTSPod1, istioSTSPod2}, "---\n")
				f.BindingContexts.Set(f.KubeStateSet(clusterState))

				f.RunHook()
			})

			It("Hook must execute successfully", func() {
				Expect(f).To(ExecuteSuccessfully())
				Expect(string(f.LogrusOutput.Contents())).To(HaveLen(0))

				m := f.MetricsCollector.CollectedMetrics()
				Expect(m).To(HaveLen(3))

				d := f.KubernetesResource("StatefulSet", nsName, stsName)
				Expect(d.Exists()).Should(BeTrue())
				Expect(d.Field("spec.template").String()).To(MatchJSON(`{}`))
			})
		})

		Context("Name space with auto-upgrade label. DaemonSet is not ready", func() {
			BeforeEach(func() {
				f.ValuesSet("istio.internal.globalVersion", "1.42")

				clusterState := strings.Join([]string{istioNsWithAutoupgradeYAML, istioStsWithAutoupgradeNotReadyYAML, istioSTSPod0, istioSTSPod1}, "---\n")
				f.BindingContexts.Set(f.KubeStateSet(clusterState))

				f.RunHook()
			})

			It("Hook must execute successfully", func() {
				Expect(f).To(ExecuteSuccessfully())
				Expect(string(f.LogrusOutput.Contents())).To(HaveLen(0))

				m := f.MetricsCollector.CollectedMetrics()
				Expect(m).To(HaveLen(3))

				d := f.KubernetesResource("StatefulSet", nsName, stsName)
				Expect(d.Exists()).Should(BeTrue())
				Expect(d.Field("spec.template").String()).To(MatchJSON(`{}`))
			})
		})
	})
})
