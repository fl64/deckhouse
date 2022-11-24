# Changelog v1.41

## Know before update


 - The `auth.password` option is deprecated in the `cilium-hubble` module.
 - The `auth.password` option is deprecated in the `dashboard` module.
 - The `auth.password` option is deprecated in the `deckhouse-web` module
 - The `auth.password` option is deprecated in the `istio` module.
 - The `auth.password` option is deprecated in the `openvpn` module.
 - The `auth.password` option is deprecated in the `prometheus` module.
 - The `auth.status.password` and `auth.webui.password` options are deprecated in the `upmeter` module.
 - The `deckhouse` ConfigMap is forbidden to edit; use `ModuleConfig` object to change module configuration.

## Features


 - **[cert-manager]** Ability to disable the `letsencrypt` and `letsencrypt-staging` ClusterIssuers creation. [#3042](https://github.com/deckhouse/deckhouse/pull/3042)
 - **[cilium-hubble]** The `auth.password` option is deprecated. Consider using the `user-authn` module. [#1729](https://github.com/deckhouse/deckhouse/pull/1729)
    The `auth.password` option is deprecated in the `cilium-hubble` module.
 - **[dashboard]** The `auth.password` option is deprecated. Consider using the `user-authn` module. [#1729](https://github.com/deckhouse/deckhouse/pull/1729)
    The `auth.password` option is deprecated in the `dashboard` module.
 - **[deckhouse-controller]** Use ModuleConfig objects to configure deckhouse modules. [#1729](https://github.com/deckhouse/deckhouse/pull/1729)
    The `deckhouse` ConfigMap is forbidden to edit; use `ModuleConfig` object to change module configuration.
 - **[deckhouse-web]** The `auth.password` option is deprecated. Consider using the `user-authn` module. [#1729](https://github.com/deckhouse/deckhouse/pull/1729)
    The `auth.password` option is deprecated in the `deckhouse-web` module
 - **[flant-integration]** Add scrape telemetry metrics (with prefix d8_telemetry) from deckhouse pod via new service [#2896](https://github.com/deckhouse/deckhouse/pull/2896)
 - **[istio]** The `auth.password` option is deprecated. Consider using the `user-authn` module. [#1729](https://github.com/deckhouse/deckhouse/pull/1729)
    The `auth.password` option is deprecated in the `istio` module.
 - **[node-manager]** Add an alert about missing control-plane taints on the `master` node group [#3057](https://github.com/deckhouse/deckhouse/pull/3057)
 - **[openvpn]** The `auth.password` option is deprecated. Consider using the `user-authn` module. [#1729](https://github.com/deckhouse/deckhouse/pull/1729)
    The `auth.password` option is deprecated in the `openvpn` module.
 - **[prometheus]** The `auth.password` option is deprecated. Consider using the `user-authn` module. [#1729](https://github.com/deckhouse/deckhouse/pull/1729)
    The `auth.password` option is deprecated in the `prometheus` module.
 - **[upmeter]** Added probe uptime in public status API to use in e2e tests [#2991](https://github.com/deckhouse/deckhouse/pull/2991)
 - **[upmeter]** The `auth.status.password` and `auth.webui.password` options are deprecated. Consider using the `user-authn` module. [#1729](https://github.com/deckhouse/deckhouse/pull/1729)
    The `auth.status.password` and `auth.webui.password` options are deprecated in the `upmeter` module.

## Fixes


 - **[admission-policy-engine]** Watch only desired (constrainted) resources by validation webhook [#3027](https://github.com/deckhouse/deckhouse/pull/3027)
 - **[cloud-provider-vsphere]** Bump csi driver to v2.5.4 [#3089](https://github.com/deckhouse/deckhouse/pull/3089)
 - **[cloud-provider-yandex]** Removed the Standard layout from the documentation, as it doesn't work. [#3108](https://github.com/deckhouse/deckhouse/pull/3108)
 - **[cloud-provider-yandex]** In case of wget and curl utility usage inside pods, proxy (and proxy ignore) will work. [#3031](https://github.com/deckhouse/deckhouse/pull/3031)
    The `cloud-provider-yandex` module will be restarted if a proxy is enabled in the cluster.
 - **[istio]** Fixed istio control-plane alerts: `D8IstioActualVersionIsNotInstalled`, `D8IstioDesiredVersionIsNotInstalled`. [#3024](https://github.com/deckhouse/deckhouse/pull/3024)
 - **[linstor]** In case of wget and curl utility usage inside pods, proxy (and proxy ignore) will work. [#3031](https://github.com/deckhouse/deckhouse/pull/3031)
    The `linstor` module will be restarted if a proxy is enabled in the cluster.
 - **[node-manager]** Calculate resource requests for a stanby-holder Pod as a percentage of a node's capacity. [#2959](https://github.com/deckhouse/deckhouse/pull/2959)
 - **[prometheus]** Update Grafana Home dashboard. [#3015](https://github.com/deckhouse/deckhouse/pull/3015)
 - **[snapshot-controller]** In case of wget and curl utility usage inside pods, proxy (and proxy ignore) will work. [#3031](https://github.com/deckhouse/deckhouse/pull/3031)
    The `snapshot-controller` module will be restarted if a proxy is enabled in the cluster.

## Chore


 - **[basic-auth]** Improved error message on unexpected number of fields in the credentials secret [#3039](https://github.com/deckhouse/deckhouse/pull/3039)
 - **[cloud-provider-yandex]** Remove duplicated keys from YAML in test. [#3094](https://github.com/deckhouse/deckhouse/pull/3094)
 - **[deckhouse-config]** fix deckhouse-config webhook build [#3111](https://github.com/deckhouse/deckhouse/pull/3111)
 - **[extended-monitoring]** Pass HTTP_PROXY, HTTPS_PROXY and NO_PROXY environment variables into image-availability-exporter [#3011](https://github.com/deckhouse/deckhouse/pull/3011)
    Pod image-availability-exporter will be restarted if Deckhouse proxy parameters are set

