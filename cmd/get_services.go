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

var getServicesCmd = &cobra.Command{
	Use:     "svc",
	Aliases: []string{"services", "service"},
	Short:   "List all services in a namespace",
	Long:    `Display all services in the specified namespace with their type, cluster IP, external IP, ports, and age.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := kube.GetClient(kubeconfig)
		if err != nil {
			return fmt.Errorf("error creating kubernetes client: %w", err)
		}

		ns := namespace
		if allNamespaces {
			ns = ""
		}

		services, err := client.CoreV1().Services(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("error listing services: %w", err)
		}

		if len(services.Items) == 0 {
			fmt.Println("No services found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)

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
			portStr := strings.Join(ports, ",")
			if portStr == "" {
				portStr = "<none>"
			}

			externalIP := "<none>"
			if len(svc.Spec.ExternalIPs) > 0 {
				externalIP = strings.Join(svc.Spec.ExternalIPs, ",")
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
					externalIP = strings.Join(ips, ",")
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
		return nil
	},
}

func init() {
	getCmd.AddCommand(getServicesCmd)
}
