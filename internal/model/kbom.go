package model

import (
	"fmt"
	"time"
)

type KBOM struct {
	ID          string    `json:"id"`
	BOMFormat   string    `json:"bom_format"`
	SpecVersion string    `json:"spec_version"`
	GeneratedAt time.Time `json:"generated_at"`
	GeneratedBy Tool      `json:"generated_by"`

	Cluster Cluster `json:"cluster"`
}

type Tool struct {
	Vendor     string `json:"vendor"`
	Name       string `json:"name"`
	BuildTime  string `json:"build_time"`
	Version    string `json:"version"`
	Commit     string `json:"commit"`
	CommitTime string `json:"commit_time"`
}

type Cluster struct {
	Name         string     `json:"name"`
	CACertDigest string     `json:"ca_cert_digest"`
	K8sVersion   string     `json:"k8s_version"`
	CNIVersion   string     `json:"cni_version,omitempty"`
	Location     *Location  `json:"location"`
	NodesCount   int        `json:"nodes_count"`
	Nodes        []Node     `json:"nodes"`
	Components   Components `json:"components"`
}

type Components struct {
	Images    []Image                 `json:"images,omitempty"`
	Resources map[string]ResourceList `json:"resources"`
}

type Resource struct {
	Kind       string `json:"kind,omitempty"`
	APIVersion string `json:"api_version,omitempty"`
	Name       string `json:"name"`
	Namespace  string `json:"namespace,omitempty"`
}

type ResourceList struct {
	Kind           string     `json:"kind"`
	APIVersion     string     `json:"api_version"`
	Namespaced     bool       `json:"namespaced"`
	ResourcesCount int        `json:"count"`
	Resources      []Resource `json:"resources,omitempty"`
}

type Location struct {
	Name   string `json:"name"`
	Region string `json:"region"`
	Zone   string `json:"zone"`
}

type Node struct {
	Name                    string            `json:"name"`
	Type                    string            `json:"type"`
	Hostname                string            `json:"hostname"`
	Capacity                *Capacity         `json:"capacity"`
	Allocatable             *Capacity         `json:"allocatable"`
	Labels                  map[string]string `json:"labels"`
	Annotations             map[string]string `json:"annotations"`
	MachineID               string            `json:"machine_id"`
	Architecture            string            `json:"architecture"`
	ContainerRuntimeVersion string            `json:"container_runtime_version"`
	BootID                  string            `json:"boot_id"`
	KernelVersion           string            `json:"kernel_version"`
	KubeProxyVersion        string            `json:"kube_proxy_version"`
	KubeletVersion          string            `json:"kubelet_version"`
	OperatingSystem         string            `json:"operating_system"`
	OsImage                 string            `json:"os_image"`
}

type Image struct {
	FullName  string `json:"full_name"`
	Name      string `json:"name"`
	Version   string `json:"version"`
	Digest    string `json:"digest"`
	Namespace string `json:"namespace"`
}

func (i *Image) PkgID() string {
	if i.Digest == "" && i.Version == "" {
		return fmt.Sprintf("pkg:%s", i.Name)
	}

	if i.Digest == "" {
		return fmt.Sprintf("pkg:%s:%s", i.Name, i.Version)
	}

	if i.Version == "" {
		return fmt.Sprintf("pkg:%s@%s", i.Name, i.Digest)
	}

	return fmt.Sprintf("pkg:%s:%s@%s", i.Name, i.Version, i.Digest)
}

type Capacity struct {
	CPU              string `json:"cpu"`
	Memory           string `json:"memory"`
	Pods             string `json:"pods"`
	EphemeralStorage string `json:"ephemeral_storage"`
}
