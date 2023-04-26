# KBOM - Kubernetes Bill of Materials

![GitHub release (latest by date)](https://img.shields.io/github/v/release/ksoclabs/kbom)
![Hex.pm](https://img.shields.io/hexpm/l/apa)
[![Go Report Card](https://goreportcard.com/badge/github.com/ksoclabs/kbom)](https://goreportcard.com/report/github.com/ksoclabs/kbom)
[![FOSSA Status](https://app.fossa.com/api/projects/custom%2B37386%2Fgithub.com%2Fksoclabs%2Fkbom.svg?type=shield)](https://app.fossa.com/projects/custom%2B37386%2Fgithub.com%2Fksoclabs%2Fkbom?ref=badge_shield)

The Kubernetes Bill of Materials (KBOM) standard provides insight into container orchestration tools widely used across the industry.

As a first draft, we have created a rough specification which should fall in line with other Bill of Materials (BOM) standards.

The KBOM project provides an initial specification in JSON and has been constructed for extensibilty across various cloud service providers (CSPs) as well as DIY Kubernetes.

## Getting Started

### Installation

```sh
go install github.com/ksoclabs/kbom
```

or

```sh
make build
```

### Usage

`KBOM generate` generates a KBOM file for your Kubernetes cluster

```sh
kbom generate [flags]
```

Optional flags include:

```
  -f, --format string     Format (json, yaml) (default "json")
  -h, --help              Help for generate
  -p, --out-path string   Path to write KBOM to (default ".")
  -o, --output string     Output (stdout, file) (default "stdout")
      --short             Short - only include metadata, nodes, images and resources counters
```

## Schema

The high level object model can be found [here](docs/schema.md).

## Contributing

KBOM is Apache 2.0 licensed and accepts contributions via GitHub pull requests. See the [CONTRIBUTING](CONTRIBUTING.md) file for details.
