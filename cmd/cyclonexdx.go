package cmd

import (
	"fmt"
	"hash/fnv"
	"slices"
	"time"

	"github.com/CycloneDX/cyclonedx-go"
	"github.com/google/uuid"
	"github.com/mitchellh/hashstructure/v2"

	"github.com/rad-security/kbom/internal/model"
)

const (
	CdxPrefix        = "cdx:"
	RADPrefix        = "rad:kbom:"
	K8sComponentType = "k8s:component:type"
	K8sComponentName = "k8s:component:name"

	ClusterType   = "cluster"
	NodeType      = "node"
	ContainerType = "container"
)

func transformToCycloneDXBOM(kbom *model.KBOM) *cyclonedx.BOM { //nolint:funlen
	cdxBOM := cyclonedx.NewBOM()

	cdxBOM.SerialNumber = uuid.New().URN()
	cdxBOM.Metadata = &cyclonedx.Metadata{
		Timestamp: time.Now().Format(time.RFC3339),
		Tools: &[]cyclonedx.Tool{
			{
				Vendor:  kbom.GeneratedBy.Vendor,
				Name:    kbom.GeneratedBy.Name,
				Version: kbom.GeneratedBy.Version,
			},
		},
	}

	components := []cyclonedx.Component{}
	dependencies := []cyclonedx.Dependency{}
	clusterProperties := []cyclonedx.Property{
		{
			Name:  CdxPrefix + K8sComponentType,
			Value: ClusterType,
		},
		{
			Name:  CdxPrefix + "k8s:component:name",
			Value: kbom.Cluster.Name,
		},
		{
			Name:  RADPrefix + "k8s:cluster:nodes",
			Value: fmt.Sprintf("%d", kbom.Cluster.NodesCount),
		},
	}

	if kbom.Cluster.Location.Name != "" && kbom.Cluster.Location.Name != "unknown" {
		clusterProperties = append(clusterProperties, cyclonedx.Property{
			Name:  RADPrefix + "k8s:cluster:location:name",
			Value: kbom.Cluster.Location.Name,
		})
	}

	if kbom.Cluster.Location.Region != "" {
		clusterProperties = append(clusterProperties, cyclonedx.Property{
			Name:  RADPrefix + "k8s:cluster:location:region",
			Value: kbom.Cluster.Location.Region,
		})
	}

	if kbom.Cluster.Location.Zone != "" {
		clusterProperties = append(clusterProperties, cyclonedx.Property{
			Name:  RADPrefix + "k8s:cluster:location:zone",
			Value: kbom.Cluster.Location.Zone,
		})
	}

	clusterComponent := cyclonedx.Component{
		BOMRef:     kbom.Cluster.BOMRef(),
		Type:       cyclonedx.ComponentTypePlatform,
		Name:       kbom.Cluster.BOMName(),
		Version:    kbom.Cluster.K8sVersion,
		Properties: &clusterProperties,
	}
	cdxBOM.Metadata.Component = &clusterComponent

	clusterDependencies := make(map[string]string)
	for i := range kbom.Cluster.Nodes {
		n := kbom.Cluster.Nodes[i]
		bomRef := id(n)
		components = append(components, cyclonedx.Component{
			BOMRef: bomRef,
			Type:   cyclonedx.ComponentTypePlatform,
			Name:   n.Name,
			Properties: &[]cyclonedx.Property{
				{
					Name:  CdxPrefix + K8sComponentType,
					Value: NodeType,
				},
				{
					Name:  CdxPrefix + K8sComponentName,
					Value: n.Name,
				},
				{
					Name:  RADPrefix + "k8s:node:osImage",
					Value: n.OsImage,
				},
				{
					Name:  RADPrefix + "k8s:node:arch",
					Value: n.Architecture,
				},
				{
					Name:  RADPrefix + "k8s:node:kernel",
					Value: n.KernelVersion,
				},
				{
					Name:  RADPrefix + "k8s:node:bootId",
					Value: n.BootID,
				},
				{
					Name:  RADPrefix + "k8s:node:type",
					Value: n.Type,
				},
				{
					Name:  RADPrefix + "k8s:node:operatingSystem",
					Value: n.OperatingSystem,
				},
				{
					Name:  RADPrefix + "k8s:node:machineId",
					Value: n.MachineID,
				},
				{
					Name:  RADPrefix + "k8s:node:hostname",
					Value: n.Hostname,
				},
				{
					Name:  RADPrefix + "k8s:node:containerRuntimeVersion",
					Value: n.ContainerRuntimeVersion,
				},
				{
					Name:  RADPrefix + "k8s:node:kubeletVersion",
					Value: n.KubeletVersion,
				},
				{
					Name:  RADPrefix + "k8s:node:kubeProxyVersion",
					Value: n.KubeProxyVersion,
				},
				{
					Name:  RADPrefix + "k8s:node:capacity:cpu",
					Value: n.Capacity.CPU,
				},
				{
					Name:  RADPrefix + "k8s:node:capacity:memory",
					Value: n.Capacity.Memory,
				},
				{
					Name:  RADPrefix + "k8s:node:capacity:pods",
					Value: n.Capacity.Pods,
				},
				{
					Name:  RADPrefix + "k8s:node:capacity:ephemeralStorage",
					Value: n.Capacity.EphemeralStorage,
				},
				{
					Name:  RADPrefix + "k8s:node:allocatable:cpu",
					Value: n.Allocatable.CPU,
				},
				{
					Name:  RADPrefix + "k8s:node:allocatable:memory",
					Value: n.Allocatable.Memory,
				},
				{
					Name:  RADPrefix + "k8s:node:allocatable:pods",
					Value: n.Allocatable.Pods,
				},
				{
					Name:  RADPrefix + "k8s:node:allocatable:ephemeralStorage",
					Value: n.Allocatable.EphemeralStorage,
				},
			},
		})
		clusterDependencies[bomRef] = bomRef
	}

	for _, img := range kbom.Cluster.Components.Images {
		bomRef := img.PkgID()
		container := cyclonedx.Component{
			BOMRef:     bomRef,
			Type:       cyclonedx.ComponentTypeContainer,
			Name:       img.Name,
			Version:    img.Digest,
			PackageURL: bomRef,
			Properties: &[]cyclonedx.Property{
				{
					Name:  CdxPrefix + K8sComponentType,
					Value: ContainerType,
				},
				{
					Name:  CdxPrefix + K8sComponentName,
					Value: img.Name,
				},
				{
					Name:  RADPrefix + "pkg:type",
					Value: "oci",
				},
				{
					Name:  RADPrefix + "pkg:name",
					Value: img.Name,
				},
				{
					Name:  RADPrefix + "pkg:version",
					Value: img.Version,
				},
				{
					Name:  RADPrefix + "pkg:digest",
					Value: img.Digest,
				},
			},
		}

		components = append(components, container)

		if img.ControlPlane {
			clusterDependencies[bomRef] = bomRef
		}
	}

	for _, resList := range kbom.Cluster.Components.Resources {
		for _, res := range resList.Resources {
			properties := []cyclonedx.Property{
				{
					Name:  CdxPrefix + K8sComponentType,
					Value: resList.Kind,
				},
				{
					Name:  CdxPrefix + K8sComponentName,
					Value: res.Name,
				},
				{
					Name:  RADPrefix + "k8s:component:apiVersion",
					Value: resList.APIVersion,
				},
			}

			if resList.Namespaced {
				properties = append(properties, cyclonedx.Property{
					Name:  RADPrefix + "k8s:component:namespace",
					Value: res.Namespace,
				})
			}

			resource := cyclonedx.Component{
				BOMRef:     id(res),
				Type:       cyclonedx.ComponentTypeApplication, // TODO: this is not perfect but we don't have a better option
				Name:       res.Name,
				Version:    res.APIVersion,
				Properties: &properties,
			}

			components = append(components, resource)
		}
	}

	clusterDependenciesArr := make([]string, 0)
	for _, dep := range clusterDependencies {
		clusterDependenciesArr = append(clusterDependenciesArr, dep)
	}
	slices.Sort(clusterDependenciesArr)

	dependencies = append(dependencies,
		cyclonedx.Dependency{
			Ref:          clusterComponent.BOMRef,
			Dependencies: &clusterDependenciesArr,
		},
	)

	cdxBOM.Components = &components
	cdxBOM.Dependencies = &dependencies

	return cdxBOM
}

func id(obj interface{}) string {
	f, err := hashstructure.Hash(obj, hashstructure.FormatV2, &hashstructure.HashOptions{
		ZeroNil:      true,
		SlicesAsSets: true,
		Hasher:       fnv.New64(),
	})

	// this should never happen, but if it does, we don't want to crash - use empty string
	if err != nil {
		fmt.Printf("failed to hash object: %v", err)
		return ""
	}

	return fmt.Sprintf("%016x", f)
}
