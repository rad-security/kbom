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
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/ksoclabs/kbom/internal/config"
	"github.com/ksoclabs/kbom/internal/kube"
	"github.com/ksoclabs/kbom/internal/model"
	"github.com/ksoclabs/kbom/internal/utils"
)

const (
	KSOCCompany = "KSOC Labs"
	BOMFormat   = "ksoc"
	SpecVersion = "0.2"

	StdOutput  = "stdout"
	FileOutput = "file"
)

var (
	short   bool
	output  string
	format  string
	outPath string

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
	k8sVersion, caCertDigest, err := k8sClient.Metadata(ctx)
	if err != nil {
		return err
	}

	clusterName, err := k8sClient.ClusterName(ctx)
	if err != nil {
		return err
	}

	full := !short
	nodes, err := k8sClient.AllNodes(ctx, full)
	if err != nil {
		return err
	}

	loc, err := k8sClient.Location(ctx)
	if err != nil {
		return err
	}

	allImages, err := k8sClient.AllImages(ctx)
	if err != nil {
		return err
	}

	resources, err := k8sClient.AllResources(ctx, full)
	if err != nil {
		return err
	}

	kbom := model.KBOM{
		ID:          kbomID,
		BOMFormat:   BOMFormat,
		SpecVersion: SpecVersion,
		GeneratedAt: generatedAt,
		GeneratedBy: model.Tool{
			Vendor:     KSOCCompany,
			BuildTime:  config.BuildTime,
			Name:       config.AppName,
			Version:    config.AppVersion,
			Commit:     config.LastCommitHash,
			CommitTime: config.LastCommitTime,
		},
		Cluster: model.Cluster{
			Name:         clusterName,
			Location:     loc,
			CNIVersion:   "", // TODO: get CNI version
			K8sVersion:   k8sVersion,
			CACertDigest: caCertDigest,
			NodesCount:   len(nodes),
			Nodes:        nodes,
			Components: model.Components{
				Images:    allImages,
				Resources: resources,
			},
		},
	}

	if err := printKBOM(&kbom, parsedFormat); err != nil {
		return err
	}

	return nil
}

func printKBOM(kbom *model.KBOM, f Format) error {
	writer, err := getWriter(kbom, f)
	if err != nil {
		return err
	}
	defer writer.Close()

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

func getWriter(kbom *model.KBOM, format Format) (io.WriteCloser, error) {
	switch output {
	case StdOutput:
		return out, nil
	case FileOutput:
		formattedTime := kbom.GeneratedAt.Format("2006-01-02-15-04-05")
		key := kbom.ID[:8]
		if len(kbom.Cluster.CACertDigest) > 8 {
			key = kbom.Cluster.CACertDigest[:8]
		}

		f, err := os.Create(path.Join(outPath, fmt.Sprintf("kbom-%s-%s.%s", key, formattedTime, format.FileExtension)))
		if err != nil {
			return nil, err
		}

		return f, nil
	default:
		return nil, fmt.Errorf("output %q is not supported", output)
	}
}
