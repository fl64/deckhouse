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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8syaml "sigs.k8s.io/yaml"

	d8config "github.com/deckhouse/deckhouse/go_lib/deckhouse-config"
	module_manager "github.com/deckhouse/deckhouse/go_lib/deckhouse-config/module-manager"
	. "github.com/deckhouse/deckhouse/testing/hooks"
)

const testGlobalHooksDir = "testdata/test-sync/global-hooks"
const testModulesDir = "testdata/test-sync/modules"

var _ = Describe("Module :: deckhouse-config :: hooks :: sync ::", func() {
	f := HookExecutionConfigInit(`{"global": {"discovery": {}}}`, `{}`)
	// Emulate ensure_crd hook.
	f.RegisterCRD("deckhouse.io", "v1", "DeckhouseConfig", false)

	BeforeEach(func() {
		// Load addon-operator with 3 modules: deckhouse, cert-manager and prometheus.
		mm, initErr := module_manager.InitBasic(testGlobalHooksDir, testModulesDir)
		d8config.InitService(mm)
		Expect(initErr).ShouldNot(HaveOccurred(), "should init module manager: %s", initErr)
	})

	Context("giving absent ConfigMap", func() {
		BeforeEach(func() {
			// See openapi schemas in testdata/test-sync directory.
			validDeckhouseConfigs := `
---
apiVersion: deckhouse.io/v1
kind: DeckhouseConfig
metadata:
  name: global
spec:
  configVersion: 1
  configValues:
    paramStr: val1
---
apiVersion: deckhouse.io/v1
kind: DeckhouseConfig
metadata:
  name: deckhouse
spec:
  configVersion: 1
  configValues:
    paramStr: Debug
---
apiVersion: deckhouse.io/v1
kind: DeckhouseConfig
metadata:
  name: unknown-module-name
spec:
  configVersion: 1
  configValues:
    paramStr: val1
`

			f.BindingContexts.Set(f.KubeStateSet(validDeckhouseConfigs))
			f.RunHook()
		})

		It("should create ConfigMap from DeckhouseConfigs", func() {
			Expect(f).To(ExecuteSuccessfully())

			gcm := f.KubernetesResource("ConfigMap", "d8-system", d8config.GeneratedConfigMapName)
			Expect(gcm.Exists()).Should(BeTrue(), "should create ConfigMap from DeckhouseConfig")

			dataMap := gcm.Field("data").Map()
			dataDump := gcm.Field("data").String()

			expectSectionValues := map[string]string{
				"global":    "paramStr: val1\n",
				"deckhouse": "paramStr: Debug\n",
			}
			Expect(dataMap).Should(HaveLen(len(expectSectionValues)), "generated ConfigMap should have sections for known DeckhouseConfig objects, got %s", dataDump)
			for moduleName, expectedSectionContent := range expectSectionValues {
				Expect(dataMap).Should(HaveKey(moduleName), "should have section for module '%s', got %s", moduleName, dataDump)
				Expect(dataMap[moduleName].String()).Should(Equal(expectedSectionContent), "should update section '%s', got %s", moduleName, dataDump)
			}
		})
	})

	Context("giving ConfigMap with some sections", func() {
		BeforeEach(func() {
			// See openapi schemas in testdata/test-deckhouse directory.
			validDeckhouseConfigs := `
---
apiVersion: deckhouse.io/v1
kind: DeckhouseConfig
metadata:
  name: global
spec:
  configVersion: 1
  configValues:
    paramStr: val1
---
apiVersion: deckhouse.io/v1
kind: DeckhouseConfig
metadata:
  name: deckhouse
spec:
  configVersion: 1
  configValues:
    paramStr: val1
---
apiVersion: deckhouse.io/v1
kind: DeckhouseConfig
metadata:
  name: prometheus
spec:
  configVersion: 1
  configValues:
    paramNum: 10
---
apiVersion: deckhouse.io/v1
kind: DeckhouseConfig
metadata:
  name: unknown-module-name
spec:
  configVersion: 1
  configValues:
    param1: val1
`

			cm := d8ConfigMap(d8config.GeneratedConfigMapName, `
global: |
  param2: val4
deckhouse: |
  logLevel: Info
`)

			f.BindingContexts.Set(f.KubeStateSet(validDeckhouseConfigs + cm))
			f.RunHook()
		})

		It("should create new sections and update values for existing", func() {
			Expect(f).To(ExecuteSuccessfully())

			gcm := f.KubernetesResource("ConfigMap", "d8-system", d8config.GeneratedConfigMapName)
			Expect(gcm.Exists()).Should(BeTrue(), "should not delete generated ConfigMap")

			dataMap := gcm.Field("data").Map()
			dataDump := gcm.Field("data").String()

			expectSectionValues := map[string]string{
				"global":     "paramStr: val1\n",
				"deckhouse":  "paramStr: val1\n",
				"prometheus": "paramNum: 10\n",
			}
			Expect(dataMap).Should(HaveLen(len(expectSectionValues)), "generated ConfigMap should have sections for known DeckhouseConfig objects, got %s", dataDump)
			for moduleName, expectedSectionContent := range expectSectionValues {
				Expect(dataMap).Should(HaveKey(moduleName), "should have section for module '%s', got %s", moduleName, dataDump)
				Expect(dataMap[moduleName].String()).Should(Equal(expectedSectionContent), "should update section '%s', got %s", moduleName, dataDump)
			}
		})
	})

	Context("giving DeckhouseConfig with invalid values", func() {
		BeforeEach(func() {
			existingConfigs := `
---
apiVersion: deckhouse.io/v1
kind: DeckhouseConfig
metadata:
  name: global
spec:
  configVersion: 1
  configValues:
    paramStr: 100
---
apiVersion: deckhouse.io/v1
kind: DeckhouseConfig
metadata:
  name: deckhouse
spec:
  configVersion: 1
  configValues:
    logLevel: Debug
`

			cm := d8ConfigMap(d8config.GeneratedConfigMapName, `
global: |
  param2: val4
deckhouse: |
  logLevel: Info
`)

			f.BindingContexts.Set(f.KubeStateSet(existingConfigs + cm))
			f.RunHook()
		})

		It("should create new sections and update values for existing", func() {
			Expect(f).ToNot(ExecuteSuccessfully(), "should fail on invalid values in DeckhouseConfig object")
		})
	})
})

func d8ConfigMap(cmName string, values string) string {
	var data map[string]string
	_ = k8syaml.Unmarshal([]byte(values), &data)

	cm := &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmName,
			Namespace: "d8-system",
		},
		Data: data,
	}

	cmDump, _ := k8syaml.Marshal(cm)
	return "---\n" + string(cmDump)
}