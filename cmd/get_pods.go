package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/knoxctl/pkg/kube"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var getPodsCmd = &cobra.Command{
	Use:     "pods",
	Aliases: []string{"pod", "po"},
	Short:   "List all pods in a namespace",
	Long:    `Display all pods in the specified namespace with their status, readiness, restarts, and age.`,
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

		if len(pods.Items) == 0 {
			fmt.Println("No pods found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)

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
		return nil
	},
}

func formatAge(t time.Time) string {
	duration := time.Since(t)

	if duration.Hours() >= 24*365 {
		years := int(duration.Hours() / (24 * 365))
		return fmt.Sprintf("%dy", years)
	}
	if duration.Hours() >= 24 {
		days := int(duration.Hours() / 24)
		return fmt.Sprintf("%dd", days)
	}
	if duration.Hours() >= 1 {
		hours := int(duration.Hours())
		return fmt.Sprintf("%dh", hours)
	}
	if duration.Minutes() >= 1 {
		minutes := int(duration.Minutes())
		return fmt.Sprintf("%dm", minutes)
	}
	seconds := int(duration.Seconds())
	return fmt.Sprintf("%ds", seconds)
}

func init() {
	getCmd.AddCommand(getPodsCmd)
}
