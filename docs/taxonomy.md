# Custom KSOC KBOM Taxonomy

This is the KSOC KBOM CycloneDX property namespace and name taxonomy. All of the namespaces are prefixed with `ksoc:kbom:`.

Following Taxonomy is used by the `KBOM` tool as extension to: [https://github.com/CycloneDX/cyclonedx-property-taxonomy](https://github.com/CycloneDX/cyclonedx-property-taxonomy).

## `ksoc:kbom:k8s:component` Namespace Taxonomy

| Namespace                            | Description                                                       |
| ------------------------------------ | ----------------------------------------------------------------- |
| `ksoc:kbom:k8s:component:apiVersion` | API Version of the Kubernetes component.                          |
| `ksoc:kbom:k8s:component:namespace`  | Namespace of the  Kubernetes component.                           |

## `ksoc:kbom:k8s:cluster` Namespace Taxonomy

| Property                                  | Description                    |
| ----------------------------------------- | ------------------------------ |
| `ksoc:kbom:k8s:cluster:location:name`     | Name of the location.          |
| `ksoc:kbom:k8s:cluster:location:region`   | Region of the cluster.         |
| `ksoc:kbom:k8s:cluster:location:zone`     | Zone where cluster is located. |

## `ksoc:kbom:k8s:node` Namespace Taxonomy

| Property                                           | Description                          |
| -------------------------------------------------- | ------------------------------------ |
| `ksoc:kbom:k8s:node:osImage`                       | Node's operating system image        |
| `ksoc:kbom:k8s:node:arch`                          | Node's architecture                  |
| `ksoc:kbom:k8s:node:kernel`                        | Node's kernel version                |
| `ksoc:kbom:k8s:node:bootId`                        | Node's Boot identifier               |
| `ksoc:kbom:k8s:node:type`                          | Node's type                          |
| `ksoc:kbom:k8s:node:operatingSystem`               | Node's operating system              |
| `ksoc:kbom:k8s:node:machineId`                     | Node's machine identifier            |
| `ksoc:kbom:k8s:node:hostname`                      | Node's hostname                      |
| `ksoc:kbom:k8s:node:containerRuntimeVersion`       | Node's container runtime version     |
| `ksoc:kbom:k8s:node:kubeletVersion`                | Node's kubelet version               |
| `ksoc:kbom:k8s:node:kubeProxyVersion`              | Node's kube proxy version            |
| `ksoc:kbom:k8s:node:capacity:cpu`                  | Node's CPU capacity                  |
| `ksoc:kbom:k8s:node:capacity:memory`               | Node's Memory capacity               |
| `ksoc:kbom:k8s:node:capacity:pods`                 | Node's Pods capacity                 |
| `ksoc:kbom:k8s:node:capacity:ephemeralStorage`     | Node's ephemeral storage capacity    |
| `ksoc:kbom:k8s:node:allocatable:cpu`               | Node's allocatable CPU               |
| `ksoc:kbom:k8s:node:allocatable:memory`            | Node's allocatable Memory            |
| `ksoc:kbom:k8s:node:allocatable:pods`              | Node's allocatable Pods              |
| `ksoc:kbom:k8s:node:allocatable:ephemeralStorage`  | Node's allocatable ephemeral storage |

## `ksoc:kbom:pkg` Namespace Taxonomy

| Property                          | Description                                        |
| --------------------------------- | -------------------------------------------------- |
| `ksoc:kbom:pkg:type`              | Type of the package.                               |
| `ksoc:kbom:pkg:name`              | Name of the package.                               |
| `ksoc:kbom:pkg:version`           | Version of the package.                            |
| `ksoc:kbom:pkg:digest`            | Digest of the package.                             |
