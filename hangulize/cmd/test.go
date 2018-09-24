package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/hangulize/hangulize"
	"github.com/hangulize/hgl"
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

		if testCover {
			coverage := cover.Coverage()
			cmd.Printf("coverage: %.1f%% of rules\n", coverage*100)
		}

		return nil
	},
}

type cov struct {
	name   string
	step   hangulize.Step
	ruleID int
}

// cover stores covered rule IDs.
type cover struct {
	covered map[cov]bool
	names   map[string]bool
}

func newCover() *cover {
	return &cover{make(map[cov]bool), make(map[string]bool)}
}

func (c *cover) Visit(name string) {
	if c == nil {
		return
	}

	c.names[name] = true
}

func (c *cover) Cover(name string, step hangulize.Step, ruleID int) {
	if c == nil {
		return
	}

	c.covered[cov{name, step, ruleID}] = true
	c.Visit(name)
}

func (c *cover) Covered(name string, step hangulize.Step, ruleID int) bool {
	if c == nil {
		return false
	}

	return c.covered[cov{name, step, ruleID}]
}

func (c *cover) Coverage() float64 {
	if c == nil {
		return 0.0
	}

	var total int

	for name := range c.names {
		file, err := os.Open(name)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		spec, err := hangulize.ParseSpec(file)
		if err != nil {
			panic(err)
		}

		total += len(spec.Rewrite)
		total += len(spec.Transcribe)
	}

	return float64(len(c.covered)) / float64(total)
}

func (c *cover) WriteProfile(w io.Writer) {
	if c == nil {
		panic("nil")
	}

	io.WriteString(w, "mode: count\n")
	template := "%s:%d.1,%d.%d 1 %d\n"

	for name := range c.names {
		file, err := os.Open(name)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		// Parse as an HGL to track line numbers.
		hgl, err := hgl.Parse(file)
		if err != nil {
			panic(err)
		}

		// Parse as a spec to collect all
		file.Seek(0, 0)
		spec, _ := hangulize.ParseSpec(file)
		file.Seek(0, 0)

		cols := make([]int, 0)
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			col := len(scanner.Text()) + 1
			cols = append(cols, col)
		}

		rewritePairs := hgl["rewrite"].Pairs()
		for _, rule := range spec.Rewrite {
			line := rewritePairs[rule.ID].Line()
			col := cols[line-1]
			covered := btoi(c.Covered(name, hangulize.Rewrite, rule.ID))
			fmt.Fprintf(w, template, name, line, line, col, covered)
		}

		transcribePairs := hgl["transcribe"].Pairs()
		for _, rule := range spec.Transcribe {
			line := transcribePairs[rule.ID].Line()
			col := cols[line-1]
			covered := btoi(c.Covered(name, hangulize.Transcribe, rule.ID))
			fmt.Fprintf(w, template, name, line, line, col, covered)
		}
	}
}

func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}
