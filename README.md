# KBOM - Kubernetes Bill of Materials

TBD

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
