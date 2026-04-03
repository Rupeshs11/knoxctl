package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/knoxctl/pkg/kube"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	yamlserializer "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
)

var applyFile string

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply a configuration to a resource from a YAML file",
	Long:  `Apply a resource configuration from a YAML/JSON file. Creates the resource if it doesn't exist, or updates it if it already exists.`,
	Example: `  # Apply a deployment from a YAML file
  knoxctl apply -f deployment.yaml

  # Apply a service from a YAML file
  knoxctl apply -f service.yaml

  # Apply all YAML files in the current directory
  knoxctl apply -f all`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if applyFile == "" {
			return fmt.Errorf("error: must specify a file with -f flag\nUsage: knoxctl apply -f <filename>\n       knoxctl apply -f all")
		}

		if strings.ToLower(applyFile) == "all" {
			return applyAllFiles()
		}

		return applySingleFile(applyFile)
	},
}

func applySingleFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file %s: %w", filePath, err)
	}

	absPath, _ := filepath.Abs(filePath)

	decUnstructured := yamlserializer.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	obj := &unstructured.Unstructured{}
	_, gvk, err := decUnstructured.Decode(data, nil, obj)
	if err != nil {
		return fmt.Errorf("error decoding YAML from %s: %w", absPath, err)
	}

	config, err := kube.GetRestConfig(kubeconfig)
	if err != nil {
		return fmt.Errorf("error creating kubernetes config: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("error creating dynamic client: %w", err)
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return fmt.Errorf("error creating discovery client: %w", err)
	}

	apiGroupResources, err := restmapper.GetAPIGroupResources(discoveryClient)
	if err != nil {
		return fmt.Errorf("error getting API group resources: %w", err)
	}

	mapper := restmapper.NewDiscoveryRESTMapper(apiGroupResources)
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return fmt.Errorf("error mapping resource: %w", err)
	}

	ns := obj.GetNamespace()
	if ns == "" {
		ns = namespace
	}

	var resourceClient dynamic.ResourceInterface
	if mapping.Scope.Name() == "namespace" {
		resourceClient = dynamicClient.Resource(mapping.Resource).Namespace(ns)
	} else {
		resourceClient = dynamicClient.Resource(mapping.Resource)
	}

	jsonData, err := obj.MarshalJSON()
	if err != nil {
		return fmt.Errorf("error marshaling resource: %w", err)
	}

	_, err = resourceClient.Patch(context.TODO(), obj.GetName(), types.ApplyPatchType, jsonData, metav1.PatchOptions{
		FieldManager: "knoxctl",
	})
	if err != nil {
		return fmt.Errorf("error applying resource: %w", err)
	}

	fmt.Printf("%s %s/%s\n", colorGreen("✔ configured"), gvk.Kind, colorBold(obj.GetName()))
	return nil
}

func applyAllFiles() error {
	files, err := findYAMLFiles(".")
	if err != nil {
		return fmt.Errorf("error scanning directory: %w", err)
	}

	if len(files) == 0 {
		fmt.Println("No YAML files found in current directory.")
		return nil
	}

	fmt.Printf("Applying %d YAML file(s)...\n\n", len(files))

	var errors []string
	applied := 0

	for _, file := range files {
		err := applySingleFile(file)
		if err != nil {
			errors = append(errors, fmt.Sprintf("  %s %s: %v", colorRed("✗"), file, err))
		} else {
			applied++
		}
	}

	if len(errors) > 0 {
		fmt.Printf("\n%d file(s) applied, %d error(s):\n", applied, len(errors))
		for _, e := range errors {
			fmt.Println(e)
		}
	} else {
		fmt.Printf("\nAll %d file(s) applied %s\n", applied, colorGreen("successfully."))
	}

	return nil
}

func findYAMLFiles(dir string) ([]string, error) {
	var files []string
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := strings.ToLower(entry.Name())
		if strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}
	return files, nil
}

func init() {
	applyCmd.Flags().StringVarP(&applyFile, "file", "f", "", "Path to the YAML file to apply (use 'all' for all YAML files)")
	rootCmd.AddCommand(applyCmd)
}

