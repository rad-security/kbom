# KBOM - Kubernetes Bill of Materials

The Kubernetes Bill of Materials (KBOM) standard provides insight into container orchestration tools widely used across the industry. 

As a first draft, we have created a rough specification which should fall in line with other Bill of Materials (BOM) standards.

The KBOM project provides an initial specification in JSON and has been constructed for extensibilty across various cloud service providers (CSPs) as well as DIY Kubernetes. 

## Getting Started

### Prerequisites

### Installation



## Schema

The high level object model can be found [here](docs/schema.md).

## Contributing

KBOM is Apache 2.0 licensed and accepts contributions via GitHub pull requests. See the [CONTRIBUTING](CONTRIBUTING.md) file for details.


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
