# Contributing

KBOM is [Apache 2.0 licensed](https://github.com/rad-security/kbom/blob/main/LICENSE) and
accepts contributions via GitHub pull requests. This document outlines
some of the conventions on to make it easier to get your contribution
accepted.

We gratefully welcome improvements to issues and documentation as well as to
code.

## Certificate of Origin

By contributing to this project you agree to the Developer Certificate of
Origin (DCO). This document was created by the Linux Kernel community and is a
simple statement that you, as a contributor, have the legal right to make the
contribution.

We require all commits to be signed. By signing off with your signature, you
certify that you wrote the patch or otherwise have the right to contribute the
material by the rules of the [DCO](DCO):

`Signed-off-by: Firstname Lastname <firstname.lastname@example.com>`

The signature must contain your real name
(sorry, no pseudonyms or anonymous contributions)
If your `user.name` and `user.email` are configured in your Git config,
you can sign your commit automatically with `git commit -s`.

## Communications

To discuss ideas and specifications we use [GitHub Discussions](https://github.com/rad-security/kbom/discussions).

## How to run the KBOM generator in local environment

Prerequisites:

* go >= 1.20
* kind
* golangci-lint

Initialise repo:

```bash
make initialise
```

To generate your first KBOM file we need to have access to a Kubernetes cluster.
If you don't have any you could create your local cluster with `Kind`.

Create kind cluster(optional):

```bash
kind create cluster --name kbom-test
```

Build `kbom` binary:

```bash
make build
```

Generate your first `kbom` file:

```bash
./kbom generate
```

## Acceptance policy

These things will make a PR more likely to be accepted:

* a well-described requirement
* tests for new code
* tests for old code!
* new code and tests follow the conventions in old code and tests
* a good commit message (see below)
* all code must abide [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
* names should abide [What's in a name](https://talks.golang.org/2014/names.slide#1)
* code must build on both Linux and Darwin, via plain `go build`
* code should have appropriate test coverage and tests should be written to work with `go test`

In general, we will merge a PR once one maintainer has endorsed it.
For substantial changes, more people may become involved, and you might
get asked to resubmit the PR or divide the changes into more than one PR.
