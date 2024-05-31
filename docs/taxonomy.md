# Custom RAD KBOM Taxonomy

This is the RAD KBOM CycloneDX property namespace and name taxonomy. All of the namespaces are prefixed with `rad:kbom:`.

Following Taxonomy is used by the `KBOM` tool as extension to: [https://github.com/CycloneDX/cyclonedx-property-taxonomy](https://github.com/CycloneDX/cyclonedx-property-taxonomy).

## `rad:kbom:k8s:component` Namespace Taxonomy

| Namespace                            | Description                                                       |
| ------------------------------------ | ----------------------------------------------------------------- |
| `rad:kbom:k8s:component:apiVersion` | API Version of the Kubernetes component.                          |
| `rad:kbom:k8s:component:namespace`  | Namespace of the  Kubernetes component.                           |

## `rad:kbom:k8s:cluster` Namespace Taxonomy

| Property                                  | Description                    |
| ----------------------------------------- | ------------------------------ |
| `rad:kbom:k8s:cluster:location:name`     | Name of the location.          |
| `rad:kbom:k8s:cluster:location:region`   | Region of the cluster.         |
| `rad:kbom:k8s:cluster:location:zone`     | Zone where cluster is located. |

## `rad:kbom:k8s:node` Namespace Taxonomy

| Property                                           | Description                          |
| -------------------------------------------------- | ------------------------------------ |
| `rad:kbom:k8s:node:osImage`                       | Node's operating system image        |
| `rad:kbom:k8s:node:arch`                          | Node's architecture                  |
| `rad:kbom:k8s:node:kernel`                        | Node's kernel version                |
| `rad:kbom:k8s:node:bootId`                        | Node's Boot identifier               |
| `rad:kbom:k8s:node:type`                          | Node's type                          |
| `rad:kbom:k8s:node:operatingSystem`               | Node's operating system              |
| `rad:kbom:k8s:node:machineId`                     | Node's machine identifier            |
| `rad:kbom:k8s:node:hostname`                      | Node's hostname                      |
| `rad:kbom:k8s:node:containerRuntimeVersion`       | Node's container runtime version     |
| `rad:kbom:k8s:node:kubeletVersion`                | Node's kubelet version               |
| `rad:kbom:k8s:node:kubeProxyVersion`              | Node's kube proxy version            |
| `rad:kbom:k8s:node:capacity:cpu`                  | Node's CPU capacity                  |
| `rad:kbom:k8s:node:capacity:memory`               | Node's Memory capacity               |
| `rad:kbom:k8s:node:capacity:pods`                 | Node's Pods capacity                 |
| `rad:kbom:k8s:node:capacity:ephemeralStorage`     | Node's ephemeral storage capacity    |
| `rad:kbom:k8s:node:allocatable:cpu`               | Node's allocatable CPU               |
| `rad:kbom:k8s:node:allocatable:memory`            | Node's allocatable Memory            |
| `rad:kbom:k8s:node:allocatable:pods`              | Node's allocatable Pods              |
| `rad:kbom:k8s:node:allocatable:ephemeralStorage`  | Node's allocatable ephemeral storage |

## `rad:kbom:pkg` Namespace Taxonomy

| Property                          | Description                                        |
| --------------------------------- | -------------------------------------------------- |
| `rad:kbom:pkg:type`              | Type of the package.                               |
| `rad:kbom:pkg:name`              | Name of the package.                               |
| `rad:kbom:pkg:version`           | Version of the package.                            |
| `rad:kbom:pkg:digest`            | Digest of the package.                             |
