package cmd

import (
	"fmt"

	"github.com/knoxctl/pkg/kube"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of knoxctl and the Kubernetes cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("%s %s\n", colorBold("knoxctl"), colorGreen("v1.1.0"))

		client, err := kube.GetClient(kubeconfig)
		if err != nil {
			fmt.Printf("%s %s\n", colorBold("Kubernetes Server:"), colorRed("unable to connect"))
			return nil
		}

		info, err := client.Discovery().ServerVersion()
		if err != nil {
			fmt.Printf("%s %s\n", colorBold("Kubernetes Server:"), colorRed("unable to retrieve version"))
			return nil
		}

		fmt.Printf("%s %s\n", colorBold("Kubernetes Server:"), colorCyan(info.GitVersion))
		fmt.Printf("%s %s\n", colorBold("Platform:"), colorCyan(info.Platform))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
