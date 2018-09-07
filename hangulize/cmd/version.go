package cmd

import (
	"github.com/spf13/cobra"

	"github.com/hangulize/hangulize"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Hangulize",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Printf("hangulize-%s\n", hangulize.Version)
	},
}
