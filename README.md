# KBOM - Kubernetes Bill of Materials

The Kubernetes Bill of Materials (KBOM) standard provides insight into container orchestration tools widely used across the industry. As a first draft, we have created a rough specification which should fall in line with other Bill of Materials (BOM) standards.

The KBOM project provides an initial specification in JSON and has been constructed for extensibilty across various cloud service providers (CSPs) as well as DIY Kubernetes. 

## BOM Requirements

### High level Object Model:

BoM Format Information 

Cluster Details

Instances
- Name
- Hostname
- CloudType
- Creation Timestamp
- Capacity
- OS Version
- Kernel Version
- Architecture
- CRI Version
- Kubelet Version
- Kube Proxy Version

Images
- Name
- Digest

KubeObjects
- Kind
- Api Version
- Count
- Details

This overall structure provides a base spec to be expanded upon by the community. The intent of the standard is to be extensible to support various use cases across the industry.

## How to generate KBOM for your cluster?

Use `kbom` CLI tool!

To install it:

```sh
go install github.com/ksoclabs/kbom
```

or

```sh
make build
```

## KBOM CLI documentation

`KBOM generate` generates KBOM file for your Kubernetes cluster

```sh
kbom generate [flags]
```

### Options

```sh
  -f, --format string     Format (json, yaml) (default "json")
  -h, --help              Help for generate
  -p, --out-path string   Path to write KBOM to (default ".")
  -o, --output string     Output (stdout, file) (default "stdout")
      --short             Short - only include metadata, nodes, images and resources counters
```

### Options inherited from parent commands

```sh
  -v, --verbose   enable verbose logging (DEBUG and below)
```
