package goapi

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newVersionCommand() *cobra.Command {
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(*cobra.Command, []string) {
			fmt.Println("0.0.1")
		},
	}
	return versionCmd
}
