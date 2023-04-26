package kube

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/ksoclabs/kbom/internal/model"
)

type K8sClient interface {
	Metadata(ctx context.Context) (string, string, error)
	Location(ctx context.Context) (*model.Location, error)
	AllImages(ctx context.Context) ([]model.Image, error)
	AllNodes(ctx context.Context, full bool) ([]model.Node, error)
	AllResources(ctx context.Context, full bool) (map[string]model.ResourceList, error)
}

func NewClient() (K8sClient, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		kubeConfigPath := os.Getenv("KUBECONFIG")
		if kubeConfigPath == "" {
			kubeConfigPath = os.Getenv("HOME") + "/.kube/config"
		}
		cfg, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get kubernetes out-cluster client: %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("can not create kubernetes client: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("can not create kubernetes dynamic client: %w", err)
	}

	rest.SetDefaultWarningHandler(rest.NoWarnings{})

	return &k8sDB{
		cfg:           cfg,
		client:        clientset,
		dynamicClient: dynamicClient,
	}, nil
}

type k8sDB struct {
	cfg           *rest.Config
	client        kubernetes.Interface
	dynamicClient dynamic.Interface
}

func (k *k8sDB) Location(ctx context.Context) (*model.Location, error) {
	// fetch first node
	node, err := k.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{Limit: 1})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %v", err)
	}

	if len(node.Items) == 0 {
		return nil, fmt.Errorf("no node found")
	}

	// get location from node labels
	return &model.Location{
		Location: getCloudLocation(node.Items[0].Labels),
		Region:   getLabelValue(node.Items[0].Labels, "topology.kubernetes.io/region"),
		Zone:     getLabelValue(node.Items[0].Labels, "topology.kubernetes.io/zone"),
	}, nil
}

// AllNodes returns all nodes in the cluster
func (k *k8sDB) AllNodes(ctx context.Context, full bool) ([]model.Node, error) {
	nodes, err := k.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %v", err)
	}

	modelNodes := make([]model.Node, 0)
	for i := range nodes.Items {
		var labels, annotations map[string]string
		if full {
			labels = nodes.Items[i].Labels
			annotations = nodes.Items[i].Annotations
		}

		modelNodes = append(modelNodes, model.Node{
			Name:     nodes.Items[i].Name,
			OsImage:  nodes.Items[i].Status.NodeInfo.OSImage,
			Hostname: getLabelValue(nodes.Items[i].Labels, "kubernetes.io/hostname"),
			Type:     getLabelValue(nodes.Items[i].Labels, "node.kubernetes.io/instance-type"),
			Capacity: &model.Capacity{
				CPU:              nodes.Items[i].Status.Capacity.Cpu().String(),
				Memory:           nodes.Items[i].Status.Capacity.Memory().String(),
				EphemeralStorage: nodes.Items[i].Status.Capacity.StorageEphemeral().String(),
				Pods:             nodes.Items[i].Status.Capacity.Pods().String(),
			},
			Labels:                  labels,
			Annotations:             annotations,
			MachineID:               nodes.Items[i].Status.NodeInfo.MachineID,
			Architecture:            nodes.Items[i].Status.NodeInfo.Architecture,
			KernelVersion:           nodes.Items[i].Status.NodeInfo.KernelVersion,
			ContainerRuntimeVersion: nodes.Items[i].Status.NodeInfo.ContainerRuntimeVersion,
			BootID:                  nodes.Items[i].Status.NodeInfo.BootID,
			KubeProxyVersion:        nodes.Items[i].Status.NodeInfo.KubeProxyVersion,
			KubeletVersion:          nodes.Items[i].Status.NodeInfo.KubeletVersion,
			OperatingSystem:         nodes.Items[i].Status.NodeInfo.OperatingSystem,
		})
	}

	return modelNodes, nil
}

func (k *k8sDB) AllImages(ctx context.Context) ([]model.Image, error) {
	namespaces, err := k.client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %v", err)
	}

	images := make(map[string]model.Image)
	for i := range namespaces.Items {
		pods, err := k.client.CoreV1().Pods(namespaces.Items[i].Name).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to list pods: %w", err)
		}

		log.Debug().Str("namespace", namespaces.Items[i].Name).Int("count", len(pods.Items)).Msg("Found pods in namespace")

		for j := range pods.Items {
			pod := pods.Items[j]

			for k := range pod.Spec.InitContainers {
				img, err := containerToImage(pod.Spec.InitContainers[k].Image, pod.Spec.InitContainers[k].Name, pod.Status.InitContainerStatuses)
				if err != nil {
					return nil, err
				}

				images[img.Name] = *img
			}

			for k := range pod.Spec.Containers {
				img, err := containerToImage(pod.Spec.Containers[k].Image, pod.Spec.Containers[k].Name, pod.Status.ContainerStatuses)
				if err != nil {
					return nil, err
				}

				images[img.Name] = *img
			}

			for k := range pod.Spec.EphemeralContainers {
				img, err := containerToImage(pod.Spec.EphemeralContainers[k].Image,
					pod.Spec.EphemeralContainers[k].Name, pod.Status.EphemeralContainerStatuses)
				if err != nil {
					return nil, err
				}

				images[img.Name] = *img
			}
		}
	}

	toReturn := make([]model.Image, 0)
	for _, v := range images {
		toReturn = append(toReturn, v)
	}

	return toReturn, nil
}

func containerToImage(img, imgName string, statuses []v1.ContainerStatus) (*model.Image, error) {
	if img == "" {
		return nil, fmt.Errorf("container %s has no image", img)
	}

	res := &model.Image{
		Name: img,
	}

	if strings.Contains(img, "@") {
		res.Digest = strings.Split(img, "@")[1]
		return res, nil
	}

	// search in statuses for ImageID to get digest
	for i := range statuses {
		if imgName == statuses[i].Name {
			if statuses[i].State.Running == nil && statuses[i].State.Terminated == nil {
				break // We can get valid digest only from running or terminated containers
			}
			if strings.Contains(statuses[i].ImageID, "@") {
				res.Digest = strings.Split(statuses[i].ImageID, "@")[1]
			}
			break
		}
	}

	return res, nil
}

// Metadata returns the kubernetes version
func (k *k8sDB) Metadata(ctx context.Context) (k8sVersion, caDigest string, err error) {
	if _, err := rest.InClusterConfig(); err != nil {
		hash := sha256.Sum256(k.cfg.CAData)
		caDigest = fmt.Sprintf("%x", hash[:])
	} else {
		caConfigMap, err := k.client.CoreV1().ConfigMaps("kube-system").Get(ctx, "kube-root-ca.crt", metav1.GetOptions{})
		if err != nil {
			log.Debug().Err(err).Msg("failed to get kube-root-ca.crt")
		} else {
			caCert, ok := caConfigMap.Data["ca.crt"]
			if !ok {
				return "", "", fmt.Errorf("can't find 'ca.crt' in configMap 'kube-root-ca.crt'")
			}

			caDigest = fmt.Sprintf("%x", sha256.Sum256([]byte(caCert)))
		}
	}

	version, err := k.client.Discovery().ServerVersion()
	if err != nil {
		return caDigest, "", fmt.Errorf("error getting k8s version: %w", err)
	}

	ver := strings.Trim(version.GitVersion, "v")

	sVer, err := semver.NewVersion(ver)
	if err != nil {
		return caDigest, "", fmt.Errorf("error parsing k8s version: %w", err)
	}

	ver = fmt.Sprintf("%d.%d.%d", sVer.Major(), sVer.Minor(), sVer.Patch())

	return ver, caDigest, nil
}

func (k *k8sDB) AllResources(ctx context.Context, full bool) (map[string]model.ResourceList, error) {
	apiResourceList, err := k.client.Discovery().ServerPreferredResources()
	if err != nil {
		return nil, fmt.Errorf("failed to get api groups: %w", err)
	}

	resourceMap := make(map[string]model.ResourceList)
	for _, apiResource := range apiResourceList {
		gv, err := schema.ParseGroupVersion(apiResource.GroupVersion)
		if err != nil {
			return nil, fmt.Errorf("failed to parse group version: %w", err)
		}

		for i := range apiResource.APIResources {
			res := apiResource.APIResources[i]
			gvr := schema.GroupVersionResource{
				Group:    gv.Group,
				Version:  gv.Version,
				Resource: res.Name,
			}

			resourceList, err := k.dynamicClient.Resource(gvr).List(ctx, metav1.ListOptions{})
			if err != nil {
				log.Debug().Err(err).Interface("gvr", gvr).Msg("Failed to list resources")
				continue
			}

			log.Debug().Interface("gvr", gvr).Int("count", len(resourceList.Items)).Msg("Found resources")

			if len(resourceList.Items) > 0 {
				resourceMap[gvr.String()] = model.ResourceList{
					Kind:           resourceList.Items[0].GetKind(),
					APIVersion:     gvr.GroupVersion().String(),
					Namespaced:     res.Namespaced,
					ResourcesCount: len(resourceList.Items),
					Resources:      make([]model.Resource, 0),
				}
			}

			if full {
				for _, item := range resourceList.Items {
					res := model.Resource{
						Name:      item.GetName(),
						Namespace: item.GetNamespace(),
					}

					val := resourceMap[gvr.String()]
					val.Resources = append(val.Resources, res)
					resourceMap[gvr.String()] = val
				}
			}
		}
	}

	return resourceMap, nil
}

func getLabelValue(labels map[string]string, key string) string {
	for k, v := range labels {
		if k == key {
			return v
		}
	}

	return ""
}

func getCloudLocation(labels map[string]string) string {
	if labels == nil {
		return "unknown"
	}

	if _, ok := labels["k8s.io/cloud-provider-aws"]; ok {
		return "aws"
	}

	if _, ok := labels["topology.gke.io/zone"]; ok {
		return "gcloud"
	}

	if _, ok := labels["kubernetes.azure.com/cluster"]; ok {
		return "azure"
	}

	return "unknown"
}
