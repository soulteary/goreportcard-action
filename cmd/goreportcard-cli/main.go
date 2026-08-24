package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/soulteary/goreportcard-action/check"
)

// dotPrintf fills in the blank space between two strings with dots. The total
// length displayed is indicated by fullLen. The left string is specified by
// lfStr. The right string is formatted. At least two dots are shown, even if
// this means the total length exceeds fullLen.
func dotPrintf(w io.Writer, fullLen int, lfStr, rtFmtStr string, args ...interface{}) {
	rtStr := fmt.Sprintf(rtFmtStr, args...)
	dotLen := fullLen - len(lfStr) - len(rtStr)
	if dotLen < 2 {
		dotLen = 2
	}
	fmt.Fprintf(w, "%s %s %s\n", lfStr, strings.Repeat(".", dotLen), rtStr)
}

// options collects the CLI flags so run can be exercised from tests without
// mutating global state.
type options struct {
	dir     string
	verbose bool
	th      float64
	jsn     bool
	svg     string
	report  string
}

// run executes the grading pipeline and writes human-readable output to w.
// It returns the process exit code so main stays a thin wrapper.
func run(opts options, w io.Writer) int {
	result, err := check.Run(opts.dir, true)
	if err != nil {
		log.Printf("Fatal error checking %s: %s", opts.dir, err.Error())
		return 1
	}

	if opts.svg != "" {
		if err := os.WriteFile(opts.svg, []byte(check.GenerateBadgeSVG(result.Grade)), 0o644); err != nil {
			log.Printf("Fatal error writing SVG badge to %s: %s", opts.svg, err.Error())
			return 1
		}
		fmt.Fprintf(w, "Wrote badge (grade %s) to %s\n", result.Grade, opts.svg)
	}

	if opts.report != "" {
		if err := os.WriteFile(opts.report, []byte(check.GenerateReportMarkdown(result)), 0o644); err != nil {
			log.Printf("Fatal error writing report to %s: %s", opts.report, err.Error())
			return 1
		}
		fmt.Fprintf(w, "Wrote report (grade %s) to %s\n", result.Grade, opts.report)
	}

	if opts.jsn {
		marshalledResults, _ := json.Marshal(result)
		fmt.Fprintln(w, string(marshalledResults))
		return 0
	}

	dotPrintf(w, 24, "Grade", "%s %.1f%%", result.Grade, result.Average*100)
	dotPrintf(w, 24, "Files", "%d", result.Files)
	dotPrintf(w, 24, "Issues", "%d", result.Issues)

	for _, c := range result.Checks {
		dotPrintf(w, 24, c.Name, "%d%%", int64(c.Percentage*100))
		if opts.verbose && len(c.FileSummaries) > 0 {
			for _, f := range c.FileSummaries {
				fmt.Fprintf(w, "\t%s\n", f.Filename)
				for _, e := range f.Errors {
					fmt.Fprintf(w, "\t\tLine %d: %s\n", e.LineNumber, e.ErrorString)
				}
			}
		}
	}

	if result.Average*100 < opts.th {
		return 1
	}
	return 0
}

func main() {
	var opts options
	flag.StringVar(&opts.dir, "d", ".", "Root directory of your Go application")
	flag.BoolVar(&opts.verbose, "v", false, "Verbose output")
	flag.Float64Var(&opts.th, "t", 0, "Threshold of failure command")
	flag.BoolVar(&opts.jsn, "j", false, "JSON output. The binary will always exit with code 0")
	flag.StringVar(&opts.svg, "svg", "", "Write a self-contained report card badge SVG to the given file path")
	flag.StringVar(&opts.report, "report", "", "Write a Markdown quality report to the given file path")
	flag.Parse()

	os.Exit(run(opts, os.Stdout))
}
