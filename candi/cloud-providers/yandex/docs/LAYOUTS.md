---
title: "Cloud provider — Yandex.Cloud: Layouts"
---

Three layouts are supported. Below is more information about each of them.

## WithoutNAT

In this layout, NAT (of any kind) is not used, and each node is assigned a public IP.

> **Caution!** Currently, the cloud-provider-yandex module does not support Security Groups; thus, is why all cluster nodes connect directly to the Internet.

![resources](https://docs.google.com/drawings/d/e/2PACX-1vTgwXWsNX6CKCRaMf5t6rl3kpKQQFHK6T8Dsg1jAwAwYaN1MRbxKFsSFQHeo1N3Qec4etPpeA0guB6-/pub?w=812&h=655)
<!--- Source: https://docs.google.com/drawings/d/1I7M9DquzLNu-aTjqLx1_6ZexPckL__-501Mt393W1fw/edit --->

Example of the layout configuration:

```yaml
apiVersion: deckhouse.io/v1
kind: YandexClusterConfiguration
layout: WithoutNAT
provider:
  cloudID: <CLOUD_ID>
  folderID: <FOLDER_ID>
  serviceAccountJSON: |
    {"test": "test"}
masterNodeGroup:
  replicas: 1
  instanceClass:
    cores: 4
    memory: 8192
    imageID: <IMAGE_ID>
    externalIPAddresses:
    - "198.51.100.5"
    - "Auto"
    externalSubnetID: <EXTERNAL_SUBNET_ID>
    zones:
    - ru-central1-a
    - ru-central1-b
nodeGroups:
- name: khm
  replicas: 1
  instanceClass:
    cores: 4
    memory: 8192
    imageID: testtest
    coreFraction: 50
    externalIPAddresses:
    - "198.51.100.5"
    - "Auto"
    externalSubnetID: tewt243tewsdf
    zones:
    - ru-central1-a
sshPublicKey: "<SSH_PUBLIC_KEY>"
nodeNetworkCIDR: 192.168.12.13/24
existingNetworkID: <EXISTING_NETWORK_ID>
dhcpOptions:
  domainName: test.local
  domainNameServers:
  - 8.8.8.8
  - 8.8.4.4
```

## WithNATInstance

In this placement strategy, Deckhouse creates a NAT instance and adds a rule to a route table containing a route to 0.0.0.0/0 with a NAT instance as the next hop.

If the `withNATInstance.externalSubnetID` parameter is set, the NAT instance will be created in this subnet.

IF the `withNATInstance.externalSubnetID` parameter is not set and `withNATInstance.internalSubnetID` is set, the NAT instance will be created in this last subnet.

If neither `withNATInstance.externalSubnetID` nor `withNATInstance.internalSubnetID` is set, the NAT instance will be created in the  `ru-central1-c` zone.

![resources](https://docs.google.com/drawings/d/e/2PACX-1vSnNqebgRdwGP8lhKMJfrn5c0QXDpe9YdmIlK4eDberysLLgYiKNuwaPLHcyQhJigvQ21SANH89uipE/pub?w=812&h=655)
<!--- Source: https://docs.google.com/drawings/d/1oVpZ_ldcuNxPnGCkx0dRtcAdL7BSEEvmsvbG8Aif1pE/edit --->

Example of the layout configuration:

```yaml
apiVersion: deckhouse.io/v1
kind: YandexClusterConfiguration
layout: WithNATInstance
withNATInstance:
  natInstanceExternalAddress: <NAT_INSTANCE_EXTERNAL_ADDRESS>
  internalSubnetID: <INTERNAL_SUBNET_ID>
  externalSubnetID: <EXTERNAL_SUBNET_ID>
provider:
  cloudID: <CLOUD_ID>
  folderID: <FOLDER_ID>
  serviceAccountJSON: |
    {"test": "test"}
masterNodeGroup:
  replicas: 1
  instanceClass:
    cores: 4
    memory: 8192
    imageID: <IMAGE_ID>
    externalIPAddresses:
    - "1.1.1.1"
    - "Auto"
    externalSubnetID: <EXTERNAL_SUBNET_ID>
    zones:
    - ru-central1-a
    - ru-central1-b
nodeGroups:
- name: khm
  replicas: 1
  instanceClass:
    cores: 4
    memory: 8192
    imageID: <IMAGE_ID>
    coreFraction: 50
    externalIPAddresses:
    - "1.1.1.1"
    - "Auto"
    externalSubnetID: <EXTERNAL_SUBNET_ID>
    zones:
    - ru-central1-a
sshPublicKey: "<SSH_PUBLIC_KEY>"
nodeNetworkCIDR: 192.168.12.13/24
existingNetworkID: <EXISTING_NETWORK_ID>
dhcpOptions:
  domainName: test.local
  domainNameServers:
  - 8.8.8.8
  - 8.8.4.4
```
