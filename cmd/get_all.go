package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/knoxctl/pkg/kube"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var getAllCmd = &cobra.Command{
	Use:   "all",
	Short: "List all resources (pods, deployments, services) in a namespace",
	Long:  `Display all pods, deployments, and services in the specified namespace. Use -A to list across all namespaces.`,
	Example: `  # Get all resources in default namespace
  knoxctl get all

  # Get all resources in a specific namespace
  knoxctl get all -n kube-system

  # Get all resources across all namespaces
  knoxctl get all -A`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := kube.GetClient(kubeconfig)
		if err != nil {
			return fmt.Errorf("error creating kubernetes client: %w", err)
		}

		ns := namespace
		if allNamespaces {
			ns = ""
		}

		pods, err := client.CoreV1().Pods(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("error listing pods: %w", err)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)

		if len(pods.Items) > 0 {
			fmt.Println(colorHeader("=== Pods ==="))
			if allNamespaces {
				fmt.Fprintln(w, colorHeader("NAMESPACE\tNAME\tREADY\tSTATUS\tRESTARTS\tAGE"))
			} else {
				fmt.Fprintln(w, colorHeader("NAME\tREADY\tSTATUS\tRESTARTS\tAGE"))
			}

			for _, pod := range pods.Items {
				readyCount := 0
				totalContainers := len(pod.Spec.Containers)
				var restartCount int32

				for _, cs := range pod.Status.ContainerStatuses {
					if cs.Ready {
						readyCount++
					}
					restartCount += cs.RestartCount
				}

				ready := fmt.Sprintf("%d/%d", readyCount, totalContainers)
				age := formatAge(pod.CreationTimestamp.Time)
				status := string(pod.Status.Phase)

				for _, cs := range pod.Status.ContainerStatuses {
					if cs.State.Waiting != nil && cs.State.Waiting.Reason != "" {
						status = cs.State.Waiting.Reason
						break
					}
					if cs.State.Terminated != nil && cs.State.Terminated.Reason != "" {
						status = cs.State.Terminated.Reason
						break
					}
				}

				if allNamespaces {
					fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%s\n",
						pod.Namespace, colorBold(pod.Name), ready, podStatusColor(status), restartCount, age)
				} else {
					fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\n",
						colorBold(pod.Name), ready, podStatusColor(status), restartCount, age)
				}
			}
			w.Flush()
			fmt.Println()
		}

		deployments, err := client.AppsV1().Deployments(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("error listing deployments: %w", err)
		}

		if len(deployments.Items) > 0 {
			fmt.Println(colorHeader("=== Deployments ==="))
			w = tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)

			if allNamespaces {
				fmt.Fprintln(w, colorHeader("NAMESPACE\tNAME\tREADY\tUP-TO-DATE\tAVAILABLE\tAGE"))
			} else {
				fmt.Fprintln(w, colorHeader("NAME\tREADY\tUP-TO-DATE\tAVAILABLE\tAGE"))
			}

			for _, deploy := range deployments.Items {
				var desired int32
				if deploy.Spec.Replicas != nil {
					desired = *deploy.Spec.Replicas
				}
				age := formatAge(deploy.CreationTimestamp.Time)

				if allNamespaces {
					fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\t%s\n",
						deploy.Namespace, colorBold(deploy.Name), replicaColor(deploy.Status.ReadyReplicas, desired),
						deploy.Status.UpdatedReplicas, deploy.Status.AvailableReplicas, age)
				} else {
					fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%s\n",
						colorBold(deploy.Name), replicaColor(deploy.Status.ReadyReplicas, desired),
						deploy.Status.UpdatedReplicas, deploy.Status.AvailableReplicas, age)
				}
			}
			w.Flush()
			fmt.Println()
		}

		services, err := client.CoreV1().Services(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("error listing services: %w", err)
		}

		if len(services.Items) > 0 {
			fmt.Println(colorHeader("=== Services ==="))
			w = tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)

			if allNamespaces {
				fmt.Fprintln(w, colorHeader("NAMESPACE\tNAME\tTYPE\tCLUSTER-IP\tEXTERNAL-IP\tPORT(S)\tAGE"))
			} else {
				fmt.Fprintln(w, colorHeader("NAME\tTYPE\tCLUSTER-IP\tEXTERNAL-IP\tPORT(S)\tAGE"))
			}

			for _, svc := range services.Items {
				var ports []string
				for _, port := range svc.Spec.Ports {
					if port.NodePort != 0 {
						ports = append(ports, fmt.Sprintf("%d:%d/%s", port.Port, port.NodePort, port.Protocol))
					} else {
						ports = append(ports, fmt.Sprintf("%d/%s", port.Port, port.Protocol))
					}
				}
				portStr := joinOrNone(ports)

				externalIP := "<none>"
				if len(svc.Spec.ExternalIPs) > 0 {
					externalIP = joinStrings(svc.Spec.ExternalIPs)
				} else if len(svc.Status.LoadBalancer.Ingress) > 0 {
					var ips []string
					for _, ingress := range svc.Status.LoadBalancer.Ingress {
						if ingress.IP != "" {
							ips = append(ips, ingress.IP)
						} else if ingress.Hostname != "" {
							ips = append(ips, ingress.Hostname)
						}
					}
					if len(ips) > 0 {
						externalIP = joinStrings(ips)
					}
				}

				clusterIP := svc.Spec.ClusterIP
				if clusterIP == "" {
					clusterIP = "<none>"
				}

				age := formatAge(svc.CreationTimestamp.Time)

				if allNamespaces {
					fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
						svc.Namespace, svc.Name, string(svc.Spec.Type), clusterIP, externalIP, portStr, age)
				} else {
					fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
						svc.Name, string(svc.Spec.Type), clusterIP, externalIP, portStr, age)
				}
			}
			w.Flush()
			fmt.Println()
		}

		if len(pods.Items) == 0 && len(deployments.Items) == 0 && len(services.Items) == 0 {
			fmt.Println("No resources found.")
		}

		return nil
	},
}

func joinOrNone(s []string) string {
	if len(s) == 0 {
		return "<none>"
	}
	return joinStrings(s)
}

func joinStrings(s []string) string {
	result := ""
	for i, v := range s {
		if i > 0 {
			result += ","
		}
		result += v
	}
	return result
}

func init() {
	getCmd.AddCommand(getAllCmd)
}
