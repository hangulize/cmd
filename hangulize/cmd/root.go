package cmd

import (
	"fmt"
	"os"

	"github.com/hangulize/hangulize"
	"github.com/hangulize/hangulize/phonemize/furigana"
	"github.com/hangulize/hangulize/phonemize/pinyin"
	"github.com/spf13/cobra"
)

func init() {
	hangulize.UsePhonemizer(&furigana.P)
	hangulize.UsePhonemizer(&pinyin.P)
}

var rootCmd = &cobra.Command{
	Use:   "hangulize LANG WORD",
	Short: "Hangulize tools",

	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		lang := args[0]
		word := args[1]

		spec, ok := hangulize.LoadSpec(lang)
		if !ok {
			fmt.Println("Lang not supported:", lang)
			os.Exit(1)
		}

		h := hangulize.NewHangulizer(spec)
		fmt.Println(h.Hangulize(word))
	},
}

// Execute runs the root command. It's the entry point for every sub commands.
// When the running command returns an error, itt will report that and exit the
// process with exit code 1. So just call it in your main function.
func Execute() {
	err := rootCmd.Execute()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
