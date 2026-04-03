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

var getNodesCmd = &cobra.Command{
	Use:     "nodes",
	Aliases: []string{"node", "no"},
	Short:   "List all nodes in the cluster",
	Long:    `Display all nodes in the Kubernetes cluster with their status, roles, age, and version.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := kube.GetClient(kubeconfig)
		if err != nil {
			return fmt.Errorf("error creating kubernetes client: %w", err)
		}

		nodes, err := client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("error listing nodes: %w", err)
		}

		if len(nodes.Items) == 0 {
			fmt.Println("No nodes found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
		fmt.Fprintln(w, colorHeader("NAME\tSTATUS\tROLES\tAGE\tVERSION"))

		for _, node := range nodes.Items {
			status := "NotReady"
			for _, condition := range node.Status.Conditions {
				if condition.Type == "Ready" {
					if condition.Status == "True" {
						status = "Ready"
					}
					break
				}
			}

			var roles []string
			for label := range node.Labels {
				if strings.HasPrefix(label, "node-role.kubernetes.io/") {
					role := strings.TrimPrefix(label, "node-role.kubernetes.io/")
					if role != "" {
						roles = append(roles, role)
					}
				}
			}
			roleStr := strings.Join(roles, ",")
			if roleStr == "" {
				roleStr = "<none>"
			}

			age := formatAge(node.CreationTimestamp.Time)
			version := node.Status.NodeInfo.KubeletVersion

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				colorBold(node.Name), nodeStatusColor(status), roleStr, age, version)
		}

		w.Flush()
		return nil
	},
}

func init() {
	getCmd.AddCommand(getNodesCmd)
}
