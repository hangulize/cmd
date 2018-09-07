package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/hangulize/hangulize"
	"github.com/hangulize/hangulize/phonemize/furigana"
	"github.com/hangulize/hangulize/phonemize/pinyin"
	"github.com/spf13/cobra"
)

var verbose bool

func init() {
	hangulize.UsePhonemizer(&furigana.P)
	hangulize.UsePhonemizer(&pinyin.P)

	rootCmd.PersistentFlags().BoolVarP(
		&verbose, "verbose", "v", false, "verbose output")
}

var rootCmd = &cobra.Command{
	Use:   "hangulize LANG WORD",
	Short: "Hangulize tools",

	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		lang := args[0]

		spec, ok := hangulize.LoadSpec(lang)
		if !ok {
			cmd.Println("Lang not supported:", lang)
			os.Exit(1)
		}

		h := hangulize.NewHangulizer(spec)
		hangulizeStream(cmd, args, h)
	},
}

func hangulizeStream(
	cmd *cobra.Command,
	args []string,
	h *hangulize.Hangulizer,
) {
	ch := make(chan string)
	go readWords(ch, args)

	for {
		word := <-ch
		if word == "" {
			break
		}

		if verbose {
			transcribed, tr := h.HangulizeTrace(word)
			for _, t := range tr {
				cmd.Println(t.String())
			}
			cmd.Println(transcribed)
		} else {
			cmd.Println(h.Hangulize(word))
		}
	}
}

func readWords(ch chan<- string, args []string) {
	if len(args) == 1 {
		reader := bufio.NewReader(os.Stdin)
		for {
			word, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			word = strings.TrimSpace(word)
			ch <- word
		}
	} else {
		for _, word := range args[1:] {
			if word != "" {
				ch <- word
			}
		}
	}
	ch <- ""
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
