package cmd

import "fmt"

type Format struct {
	Name          string
	FileExtension string
}

var JSONFormat = Format{
	Name:          "json",
	FileExtension: "json",
}

var YAMLFormat = Format{
	Name:          "yaml",
	FileExtension: "yaml",
}

var CycloneDXJsonFormat = Format{
	Name:          "cyclonedx-json",
	FileExtension: "json",
}

var CycloneDXXMLFormat = Format{
	Name:          "cyclonedx-xml",
	FileExtension: "xml",
}

func formatNames() []string {
	return []string{
		JSONFormat.Name,
		YAMLFormat.Name,
		CycloneDXJsonFormat.Name,
		CycloneDXXMLFormat.Name,
	}
}

func formatFromName(name string) (Format, error) {
	switch name {
	case JSONFormat.Name:
		return JSONFormat, nil
	case YAMLFormat.Name:
		return YAMLFormat, nil
	case CycloneDXJsonFormat.Name:
		return CycloneDXJsonFormat, nil
	case CycloneDXXMLFormat.Name:
		return CycloneDXXMLFormat, nil
	default:
		return Format{}, fmt.Errorf("format %q is not supported", name)
	}
}
