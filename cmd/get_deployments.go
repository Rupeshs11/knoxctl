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

var getDeploymentsCmd = &cobra.Command{
	Use:     "deployments",
	Aliases: []string{"deployment", "deploy"},
	Short:   "List all deployments in a namespace",
	Long:    `Display all deployments in the specified namespace with their readiness, up-to-date, available replicas, and age.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := kube.GetClient(kubeconfig)
		if err != nil {
			return fmt.Errorf("error creating kubernetes client: %w", err)
		}

		ns := namespace
		if allNamespaces {
			ns = ""
		}

		deployments, err := client.AppsV1().Deployments(ns).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("error listing deployments: %w", err)
		}

		if len(deployments.Items) == 0 {
			fmt.Println("No deployments found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)

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
			ready := replicaColor(deploy.Status.ReadyReplicas, desired)
			age := formatAge(deploy.CreationTimestamp.Time)

			if allNamespaces {
				fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\t%s\n",
					deploy.Namespace, colorBold(deploy.Name), ready, deploy.Status.UpdatedReplicas,
					deploy.Status.AvailableReplicas, age)
			} else {
				fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%s\n",
					colorBold(deploy.Name), ready, deploy.Status.UpdatedReplicas,
					deploy.Status.AvailableReplicas, age)
			}
		}

		w.Flush()
		return nil
	},
}

func init() {
	getCmd.AddCommand(getDeploymentsCmd)
}
