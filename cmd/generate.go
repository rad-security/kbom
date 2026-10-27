package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/CycloneDX/cyclonedx-go"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/rad-security/kbom/internal/addons"
	"github.com/rad-security/kbom/internal/config"
	"github.com/rad-security/kbom/internal/kube"
	"github.com/rad-security/kbom/internal/model"
	"github.com/rad-security/kbom/internal/progress"
	"github.com/rad-security/kbom/internal/utils"
)

const (
	Company     = "RAD Security"
	BOMFormat   = "rad"
	SpecVersion = "0.3"

	StdOutput  = "stdout"
	FileOutput = "file"
)

var (
	short       bool
	quiet       bool
	clusterName string
	output      string
	format      string
	outPath     string

	progressOut io.Writer = os.Stderr

	generatedAt = time.Now()
	kbomID      = uuid.New().String()
)

var GenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate KBOM for the provided K8s cluster",
	RunE:  runGenerate,
}

func init() {
	GenerateCmd.Flags().BoolVar(&short, "short", false, "Short - only include metadata, nodes, images and resources counters")
	GenerateCmd.Flags().StringVarP(&output, "output", "o", StdOutput, "Output (stdout, file)")
	GenerateCmd.Flags().StringVarP(&format, "format", "f", JSONFormat.Name, fmt.Sprintf("Format (%s)", strings.Join(formatNames(), ", ")))
	GenerateCmd.Flags().StringVarP(&outPath, "out-path", "p", ".", "Path to write KBOM file to. Works only with --output=file")
	GenerateCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Quiet - do not print progress to stderr")
	GenerateCmd.Flags().StringVar(&clusterName, "cluster-name", "", "Cluster name to use instead of the detected one")

	utils.BindFlags(GenerateCmd)
}

func runGenerate(cmd *cobra.Command, _ []string) error {
	k8sClient, err := kube.NewClient(k8sContext)
	if err != nil {
		return err
	}

	return generateKBOM(k8sClient)
}

func generateKBOM(k8sClient kube.K8sClient) error {
	parsedFormat, err := formatFromName(format)
	if err != nil {
		return err
	}

	ctx := context.Background()
	reporter := newReporter()

	step := reporter.Start("Connecting to cluster")
	k8sVersion, caCertDigest, err := k8sClient.Metadata(ctx)
	if err != nil {
		step.Fail()
		return err
	}

	name, id, err := clusterIdentity(ctx, k8sClient)
	if err != nil {
		step.Fail()
		return err
	}
	step.Done(fmt.Sprintf("Connected to %s (Kubernetes %s)", name, k8sVersion))

	full := !short
	var loc *model.Location
	nodes, err := runStep(reporter, "Collecting nodes", func(*progress.Step) ([]model.Node, error) {
		nodes, err := k8sClient.AllNodes(ctx, full)
		if err != nil {
			return nil, err
		}

		loc, err = k8sClient.Location(ctx)
		return nodes, err
	}, func(nodes []model.Node) string {
		return fmt.Sprintf("Found %d nodes", len(nodes))
	})
	if err != nil {
		return err
	}

	allImages, err := runStep(reporter, "Collecting images", func(*progress.Step) ([]model.Image, error) {
		return k8sClient.AllImages(ctx)
	}, func(images []model.Image) string {
		return fmt.Sprintf("Found %d images", len(images))
	})
	if err != nil {
		return err
	}

	resources, err := runStep(reporter, "Listing resources", func(s *progress.Step) (map[string]model.ResourceList, error) {
		return k8sClient.AllResources(ctx, full, func(done, total int, resource string) {
			s.Update(done, total, resource)
		})
	}, resourcesSummary)
	if err != nil {
		return err
	}

	detectedAddons := addons.Detect(allImages)

	kbom := model.KBOM{
		ID:          kbomID,
		BOMFormat:   BOMFormat,
		SpecVersion: SpecVersion,
		GeneratedAt: generatedAt,
		GeneratedBy: model.Tool{
			Vendor:     Company,
			BuildTime:  config.BuildTime,
			Name:       config.AppName,
			Version:    config.AppVersion,
			Commit:     config.LastCommitHash,
			CommitTime: config.LastCommitTime,
		},
		Cluster: model.Cluster{
			Name:         name,
			ID:           id,
			Location:     loc,
			CNIVersion:   addons.CNIVersion(detectedAddons),
			K8sVersion:   k8sVersion,
			CACertDigest: caCertDigest,
			NodesCount:   len(nodes),
			Nodes:        nodes,
			Components: model.Components{
				Images:    allImages,
				Addons:    detectedAddons,
				Resources: resources,
			},
		},
	}

	_, err = runStep(reporter, "Writing KBOM", func(*progress.Step) (string, error) {
		return printKBOM(&kbom, parsedFormat)
	}, func(filename string) string {
		if filename == "" {
			return "KBOM written to stdout"
		}
		return "KBOM written to " + filename
	})

	return err
}

func resourcesSummary(resources map[string]model.ResourceList) string {
	count := 0
	for _, r := range resources {
		count += r.ResourcesCount
	}

	return fmt.Sprintf("Found %d resource types (%d objects)", len(resources), count)
}

// clusterIdentity returns the cluster name and ID. The name is taken from --cluster-name,
// detected from the cluster, or falls back to the ID, so it is never empty for a reachable cluster.
func clusterIdentity(ctx context.Context, k8sClient kube.K8sClient) (name, id string, err error) {
	id, err = k8sClient.ClusterID(ctx)
	if err != nil {
		return "", "", err
	}

	name = clusterName
	if name == "" {
		name, err = k8sClient.ClusterName(ctx)
		if err != nil {
			return "", "", err
		}
	}

	if name == "" {
		name = id
	}

	return name, id, nil
}

func newReporter() *progress.Reporter {
	if quiet {
		return progress.New(nil, false)
	}

	// verbose logs are written to stderr too, so do not redraw lines in place
	interactive := !verbose && isTerminalWriter(progressOut)
	reporter := progress.New(progressOut, interactive)

	if interactive {
		// print logs (e.g. warnings) above the spinner instead of into its line
		log.Logger = log.Logger.Output(zerolog.ConsoleWriter{Out: reporter})
	}

	return reporter
}

// runStep runs fn as a progress step and finishes it with summary on success.
func runStep[T any](r *progress.Reporter, title string, fn func(*progress.Step) (T, error), summary func(T) string) (T, error) {
	step := r.Start(title)

	result, err := fn(step)
	if err != nil {
		step.Fail()
		return result, err
	}

	step.Done(summary(result))

	return result, nil
}

// printKBOM writes the KBOM and returns the file name, or an empty string for stdout.
func printKBOM(kbom *model.KBOM, f Format) (string, error) {
	writer, filename, err := getWriter(kbom, f)
	if err != nil {
		return "", err
	}
	defer writer.Close()

	return filename, encodeKBOM(writer, kbom)
}

func encodeKBOM(writer io.Writer, kbom *model.KBOM) error {
	switch format {
	case JSONFormat.Name:
		enc := json.NewEncoder(writer)
		enc.SetIndent("", "  ")
		return enc.Encode(kbom)
	case YAMLFormat.Name:
		enc := yaml.NewEncoder(writer)
		enc.SetIndent(2)
		return enc.Encode(kbom)
	case CycloneDXJsonFormat.Name:
		cyclonexKbom := transformToCycloneDXBOM(kbom)
		enc := cyclonedx.NewBOMEncoder(writer, cyclonedx.BOMFileFormatJSON)
		enc.SetPretty(true)
		enc.SetEscapeHTML(false)
		return enc.Encode(cyclonexKbom)
	case CycloneDXXMLFormat.Name:
		cyclonexKbom := transformToCycloneDXBOM(kbom)
		enc := cyclonedx.NewBOMEncoder(writer, cyclonedx.BOMFileFormatXML)
		enc.SetPretty(true)
		enc.SetEscapeHTML(false)
		return enc.Encode(cyclonexKbom)
	default:
		return fmt.Errorf("format %q is not supported", format)
	}
}

func getWriter(kbom *model.KBOM, format Format) (io.WriteCloser, string, error) {
	switch output {
	case StdOutput:
		return out, "", nil
	case FileOutput:
		formattedTime := kbom.GeneratedAt.Format("2006-01-02-15-04-05")
		key := kbom.ID[:8]
		if len(kbom.Cluster.CACertDigest) > 8 {
			key = kbom.Cluster.CACertDigest[:8]
		}

		filename := path.Join(outPath, fmt.Sprintf("kbom-%s-%s.%s", key, formattedTime, format.FileExtension))
		f, err := os.Create(filename)
		if err != nil {
			return nil, "", err
		}

		return f, filename, nil
	default:
		return nil, "", fmt.Errorf("output %q is not supported", output)
	}
}
