package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/knoxctl/pkg/kube"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var describeCmd = &cobra.Command{
	Use:   "describe <resource> <name>",
	Short: "Show detailed information about a specific resource",
	Long:  `Display detailed information about a Kubernetes resource including metadata, spec, and status.`,
	Example: `  knoxctl describe pod my-pod
  knoxctl describe deployment nginx
  knoxctl describe svc my-service
  knoxctl describe node minikube
  knoxctl describe ns kube-system`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		resourceType := strings.ToLower(args[0])
		resourceName := args[1]

		switch resourceType {
		case "pod", "pods", "po":
			return describePod(resourceName)
		case "deployment", "deployments", "deploy":
			return describeDeployment(resourceName)
		case "svc", "service", "services":
			return describeService(resourceName)
		case "node", "nodes", "no":
			return describeNode(resourceName)
		case "ns", "namespace", "namespaces":
			return describeNamespace(resourceName)
		default:
			return fmt.Errorf("unknown resource type %q\nSupported: pod, deployment, svc, node, ns", resourceType)
		}
	},
}

func printField(label, value string) {
	fmt.Printf("%-24s %s\n", colorCyan(label+":"), value)
}

func printSection(title string) {
	fmt.Printf("\n%s\n%s\n", colorHeader(title), colorHeader(strings.Repeat("─", len(title))))
}

func describePod(podName string) error {
	c, err := kube.GetClient(kubeconfig)
	if err != nil {
		return err
	}
	pod, err := c.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("pod %q not found in namespace %q: %w", podName, namespace, err)
	}

	fmt.Printf("\n%s %s\n", colorHeader("Pod:"), colorBold(pod.Name))
	printField("Namespace", pod.Namespace)
	printField("Node", pod.Spec.NodeName)
	printField("Status", podStatusColor(string(pod.Status.Phase)))
	printField("IP", pod.Status.PodIP)
	printField("Created", formatAge(pod.CreationTimestamp.Time)+" ago")

	printSection("Labels")
	if len(pod.Labels) == 0 {
		fmt.Println("  <none>")
	}
	for k, v := range pod.Labels {
		fmt.Printf("  %s=%s\n", colorCyan(k), v)
	}

	printSection("Containers")
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
	fmt.Fprintln(w, colorHeader("  NAME\tIMAGE\tREADY\tRESTARTS"))
	for _, cs := range pod.Status.ContainerStatuses {
		ready := colorRed("false")
		if cs.Ready {
			ready = colorGreen("true")
		}
		fmt.Fprintf(w, "  %s\t%s\t%s\t%d\n", colorBold(cs.Name), cs.Image, ready, cs.RestartCount)
	}
	w.Flush()

	printSection("Volumes")
	if len(pod.Spec.Volumes) == 0 {
		fmt.Println("  <none>")
	}
	for _, vol := range pod.Spec.Volumes {
		fmt.Printf("  %s\n", colorBold(vol.Name))
	}
	return nil
}

func describeDeployment(deployName string) error {
	c, err := kube.GetClient(kubeconfig)
	if err != nil {
		return err
	}
	deploy, err := c.AppsV1().Deployments(namespace).Get(context.TODO(), deployName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("deployment %q not found in namespace %q: %w", deployName, namespace, err)
	}

	var desired int32
	if deploy.Spec.Replicas != nil {
		desired = *deploy.Spec.Replicas
	}

	fmt.Printf("\n%s %s\n", colorHeader("Deployment:"), colorBold(deploy.Name))
	printField("Namespace", deploy.Namespace)
	printField("Created", formatAge(deploy.CreationTimestamp.Time)+" ago")
	printField("Replicas", replicaColor(deploy.Status.ReadyReplicas, desired))
	printField("Updated", fmt.Sprintf("%d", deploy.Status.UpdatedReplicas))
	printField("Available", fmt.Sprintf("%d", deploy.Status.AvailableReplicas))
	printField("Strategy", string(deploy.Spec.Strategy.Type))

	printSection("Labels")
	if len(deploy.Labels) == 0 {
		fmt.Println("  <none>")
	}
	for k, v := range deploy.Labels {
		fmt.Printf("  %s=%s\n", colorCyan(k), v)
	}

	printSection("Selector")
	for k, v := range deploy.Spec.Selector.MatchLabels {
		fmt.Printf("  %s=%s\n", colorCyan(k), v)
	}

	printSection("Containers")
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
	fmt.Fprintln(w, colorHeader("  NAME\tIMAGE"))
	for _, cont := range deploy.Spec.Template.Spec.Containers {
		fmt.Fprintf(w, "  %s\t%s\n", colorBold(cont.Name), cont.Image)
	}
	w.Flush()
	return nil
}

func describeService(svcName string) error {
	c, err := kube.GetClient(kubeconfig)
	if err != nil {
		return err
	}
	svc, err := c.CoreV1().Services(namespace).Get(context.TODO(), svcName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("service %q not found in namespace %q: %w", svcName, namespace, err)
	}

	fmt.Printf("\n%s %s\n", colorHeader("Service:"), colorBold(svc.Name))
	printField("Namespace", svc.Namespace)
	printField("Type", string(svc.Spec.Type))
	printField("Cluster IP", svc.Spec.ClusterIP)
	printField("Created", formatAge(svc.CreationTimestamp.Time)+" ago")

	printSection("Ports")
	for _, port := range svc.Spec.Ports {
		if port.NodePort != 0 {
			fmt.Printf("  %s %d → %d/%s\n", colorBold(port.Name), port.Port, port.NodePort, port.Protocol)
		} else {
			fmt.Printf("  %s %d/%s\n", colorBold(port.Name), port.Port, port.Protocol)
		}
	}

	printSection("Selector")
	if len(svc.Spec.Selector) == 0 {
		fmt.Println("  <none>")
	}
	for k, v := range svc.Spec.Selector {
		fmt.Printf("  %s=%s\n", colorCyan(k), v)
	}
	return nil
}

func describeNode(nodeName string) error {
	c, err := kube.GetClient(kubeconfig)
	if err != nil {
		return err
	}
	node, err := c.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("node %q not found: %w", nodeName, err)
	}

	status := "NotReady"
	for _, cond := range node.Status.Conditions {
		if cond.Type == "Ready" && cond.Status == "True" {
			status = "Ready"
		}
	}

	fmt.Printf("\n%s %s\n", colorHeader("Node:"), colorBold(node.Name))
	printField("Status", nodeStatusColor(status))
	printField("OS", node.Status.NodeInfo.OSImage)
	printField("Kernel", node.Status.NodeInfo.KernelVersion)
	printField("Container Runtime", node.Status.NodeInfo.ContainerRuntimeVersion)
	printField("Kubelet", node.Status.NodeInfo.KubeletVersion)
	printField("Architecture", node.Status.NodeInfo.Architecture)
	printField("CPU", node.Status.Capacity.Cpu().String())
	printField("Memory", node.Status.Capacity.Memory().String())
	printField("Created", formatAge(node.CreationTimestamp.Time)+" ago")
	return nil
}

func describeNamespace(nsName string) error {
	c, err := kube.GetClient(kubeconfig)
	if err != nil {
		return err
	}
	ns, err := c.CoreV1().Namespaces().Get(context.TODO(), nsName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("namespace %q not found: %w", nsName, err)
	}

	fmt.Printf("\n%s %s\n", colorHeader("Namespace:"), colorBold(ns.Name))
	printField("Status", colorGreen(string(ns.Status.Phase)))
	printField("Created", formatAge(ns.CreationTimestamp.Time)+" ago")

	printSection("Labels")
	if len(ns.Labels) == 0 {
		fmt.Println("  <none>")
	}
	for k, v := range ns.Labels {
		fmt.Printf("  %s=%s\n", colorCyan(k), v)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(describeCmd)
}
