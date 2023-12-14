# KBOM Schema

The section below describes the high level object model for KBOM.

## Cluster Details

Instances:

- Name
- Hostname
- CloudType
- Creation Timestamp
- Capacity
- Allocatable resources
- OS Version
- Kernel Version
- Architecture
- CRI Version
- Kubelet Version
- Kube Proxy Version

Images:

- Name
- FullName
- Version
- Digest

KubeObjects:

- Kind
- Api Version
- Count
- Details

This overall structure provides a base spec to be expanded upon by the community.

The intent of the standard is to be extensible to support various use cases across the industry.
