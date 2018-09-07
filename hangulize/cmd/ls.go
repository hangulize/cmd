package cmd

import (
	"github.com/spf13/cobra"

	"github.com/hangulize/hangulize"
)

func init() {
	rootCmd.AddCommand(lsCmd)
}

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List of bundled lang specs",
	Run: func(cmd *cobra.Command, args []string) {
		template := "%-8s %-8s %-24s %-24s\n"

		cmd.Printf(template, "LANG", "STAGE", "ENG", "KOR")

		for _, lang := range hangulize.ListLangs() {
			spec, _ := hangulize.LoadSpec(lang)
			cmd.Printf(template,
				lang,
				spec.Config.Stage,
				spec.Lang.English,
				spec.Lang.Korean,
			)
		}
	},
}
