package check

import (
	"fmt"
	"strings"
	"time"
)

// GradeColor returns the color (hex, without leading '#') associated with a
// grade. Colors mirror the palette used by the Go Report Card website.
func GradeColor(grade Grade) string {
	switch grade {
	case GradeAPlus:
		return "44cc11" // brightgreen
	case GradeA:
		return "97ca00" // green
	case GradeB:
		return "a4a61d" // yellowgreen
	case GradeC:
		return "dfb317" // yellow
	case GradeD:
		return "fe7d37" // orange
	case GradeE, GradeF:
		return "e05d44" // red
	default:
		return "9f9f9f" // lightgrey
	}
}

// approximate text width (in pixels) for the 11px Verdana/DejaVu font that the
// shields.io "flat" style uses. This keeps the badge nicely sized without
// requiring a font-metrics library.
func textWidth(s string) int {
	w := 0
	for _, r := range s {
		switch {
		case r >= 'A' && r <= 'Z':
			w += 8
		case r == 'i' || r == 'l' || r == 'j' || r == '.' || r == ' ':
			w += 3
		case r == '+' || r == '%':
			w += 8
		default:
			w += 6
		}
	}
	return w
}

// GenerateBadgeSVG renders a self-contained "flat" style badge SVG for the
// given grade. It does not depend on any external service, so it is suitable
// for committing directly into a repository.
func GenerateBadgeSVG(grade Grade) string {
	label := "go report"
	message := string(grade)
	color := GradeColor(grade)

	const pad = 10
	labelWidth := textWidth(label) + pad*2
	messageWidth := textWidth(message) + pad*2
	totalWidth := labelWidth + messageWidth

	labelX := labelWidth * 10 / 2
	messageX := (labelWidth + messageWidth/2) * 10
	labelTextLen := (textWidth(label)) * 10
	messageTextLen := (textWidth(message)) * 10

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="%d" height="20" role="img" aria-label="%s: %s">`, totalWidth, label, message)
	fmt.Fprintf(&b, `<title>%s: %s</title>`, label, message)
	b.WriteString(`<linearGradient id="s" x2="0" y2="100%">`)
	b.WriteString(`<stop offset="0" stop-color="#bbb" stop-opacity=".1"/>`)
	b.WriteString(`<stop offset="1" stop-opacity=".1"/>`)
	b.WriteString(`</linearGradient>`)
	fmt.Fprintf(&b, `<clipPath id="r"><rect width="%d" height="20" rx="3" fill="#fff"/></clipPath>`, totalWidth)
	b.WriteString(`<g clip-path="url(#r)">`)
	fmt.Fprintf(&b, `<rect width="%d" height="20" fill="#555"/>`, labelWidth)
	fmt.Fprintf(&b, `<rect x="%d" width="%d" height="20" fill="#%s"/>`, labelWidth, messageWidth, color)
	fmt.Fprintf(&b, `<rect width="%d" height="20" fill="url(#s)"/>`, totalWidth)
	b.WriteString(`</g>`)
	b.WriteString(`<g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" text-rendering="geometricPrecision" font-size="110">`)
	fmt.Fprintf(&b, `<text aria-hidden="true" x="%d" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="%d">%s</text>`, labelX, labelTextLen, label)
	fmt.Fprintf(&b, `<text x="%d" y="140" transform="scale(.1)" fill="#fff" textLength="%d">%s</text>`, labelX, labelTextLen, label)
	fmt.Fprintf(&b, `<text aria-hidden="true" x="%d" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="%d">%s</text>`, messageX, messageTextLen, message)
	fmt.Fprintf(&b, `<text x="%d" y="140" transform="scale(.1)" fill="#fff" textLength="%d">%s</text>`, messageX, messageTextLen, message)
	b.WriteString(`</g>`)
	b.WriteString(`</svg>`)

	return b.String()
}

// GenerateReportMarkdown renders a human-readable Markdown quality report for a
// completed ChecksResult. It embeds the overall grade, a per-check score table,
// and a breakdown of the files with issues. It is suitable for committing into
// a repository (e.g. under .github/).
func GenerateReportMarkdown(result ChecksResult) string {
	var b strings.Builder

	b.WriteString("# Go Report Card\n\n")
	fmt.Fprintf(&b, "**Grade: %s** (%.1f%%)\n\n", result.Grade, result.Average*100)

	b.WriteString("| Metric | Value |\n")
	b.WriteString("| ------ | ----- |\n")
	fmt.Fprintf(&b, "| Files | %d |\n", result.Files)
	fmt.Fprintf(&b, "| Issues | %d |\n\n", result.Issues)

	b.WriteString("## Checks\n\n")
	b.WriteString("| Check | Score |\n")
	b.WriteString("| ----- | ----- |\n")
	for _, c := range result.Checks {
		fmt.Fprintf(&b, "| %s | %d%% |\n", c.Name, int64(c.Percentage*100))
	}
	b.WriteString("\n")

	hasIssues := false
	for _, c := range result.Checks {
		if len(c.FileSummaries) == 0 {
			continue
		}
		if !hasIssues {
			b.WriteString("## Issues\n\n")
			hasIssues = true
		}
		fmt.Fprintf(&b, "### %s\n\n", c.Name)
		for _, f := range c.FileSummaries {
			switch {
			case f.Filename == "" && f.FileURL != "":
				fmt.Fprintf(&b, "- See [%s](%s)\n", f.FileURL, f.FileURL)
			case f.FileURL != "":
				fmt.Fprintf(&b, "- [`%s`](%s)\n", f.Filename, f.FileURL)
			default:
				fmt.Fprintf(&b, "- `%s`\n", f.Filename)
			}
			for _, e := range f.Errors {
				fmt.Fprintf(&b, "  - Line %d: %s\n", e.LineNumber, strings.TrimSpace(e.ErrorString))
			}
		}
		b.WriteString("\n")
	}

	if !hasIssues {
		b.WriteString("No issues found. Nice work!\n\n")
	}

	fmt.Fprintf(&b, "---\n\n_Generated by [Go Report Card](https://github.com/gojp/goreportcard) on %s._\n",
		time.Now().UTC().Format("2006-01-02 15:04:05 UTC"))

	return b.String()
}
