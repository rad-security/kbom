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

func (c *Cluster) BOMName() string {
	return kubernetesPkgName
}

type Components struct {
	Images    []Image                 `json:"images,omitempty" yaml:"images,omitempty"`
	Resources map[string]ResourceList `json:"resources" yaml:"resources"`
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
	Labels                  map[string]string `json:"labels" yaml:"labels"`
	Annotations             map[string]string `json:"annotations" yaml:"annotations"`
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
