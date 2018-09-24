package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/hangulize/hangulize"
	"github.com/hangulize/hgl"
)

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

// newCover creates a cover.
func newCover() *cover {
	return &cover{make(map[cov]bool), make(map[string]bool)}
}

// Visit marks a visited spec file name.
func (c *cover) Visit(name string) {
	if c == nil {
		return
	}

	c.names[name] = true
}

// Cover marks a covered rule.
func (c *cover) Cover(name string, step hangulize.Step, ruleID int) {
	if c == nil {
		return
	}

	c.covered[cov{name, step, ruleID}] = true
	c.Visit(name)
}

// Covered returns true if the rule has been covered within a test.
func (c *cover) Covered(name string, step hangulize.Step, ruleID int) bool {
	if c == nil {
		return false
	}

	return c.covered[cov{name, step, ruleID}]
}

// Coverage returns the test coverage.
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

// WriteProfile writes the coverage profile.
//
// The coverage profile format is similar to the Go test's one. So we can
// generate an HTML representation of coverage profile by "go tool cover":
//
//   $ hangulize test --coverprofile=cover.txt *.hgl
//   coverage: 100.0% of rules
//   $ go tool cover -html cover.txt -o cover.html
//
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
