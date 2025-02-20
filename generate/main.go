package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"text/template"

	"github.com/hashicorp/hcl/v2/hclsimple"
)

func main() {
	resourceName := flag.String("name", "", "")
	flag.Parse()

	if *resourceName == "" {
		fmt.Fprintf(os.Stderr, "name is required\n")
		os.Exit(1)
	}

	templates := map[string]string{
		"ephemeral":     "generate/templates/ephemeral/ephemeral.tmpl",
		"ephemeraltest": "generate/templates/ephemeral/ephemeraltest.tmpl",
		//"websitedoc":    "generate/templates/ephemeral/websitedoc.tmpl",
	}

	var cfg GeneratorConfig
	err := hclsimple.DecodeFile("generate/config.hcl", nil, &cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading config: %v\n", err)
		os.Exit(1)
	}

	var model EphemeralResource
	for _, resource := range cfg.EphemeralResources {
		if resource.Name == *resourceName {
			model = resource
		}
	}

	if model.Name == "" {
		fmt.Fprintf(os.Stderr, "resource %s not found in generator config\n", *resourceName)
		os.Exit(1)
	}

	for name, path := range templates {
		outputPath := "internal/provider/ephemeral_resource_" + model.Name

		if name == "ephemeraltest" {
			outputPath += "_test"
		}

		outputPath += ".go"

		if name == "websitedoc" {
			outputPath = "website/docs/e/{{.Name}}.md"
		}

		err := generateFromTemplate(model, path, outputPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error generating %s: %v\n", name, err)
			os.Exit(1)
		}
	}
}

func generateFromTemplate(model EphemeralResource, templatePath string, outputPath string) error {
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, model)
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, buf.Bytes(), 0o600)
}

type GeneratorConfig struct {
	EphemeralResources []EphemeralResource `hcl:"ephemeral_resource,block"`
}

type EphemeralResource struct {
	Name        string  `hcl:"name,label"`
	UpcaseName  string  `hcl:"upcase_name"`
	Description string  `hcl:"description"`
	Fields      []Field `hcl:"field,block"`
}

type Field struct {
	Name              string `hcl:"name,label"`
	UpcaseName        string `hcl:"upcase_name"`
	Description       string `hcl:"description"`
	Type              string `hcl:"type"`
	Required          bool   `hcl:"required,optional"`
	Computed          bool   `hcl:"computed,optional"`
	ModelAttr         string `hcl:"model_attr,optional"`
	SuppressTestCheck bool   `hcl:"suppress_test_check,optional"`
}
