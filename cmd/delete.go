package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/knoxctl/pkg/kube"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
)

var deleteFile string

var deleteCmd = &cobra.Command{
	Use:   "delete <resource> <name>",
	Short: "Delete resources by name, from a YAML file, or all resources",
	Long:  `Delete Kubernetes resources either by specifying the resource type and name, by providing a YAML file, or delete all pods/deployments/services in a namespace.`,
	Example: `  # Delete a pod by name
  knoxctl delete pod nginx-pod

  # Delete a deployment by name
  knoxctl delete deployment nginx-deploy

  # Delete a service in a specific namespace
  knoxctl delete svc my-service -n production

  # Delete resources from a YAML file
  knoxctl delete -f deployment.yaml

  # Delete ALL pods, deployments, and services in the namespace
  knoxctl delete all
  knoxctl delete all -n production`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if deleteFile != "" {
			return deleteFromFile()
		}

		if len(args) < 1 {
			return fmt.Errorf("error: must specify resource type and name, or use 'all'\nUsage: knoxctl delete <resource> <name>\n       knoxctl delete all\n       knoxctl delete -f <filename>")
		}

		resourceType := strings.ToLower(args[0])

		if resourceType == "all" {
			return deleteAllResources()
		}

		if len(args) < 2 {
			return fmt.Errorf("error: must specify resource name\nUsage: knoxctl delete %s <name>", resourceType)
		}

		resourceName := args[1]

		client, err := kube.GetClient(kubeconfig)
		if err != nil {
			return fmt.Errorf("error creating kubernetes client: %w", err)
		}

		ns := namespace

		switch resourceType {
		case "pod", "pods", "po":
			err = client.CoreV1().Pods(ns).Delete(context.TODO(), resourceName, metav1.DeleteOptions{})
			if err != nil {
				return fmt.Errorf("error deleting pod %s: %w", resourceName, err)
			}
			fmt.Printf("%s %s\n", colorGreen("pod"), colorBold("\"" + resourceName + "\" deleted"))

		case "deployment", "deployments", "deploy":
			err = client.AppsV1().Deployments(ns).Delete(context.TODO(), resourceName, metav1.DeleteOptions{})
			if err != nil {
				return fmt.Errorf("error deleting deployment %s: %w", resourceName, err)
			}
			fmt.Printf("%s %s\n", colorGreen("deployment.apps"), colorBold("\"" + resourceName + "\" deleted"))

		case "service", "services", "svc":
			err = client.CoreV1().Services(ns).Delete(context.TODO(), resourceName, metav1.DeleteOptions{})
			if err != nil {
				return fmt.Errorf("error deleting service %s: %w", resourceName, err)
			}
			fmt.Printf("%s %s\n", colorGreen("service"), colorBold("\"" + resourceName + "\" deleted"))

		case "namespace", "namespaces", "ns":
			err = client.CoreV1().Namespaces().Delete(context.TODO(), resourceName, metav1.DeleteOptions{})
			if err != nil {
				return fmt.Errorf("error deleting namespace %s: %w", resourceName, err)
			}
			fmt.Printf("%s %s\n", colorGreen("namespace"), colorBold("\"" + resourceName + "\" deleted"))

		case "node", "nodes", "no":
			err = client.CoreV1().Nodes().Delete(context.TODO(), resourceName, metav1.DeleteOptions{})
			if err != nil {
				return fmt.Errorf("error deleting node %s: %w", resourceName, err)
			}
			fmt.Printf("%s %s\n", colorGreen("node"), colorBold("\"" + resourceName + "\" deleted"))

		default:
			return fmt.Errorf("error: unknown resource type \"%s\"\nSupported: pod, deployment, service, namespace, node, all", resourceType)
		}

		return nil
	},
}

func deleteAllResources() error {
	client, err := kube.GetClient(kubeconfig)
	if err != nil {
		return fmt.Errorf("error creating kubernetes client: %w", err)
	}

	ns := namespace
	deleted := 0

	fmt.Printf("Deleting all resources in namespace \"%s\"...\n\n", ns)

	deployments, err := client.AppsV1().Deployments(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error listing deployments: %w", err)
	}
	for _, deploy := range deployments.Items {
		err = client.AppsV1().Deployments(ns).Delete(context.TODO(), deploy.Name, metav1.DeleteOptions{})
		if err != nil {
			fmt.Printf("  ✗ deployment.apps \"%s\": %v\n", deploy.Name, err)
		} else {
			fmt.Printf("  deployment.apps \"%s\" deleted\n", deploy.Name)
			deleted++
		}
	}

	services, err := client.CoreV1().Services(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error listing services: %w", err)
	}
	for _, svc := range services.Items {
		if svc.Name == "kubernetes" && ns == "default" {
			continue
		}
		err = client.CoreV1().Services(ns).Delete(context.TODO(), svc.Name, metav1.DeleteOptions{})
		if err != nil {
			fmt.Printf("  ✗ service \"%s\": %v\n", svc.Name, err)
		} else {
			fmt.Printf("  service \"%s\" deleted\n", svc.Name)
			deleted++
		}
	}

	pods, err := client.CoreV1().Pods(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error listing pods: %w", err)
	}
	for _, pod := range pods.Items {
		err = client.CoreV1().Pods(ns).Delete(context.TODO(), pod.Name, metav1.DeleteOptions{})
		if err != nil {
			fmt.Printf("  ✗ pod \"%s\": %v\n", pod.Name, err)
		} else {
			fmt.Printf("  pod \"%s\" deleted\n", pod.Name)
			deleted++
		}
	}

	if deleted == 0 {
		fmt.Println("No resources found to delete.")
	} else {
		fmt.Printf("\n%d resource(s) deleted.\n", deleted)
	}

	return nil
}

func deleteFromFile() error {
	data, err := os.ReadFile(deleteFile)
	if err != nil {
		return fmt.Errorf("error reading file %s: %w", deleteFile, err)
	}

	decUnstructured := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	obj := &unstructured.Unstructured{}
	_, gvk, err := decUnstructured.Decode(data, nil, obj)
	if err != nil {
		return fmt.Errorf("error decoding YAML: %w", err)
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

	err = resourceClient.Delete(context.TODO(), obj.GetName(), metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("error deleting resource: %w", err)
	}

	fmt.Printf("%s %s\n", colorGreen(strings.ToLower(gvk.Kind)), colorBold("\"" + obj.GetName() + "\" deleted"))
	return nil
}

func init() {
	deleteCmd.Flags().StringVarP(&deleteFile, "file", "f", "", "Path to the YAML file to delete")
	rootCmd.AddCommand(deleteCmd)
}

