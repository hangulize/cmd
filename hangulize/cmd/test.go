package cmd

import (
	"os"

	"github.com/hangulize/hangulize"
	"github.com/spf13/cobra"
)

var testCover bool
var testCoverProfile string

func init() {
	testCmd.Flags().BoolVarP(
		&testCover, "cover", "", false,
		"Enable coverage analysis.",
	)
	testCmd.Flags().StringVarP(
		&testCoverProfile, "coverprofile", "", "",
		"Write a coverage profile to the file after all tests have passed.",
	)

	rootCmd.AddCommand(testCmd)
}

var testCmd = &cobra.Command{
	Use:   "test HGL [HGL...]",
	Short: "Run test of HGL specs",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var cover *cover

		if testCoverProfile != "" {
			testCover = true
		}
		if testCover {
			cover = newCover()
		}

		failedAtLeastOnce := false

		for _, name := range args {
			// Open the HGL spec file.
			file, err := os.Open(name)
			if err != nil {
				return err
			}

			// Parse the spec.
			spec, err := hangulize.ParseSpec(file)
			if err != nil {
				return err
			}

			// Remember the name.
			cover.Visit(name)

			var (
				word        string
				expected    string
				transcribed string
				traces      []hangulize.Trace
			)
			h := hangulize.NewHangulizer(spec)

			// Run test.
			for _, exm := range spec.Test {
				word, expected = exm[0], exm[1]

				if testCover {
					transcribed, traces = h.HangulizeTrace(word)
					for _, tr := range traces {
						if tr.HasRule {
							cover.Cover(name, tr.Step, tr.Rule.ID)
						}
					}
				} else {
					transcribed = h.Hangulize(word)
				}

				if transcribed == expected {
					continue
				}

				// Test failed.
				cmd.Printf("%s: ", name)
				cmd.Printf(`"%s" -> "%s"`, word, transcribed)
				cmd.Printf(`, expected: "%s"`, expected)
				cmd.Println()
				failedAtLeastOnce = true
			}
		}

		// Exit with 1 if failed at least once.
		if failedAtLeastOnce {
			os.Exit(1)
		}

		// Save the coverage profile.
		if testCoverProfile != "" {
			file, _ := os.OpenFile(
				testCoverProfile,
				os.O_WRONLY|os.O_CREATE,
				0644,
			)
			defer file.Close()

			cover.WriteProfile(file)
		}

		// Print the coverage.
		if testCover {
			coverage := cover.Coverage()
			cmd.Printf("coverage: %.1f%% of rules\n", coverage*100)
		}

		return nil
	},
}
