package cmd

import (
	"fmt"

	"github.com/fatih/color"
)

var (
	colorHeader = color.New(color.FgCyan, color.Bold).SprintFunc()
	colorBold   = color.New(color.Bold).SprintFunc()
	colorGreen  = color.New(color.FgGreen).SprintFunc()
	colorRed    = color.New(color.FgRed).SprintFunc()
	colorYellow = color.New(color.FgYellow).SprintFunc()
	colorCyan   = color.New(color.FgCyan).SprintFunc()
)

func podStatusColor(status string) string {
	switch status {
	case "Running":
		return colorGreen(status)
	case "Pending", "ContainerCreating", "Init:0/1":
		return colorYellow(status)
	case "Failed", "CrashLoopBackOff", "Error", "OOMKilled", "ImagePullBackOff", "ErrImagePull", "Terminating":
		return colorRed(status)
	default:
		return status
	}
}

func nodeStatusColor(status string) string {
	if status == "Ready" {
		return colorGreen(status)
	}
	return colorRed(status)
}

func replicaColor(ready, desired int32) string {
	if ready == desired && desired > 0 {
		return colorGreen(fmt.Sprintf("%d/%d", ready, desired))
	} else if ready == 0 {
		return colorRed(fmt.Sprintf("%d/%d", ready, desired))
	}
	return colorYellow(fmt.Sprintf("%d/%d", ready, desired))
}
