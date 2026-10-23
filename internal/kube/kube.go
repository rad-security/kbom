package kube

import (
	"context"
	"crypto/sha256"
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/Masterminds/semver"
	"github.com/distribution/reference"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/metadata"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/rad-security/kbom/internal/model"
)

const (
	// client-go defaults (5 QPS, burst 10) throttle the hundreds of list calls needed for a KBOM
	clientQPS   = 50
	clientBurst = 100

	listPageSize = 500
	listWorkers  = 8
)

// metadataOnlyGroups are built-in API groups whose resources have no spec.version,
// so they can be listed without fetching full objects (e.g. secret data).
var metadataOnlyGroups = map[string]bool{
	"":                             true,
	"admissionregistration.k8s.io": true,
	"apiextensions.k8s.io":         true,
	"apps":                         true,
	"autoscaling":                  true,
	"batch":                        true,
	"certificates.k8s.io":          true,
	"coordination.k8s.io":          true,
	"discovery.k8s.io":             true,
	"events.k8s.io":                true,
	"flowcontrol.apiserver.k8s.io": true,
	"networking.k8s.io":            true,
	"node.k8s.io":                  true,
	"policy":                       true,
	"rbac.authorization.k8s.io":    true,
	"resource.k8s.io":              true,
	"scheduling.k8s.io":            true,
	"storage.k8s.io":               true,
}

// ProgressFunc is called after each resource type is listed.
type ProgressFunc func(done, total int, resource string)

type K8sClient interface {
	ClusterName(ctx context.Context) (string, error)
	Metadata(ctx context.Context) (string, string, error)
	Location(ctx context.Context) (*model.Location, error)
	AllImages(ctx context.Context) ([]model.Image, error)
	AllNodes(ctx context.Context, full bool) ([]model.Node, error)
	AllResources(ctx context.Context, full bool, progress ProgressFunc) (map[string]model.ResourceList, error)
}

func NewClient(k8sContext string) (K8sClient, error) {
	currentK8sContext := k8sContext

	cfg, err := rest.InClusterConfig()
	if err != nil {
		// Default loading rules handle KUBECONFIG with multiple files and fall back to ~/.kube/config
		clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			clientcmd.NewDefaultClientConfigLoadingRules(),
			&clientcmd.ConfigOverrides{
				CurrentContext: k8sContext,
			})

		rawConfig, err := clientConfig.RawConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to get kubernetes out-cluster client: %w", err)
		}

		if k8sContext == "" {
			currentK8sContext = rawConfig.CurrentContext
		}

		cfg, err = clientConfig.ClientConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to get kubernetes out-cluster client: %w", err)
		}
	}

	cfg.QPS = clientQPS
	cfg.Burst = clientBurst

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("can not create kubernetes client: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("can not create kubernetes dynamic client: %w", err)
	}

	metadataClient, err := metadata.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("can not create kubernetes metadata client: %w", err)
	}

	rest.SetDefaultWarningHandler(rest.NoWarnings{})

	return &k8sDB{
		k8sContext:     currentK8sContext,
		cfg:            cfg,
		client:         clientset,
		dynamicClient:  dynamicClient,
		metadataClient: metadataClient,
	}, nil
}

type k8sDB struct {
	k8sContext     string
	cfg            *rest.Config
	client         kubernetes.Interface
	dynamicClient  dynamic.Interface
	metadataClient metadata.Interface
}

func (k *k8sDB) ClusterName(ctx context.Context) (string, error) {
	// TODO: find a better way to get cluster name, but for now use cluster context name
	return k.k8sContext, nil
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
		Name:   getCloudName(node.Items[0].Labels),
		Region: getLabelValue(node.Items[0].Labels, "topology.kubernetes.io/region"),
		Zone:   getLabelValue(node.Items[0].Labels, "topology.kubernetes.io/zone"),
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
			Allocatable: &model.Capacity{
				CPU:              nodes.Items[i].Status.Allocatable.Cpu().String(),
				Memory:           nodes.Items[i].Status.Allocatable.Memory().String(),
				EphemeralStorage: nodes.Items[i].Status.Allocatable.StorageEphemeral().String(),
				Pods:             nodes.Items[i].Status.Allocatable.Pods().String(),
			},
			Labels:                  labels,
			Annotations:             annotations,
			MachineID:               nodes.Items[i].Status.NodeInfo.MachineID,
			Architecture:            nodes.Items[i].Status.NodeInfo.Architecture,
			KernelVersion:           nodes.Items[i].Status.NodeInfo.KernelVersion,
			ContainerRuntimeVersion: nodes.Items[i].Status.NodeInfo.ContainerRuntimeVersion,
			BootID:                  nodes.Items[i].Status.NodeInfo.BootID,
			KubeProxyVersion:        nodes.Items[i].Status.NodeInfo.KubeProxyVersion, //nolint:staticcheck // part of KBOM schema
			KubeletVersion:          nodes.Items[i].Status.NodeInfo.KubeletVersion,
			OperatingSystem:         nodes.Items[i].Status.NodeInfo.OperatingSystem,
		})
	}

	return modelNodes, nil
}

func (k *k8sDB) AllImages(ctx context.Context) ([]model.Image, error) {
	pods, err := listPages(ctx, func(opts metav1.ListOptions) ([]v1.Pod, string, error) {
		list, err := k.client.CoreV1().Pods(metav1.NamespaceAll).List(ctx, opts)
		if err != nil {
			return nil, "", err
		}
		return list.Items, list.Continue, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	log.Debug().Int("count", len(pods)).Msg("Found pods")

	images := make(map[string]model.Image)
	for j := range pods {
		pod := pods[j]
		namespace := pod.Namespace

		for k := range pod.Spec.InitContainers {
			img, err := containerToImage(pod.Spec.InitContainers[k].Image,
				pod.Spec.InitContainers[k].Name, pod.Status.InitContainerStatuses, namespace)
			if err != nil {
				return nil, err
			}

			images[img.FullName] = *img
		}

		for k := range pod.Spec.Containers {
			img, err := containerToImage(pod.Spec.Containers[k].Image, pod.Spec.Containers[k].Name, pod.Status.ContainerStatuses, namespace)
			if err != nil {
				return nil, err
			}

			images[img.FullName] = *img
		}

		for k := range pod.Spec.EphemeralContainers {
			img, err := containerToImage(pod.Spec.EphemeralContainers[k].Image,
				pod.Spec.EphemeralContainers[k].Name, pod.Status.EphemeralContainerStatuses, namespace)
			if err != nil {
				return nil, err
			}

			images[img.FullName] = *img
		}
	}

	toReturn := make([]model.Image, 0)
	for _, v := range images {
		toReturn = append(toReturn, v)
	}

	return toReturn, nil
}

func containerToImage(img, imgName string, statuses []v1.ContainerStatus, namespace string) (*model.Image, error) {
	if img == "" {
		return nil, fmt.Errorf("container %s has no image", img)
	}

	named, err := reference.ParseNormalizedNamed(img)
	if err != nil {
		return nil, err
	}

	controlPlane := namespace == "kube-system"

	res := &model.Image{
		FullName:     img,
		ControlPlane: controlPlane,
	}

	res.Name = named.Name()
	tagged, ok := named.(reference.Tagged)
	if ok {
		res.Version = tagged.Tag()
	}

	digested, ok := named.(reference.Digested)
	if ok {
		res.Digest = digested.Digest().String()
	}

	// search in statuses for ImageID to get digest
	for i := range statuses {
		if imgName == statuses[i].Name {
			if statuses[i].State.Running == nil && statuses[i].State.Terminated == nil {
				break // We can get valid digest only from running or terminated containers
			}
			if strings.Contains(statuses[i].ImageID, "@") {
				res.Digest = strings.Split(statuses[i].ImageID, "@")[1]
			} else if strings.HasPrefix(statuses[i].ImageID, "sha256:") {
				res.Digest = statuses[i].ImageID
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

type listTarget struct {
	gvr        schema.GroupVersionResource
	kind       string
	namespaced bool
}

func (k *k8sDB) AllResources(ctx context.Context, full bool, progress ProgressFunc) (map[string]model.ResourceList, error) {
	apiResourceList, err := k.client.Discovery().ServerPreferredResources()
	if err != nil {
		return nil, fmt.Errorf("failed to get api groups: %w", err)
	}

	var targets []listTarget
	for _, apiResource := range apiResourceList {
		gv, err := schema.ParseGroupVersion(apiResource.GroupVersion)
		if err != nil {
			return nil, fmt.Errorf("failed to parse group version: %w", err)
		}

		for i := range apiResource.APIResources {
			res := apiResource.APIResources[i]
			if !slices.Contains(res.Verbs, "list") {
				continue
			}

			targets = append(targets, listTarget{gvr: gv.WithResource(res.Name), kind: res.Kind, namespaced: res.Namespaced})
		}
	}

	var (
		mu          sync.Mutex
		wg          sync.WaitGroup
		done        int
		resourceMap = make(map[string]model.ResourceList)
		workers     = make(chan struct{}, listWorkers)
	)

	for _, target := range targets {
		wg.Go(func() {
			workers <- struct{}{}
			defer func() { <-workers }()

			resources, err := k.listResources(ctx, target.gvr, full)

			mu.Lock()
			defer mu.Unlock()

			done++
			if progress != nil {
				progress(done, len(targets), target.gvr.Resource)
			}

			if err != nil {
				log.Debug().Err(err).Interface("gvr", target.gvr).Msg("Failed to list resources")
				return
			}

			log.Debug().Interface("gvr", target.gvr).Int("count", len(resources)).Msg("Found resources")

			if len(resources) == 0 {
				return
			}

			resourceList := model.ResourceList{
				Kind:           target.kind,
				APIVersion:     target.gvr.GroupVersion().String(),
				Namespaced:     target.namespaced,
				ResourcesCount: len(resources),
				Resources:      make([]model.Resource, 0),
			}
			if full {
				resourceList.Resources = resources
			}

			resourceMap[target.gvr.String()] = resourceList
		})
	}

	wg.Wait()

	return resourceMap, nil
}

// listResources lists only object metadata when possible and falls back to full objects
// when spec.version is needed or the API server does not support metadata-only lists.
func (k *k8sDB) listResources(ctx context.Context, gvr schema.GroupVersionResource, full bool) ([]model.Resource, error) {
	if !full || metadataOnlyGroups[gvr.Group] {
		resources, err := k.listMetadata(ctx, gvr)
		if err == nil {
			return resources, nil
		}

		log.Debug().Err(err).Interface("gvr", gvr).Msg("Failed to list metadata, falling back to full objects")
	}

	items, err := listPages(ctx, func(opts metav1.ListOptions) ([]unstructured.Unstructured, string, error) {
		list, err := k.dynamicClient.Resource(gvr).List(ctx, opts)
		if err != nil {
			return nil, "", err
		}
		return list.Items, list.GetContinue(), nil
	})
	if err != nil {
		return nil, err
	}

	resources := make([]model.Resource, 0, len(items))
	for _, item := range items {
		res := model.Resource{
			Name:                 item.GetName(),
			Namespace:            item.GetNamespace(),
			AdditionalProperties: map[string]string{},
		}
		if version, ok := getVersion(item); ok {
			res.AdditionalProperties["version"] = version
		}
		resources = append(resources, res)
	}

	return resources, nil
}

func (k *k8sDB) listMetadata(ctx context.Context, gvr schema.GroupVersionResource) ([]model.Resource, error) {
	items, err := listPages(ctx, func(opts metav1.ListOptions) ([]metav1.PartialObjectMetadata, string, error) {
		list, err := k.metadataClient.Resource(gvr).List(ctx, opts)
		if err != nil {
			return nil, "", err
		}
		return list.Items, list.Continue, nil
	})
	if err != nil {
		return nil, err
	}

	resources := make([]model.Resource, 0, len(items))
	for i := range items {
		resources = append(resources, model.Resource{
			Name:                 items[i].Name,
			Namespace:            items[i].Namespace,
			AdditionalProperties: map[string]string{},
		})
	}

	return resources, nil
}

// listPages calls list until all pages are fetched.
func listPages[T any](ctx context.Context, list func(metav1.ListOptions) ([]T, string, error)) ([]T, error) {
	var all []T

	opts := metav1.ListOptions{Limit: listPageSize}
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		items, next, err := list(opts)
		if err != nil {
			return nil, err
		}

		all = append(all, items...)
		if next == "" {
			return all, nil
		}

		opts.Continue = next
	}
}

func getVersion(item unstructured.Unstructured) (version string, ok bool) {
	obj := item.Object
	if obj == nil {
		return "", false
	}

	spec, ok := obj["spec"].(map[string]interface{})
	if !ok {
		return "", false
	}

	version, ok = spec["version"].(string)
	return
}

func getLabelValue(labels map[string]string, key string) string {
	for k, v := range labels {
		if k == key {
			return v
		}
	}

	return ""
}

func getCloudName(labels map[string]string) string {
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
