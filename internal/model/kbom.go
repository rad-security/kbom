package model

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

const (
	ociPrefix         = "oci"
	k8sPrefix         = "k8s"
	pkgPrefix         = "pkg"
	kubernetesPkgName = "k8s.io/kubernetes"
)

type KBOM struct {
	ID          string    `json:"id" yaml:"id"`
	BOMFormat   string    `json:"bom_format" yaml:"bom_format"`
	SpecVersion string    `json:"spec_version" yaml:"spec_version"`
	GeneratedAt time.Time `json:"generated_at" yaml:"generated_at"`
	GeneratedBy Tool      `json:"generated_by" yaml:"generated_by"`

	Cluster Cluster `json:"cluster" yaml:"cluster"`
}

type Tool struct {
	Vendor     string `json:"vendor" yaml:"vendor"`
	Name       string `json:"name" yaml:"name"`
	BuildTime  string `json:"build_time" yaml:"build_time"`
	Version    string `json:"version" yaml:"version"`
	Commit     string `json:"commit" yaml:"commit"`
	CommitTime string `json:"commit_time" yaml:"commit_time"`
}

type Cluster struct {
	Name         string     `json:"name" yaml:"name"`
	ID           string     `json:"id,omitempty" yaml:"id,omitempty"`
	CACertDigest string     `json:"ca_cert_digest" yaml:"ca_cert_digest"`
	K8sVersion   string     `json:"k8s_version" yaml:"k8s_version"`
	CNIVersion   string     `json:"cni_version,omitempty" yaml:"cni_version,omitempty"`
	Location     *Location  `json:"location" yaml:"location"`
	NodesCount   int        `json:"nodes_count" yaml:"nodes_count"`
	Nodes        []Node     `json:"nodes" yaml:"nodes"`
	Components   Components `json:"components" yaml:"components"`
}

func (c *Cluster) BOMRef() string {
	return fmt.Sprintf("%s:%s/%s@%s", pkgPrefix, k8sPrefix, url.QueryEscape(kubernetesPkgName), c.K8sVersion)
}

// PackageURL returns the Go module purl of Kubernetes, which vulnerability scanners match against
// Kubernetes advisories, e.g. pkg:golang/k8s.io/kubernetes@v1.35.8. Distribution suffixes of the
// version (1.35.8-eks-1234) are dropped.
func (c *Cluster) PackageURL() string {
	version, _, _ := strings.Cut(c.K8sVersion, "-")
	version, _, _ = strings.Cut(version, "+")
	if version == "" {
		return ""
	}

	return fmt.Sprintf("%s:golang/%s@v%s", pkgPrefix, kubernetesPkgName, strings.TrimPrefix(version, "v"))
}

func (c *Cluster) BOMName() string {
	return kubernetesPkgName
}

type Components struct {
	Images    []Image                 `json:"images,omitempty" yaml:"images,omitempty"`
	Addons    []Addon                 `json:"addons,omitempty" yaml:"addons,omitempty"`
	Resources map[string]ResourceList `json:"resources" yaml:"resources"`
}

// Addon is a well-known cluster add-on detected from the images running in the cluster.
type Addon struct {
	Name     string   `json:"name" yaml:"name"`
	Category string   `json:"category" yaml:"category"`
	Version  string   `json:"version" yaml:"version"`
	CPE      string   `json:"cpe,omitempty" yaml:"cpe,omitempty"`
	Images   []string `json:"images" yaml:"images"`
}

type Resource struct {
	Kind                 string            `json:"kind,omitempty" yaml:"kind,omitempty"`
	APIVersion           string            `json:"api_version,omitempty" yaml:"api_version,omitempty"`
	Name                 string            `json:"name" yaml:"name"`
	Namespace            string            `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	AdditionalProperties map[string]string `json:"additional_properties,omitempty" yaml:"additional_properties,omitempty"`
}

type ResourceList struct {
	Kind           string     `json:"kind" yaml:"kind"`
	APIVersion     string     `json:"api_version" yaml:"api_version"`
	Namespaced     bool       `json:"namespaced" yaml:"namespaced"`
	ResourcesCount int        `json:"count" yaml:"count"`
	Resources      []Resource `json:"resources,omitempty" yaml:"resources,omitempty"`
}

type Location struct {
	Name   string `json:"name" yaml:"name"`
	Region string `json:"region" yaml:"region"`
	Zone   string `json:"zone" yaml:"zone"`
}

type Node struct {
	Name                    string            `json:"name" yaml:"name"`
	Type                    string            `json:"type" yaml:"type"`
	Hostname                string            `json:"hostname" yaml:"hostname"`
	Capacity                *Capacity         `json:"capacity" yaml:"capacity"`
	Allocatable             *Capacity         `json:"allocatable" yaml:"allocatable"`
	Labels                  map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Annotations             map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	MachineID               string            `json:"machine_id" yaml:"machine_id"`
	Architecture            string            `json:"architecture" yaml:"architecture"`
	ContainerRuntimeVersion string            `json:"container_runtime_version" yaml:"container_runtime_version"`
	BootID                  string            `json:"boot_id" yaml:"boot_id"`
	KernelVersion           string            `json:"kernel_version" yaml:"kernel_version"`
	KubeProxyVersion        string            `json:"kube_proxy_version" yaml:"kube_proxy_version"`
	KubeletVersion          string            `json:"kubelet_version" yaml:"kubelet_version"`
	OperatingSystem         string            `json:"operating_system" yaml:"operating_system"`
	OsImage                 string            `json:"os_image" yaml:"os_image"`
}

type Image struct {
	FullName     string `json:"full_name" yaml:"full_name"`
	Name         string `json:"name" yaml:"name"`
	Version      string `json:"version" yaml:"version"`
	Digest       string `json:"digest" yaml:"digest"`
	ControlPlane bool   `json:"-" yaml:"-"`
}

func (i *Image) PkgID() string {
	parts := strings.Split(i.Name, "/")
	baseName := fmt.Sprintf("%s:%s/%s", pkgPrefix, ociPrefix, parts[len(parts)-1])

	urlValues := url.Values{
		"repository_url": []string{i.Name},
	}

	if i.Version != "" {
		urlValues.Add("tag", i.Version)
	}

	if i.Digest != "" {
		baseName = fmt.Sprintf("%s@%s", baseName, url.QueryEscape(i.Digest))
	}

	return fmt.Sprintf("%s?%s", baseName, urlValues.Encode())
}

type Capacity struct {
	CPU              string `json:"cpu" yaml:"cpu"`
	Memory           string `json:"memory" yaml:"memory"`
	Pods             string `json:"pods" yaml:"pods"`
	EphemeralStorage string `json:"ephemeral_storage" yaml:"ephemeral_storage"`
}
