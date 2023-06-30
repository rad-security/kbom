package cmd

import (
	"fmt"
	"hash/fnv"
	"time"

	"github.com/CycloneDX/cyclonedx-go"
	"github.com/google/uuid"
	"github.com/ksoclabs/kbom/internal/model"
	"github.com/mitchellh/hashstructure/v2"
)

const (
	CdxPrefix  = "cdx:"
	KSOCPrefix = "ksoc:kbom:"

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
		Component: &cyclonedx.Component{
			BOMRef: id(kbom.GeneratedBy),
			Type:   cyclonedx.ComponentTypeApplication,
			Name:   kbom.GeneratedBy.Name,
			Hashes: &[]cyclonedx.Hash{
				{
					Algorithm: cyclonedx.HashAlgoSHA256,
					Value:     kbom.GeneratedBy.Commit,
				},
			},
			Version: kbom.GeneratedBy.Version,
		},
	}

	components := []cyclonedx.Component{}
	clusterProperties := []cyclonedx.Property{
		{
			Name:  CdxPrefix + "k8s:component:type",
			Value: ClusterType,
		},
		{
			Name:  CdxPrefix + "k8s:component:name",
			Value: kbom.Cluster.Name,
		},
		{
			Name:  KSOCPrefix + "k8s:cluster:nodes",
			Value: fmt.Sprintf("%d", kbom.Cluster.NodesCount),
		},
	}

	if kbom.Cluster.Location.Name != "" && kbom.Cluster.Location.Name != "unknown" {
		clusterProperties = append(clusterProperties, cyclonedx.Property{
			Name:  KSOCPrefix + "k8s:cluster:location:name",
			Value: kbom.Cluster.Location.Name,
		})
	}

	if kbom.Cluster.Location.Region != "" {
		clusterProperties = append(clusterProperties, cyclonedx.Property{
			Name:  KSOCPrefix + "k8s:cluster:location:region",
			Value: kbom.Cluster.Location.Region,
		})
	}

	if kbom.Cluster.Location.Zone != "" {
		clusterProperties = append(clusterProperties, cyclonedx.Property{
			Name:  KSOCPrefix + "k8s:cluster:location:zone",
			Value: kbom.Cluster.Location.Zone,
		})
	}

	clusterComponent := cyclonedx.Component{
		BOMRef:     id(kbom.Cluster),
		Type:       cyclonedx.ComponentTypePlatform,
		Name:       "cluster",
		Version:    kbom.Cluster.K8sVersion,
		Properties: &clusterProperties,
	}

	components = append(components, clusterComponent)

	for i := range kbom.Cluster.Nodes {
		n := kbom.Cluster.Nodes[i]
		components = append(components, cyclonedx.Component{
			BOMRef: id(n),
			Type:   cyclonedx.ComponentTypePlatform,
			Name:   n.Name,
			Properties: &[]cyclonedx.Property{
				{
					Name:  CdxPrefix + "k8s:component:type",
					Value: NodeType,
				},
				{
					Name:  CdxPrefix + "k8s:component:name",
					Value: n.Name,
				},
				{
					Name:  KSOCPrefix + "k8s:node:osImage",
					Value: n.OsImage,
				},
				{
					Name:  KSOCPrefix + "k8s:node:arch",
					Value: n.Architecture,
				},
				{
					Name:  KSOCPrefix + "k8s:node:kernel",
					Value: n.KernelVersion,
				},
				{
					Name:  KSOCPrefix + "k8s:node:bootId",
					Value: n.BootID,
				},
				{
					Name:  KSOCPrefix + "k8s:node:type",
					Value: n.Type,
				},
				{
					Name:  KSOCPrefix + "k8s:node:operatingSystem",
					Value: n.OperatingSystem,
				},
				{
					Name:  KSOCPrefix + "k8s:node:machineId",
					Value: n.MachineID,
				},
				{
					Name:  KSOCPrefix + "k8s:node:hostname",
					Value: n.Hostname,
				},
				{
					Name:  KSOCPrefix + "k8s:node:containerRuntimeVersion",
					Value: n.ContainerRuntimeVersion,
				},
				{
					Name:  KSOCPrefix + "k8s:node:kubeletVersion",
					Value: n.KubeletVersion,
				},
				{
					Name:  KSOCPrefix + "k8s:node:kubeProxyVersion",
					Value: n.KubeProxyVersion,
				},
				{
					Name:  KSOCPrefix + "k8s:node:capacity:cpu",
					Value: n.Capacity.CPU,
				},
				{
					Name:  KSOCPrefix + "k8s:node:capacity:memory",
					Value: n.Capacity.Memory,
				},
				{
					Name:  KSOCPrefix + "k8s:node:capacity:pods",
					Value: n.Capacity.Pods,
				},
				{
					Name:  KSOCPrefix + "k8s:node:capacity:ephemeralStorage",
					Value: n.Capacity.EphemeralStorage,
				},
			},
		})
	}

	for _, img := range kbom.Cluster.Components.Images {
		container := cyclonedx.Component{
			BOMRef:     img.PkgID(),
			Type:       cyclonedx.ComponentTypeContainer,
			Name:       img.Name,
			Version:    img.Digest,
			PackageURL: img.PkgID(),
			Properties: &[]cyclonedx.Property{
				{
					Name:  CdxPrefix + "k8s:component:type",
					Value: ContainerType,
				},
				{
					Name:  CdxPrefix + "k8s:component:name",
					Value: img.Name,
				},
				{
					Name:  KSOCPrefix + "pkg:type",
					Value: "oci",
				},
				{
					Name:  KSOCPrefix + "pkg:name",
					Value: img.Name,
				},
				{
					Name:  KSOCPrefix + "pkg:version",
					Value: img.Version,
				},
				{
					Name:  KSOCPrefix + "pkg:digest",
					Value: img.Digest,
				},
			},
		}

		components = append(components, container)
	}

	for _, resList := range kbom.Cluster.Components.Resources {
		for _, res := range resList.Resources {
			properties := []cyclonedx.Property{
				{
					Name:  CdxPrefix + "k8s:component:type",
					Value: resList.Kind,
				},
				{
					Name:  CdxPrefix + "k8s:component:name",
					Value: res.Name,
				},
				{
					Name:  KSOCPrefix + "k8s:component:apiVersion",
					Value: resList.APIVersion,
				},
			}

			if resList.Namespaced {
				properties = append(properties, cyclonedx.Property{
					Name:  KSOCPrefix + "k8s:component:namespace",
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

	cdxBOM.Components = &components

	// TODO: add relationships and dependencies

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
