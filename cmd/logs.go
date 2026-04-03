package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/knoxctl/pkg/kube"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
)

var (
	logsContainer string
	logsTail      int64
	logsFollow    bool
)

var logsCmd = &cobra.Command{
	Use:   "logs <pod-name>",
	Short: "Print the logs of a pod",
	Long:  `Fetch and stream logs from a pod running in the Kubernetes cluster.`,
	Example: `  # Print last 50 lines of logs from a pod
  knoxctl logs my-pod

  # Print logs from a specific container in a pod
  knoxctl logs my-pod -c nginx

  # Stream live logs from a pod
  knoxctl logs my-pod -f

  # Show last 100 lines of logs
  knoxctl logs my-pod --tail 100`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		podName := args[0]

		client, err := kube.GetClient(kubeconfig)
		if err != nil {
			return fmt.Errorf("error creating kubernetes client: %w", err)
		}

		opts := &corev1.PodLogOptions{
			Follow: logsFollow,
		}

		if logsTail > 0 {
			opts.TailLines = &logsTail
		}

		if logsContainer != "" {
			opts.Container = logsContainer
		}

		fmt.Printf("%s %s/%s\n\n", colorHeader("Logs for pod:"), colorCyan(namespace), colorBold(podName))

		req := client.CoreV1().Pods(namespace).GetLogs(podName, opts)
		stream, err := req.Stream(context.TODO())
		if err != nil {
			return fmt.Errorf("error opening log stream for pod %s: %w\nTip: use -c to specify a container if this pod has multiple", podName, err)
		}
		defer stream.Close()

		_, err = io.Copy(os.Stdout, stream)
		if err != nil && err != io.EOF {
			return fmt.Errorf("error reading logs: %w", err)
		}

		return nil
	},
}

func init() {
	logsCmd.Flags().StringVarP(&logsContainer, "container", "c", "", "Container name (required for pods with multiple containers)")
	logsCmd.Flags().Int64Var(&logsTail, "tail", 50, "Number of recent log lines to show (default: 50)")
	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "Stream logs live")
	rootCmd.AddCommand(logsCmd)
}
