package check

import (
	"strings"
	"testing"
)

func TestGradeFromPercentage(t *testing.T) {
	cases := []struct {
		pct  float64
		want Grade
	}{
		{100, GradeAPlus},
		{90.1, GradeAPlus},
		{90, GradeA},
		{80.5, GradeA},
		{80, GradeB},
		{70.5, GradeB},
		{70, GradeC},
		{60.5, GradeC},
		{60, GradeD},
		{50.5, GradeD},
		{50, GradeE},
		{40.5, GradeE},
		{40, GradeF},
		{0, GradeF},
	}
	for _, tt := range cases {
		if got := GradeFromPercentage(tt.pct); got != tt.want {
			t.Errorf("GradeFromPercentage(%v) = %q, want %q", tt.pct, got, tt.want)
		}
	}
}

func TestGradeColor(t *testing.T) {
	cases := []struct {
		grade Grade
		want  string
	}{
		{GradeAPlus, "44cc11"},
		{GradeA, "97ca00"},
		{GradeB, "a4a61d"},
		{GradeC, "dfb317"},
		{GradeD, "fe7d37"},
		{GradeE, "e05d44"},
		{GradeF, "e05d44"},
		{Grade("unknown"), "9f9f9f"},
	}
	for _, tt := range cases {
		if got := GradeColor(tt.grade); got != tt.want {
			t.Errorf("GradeColor(%q) = %q, want %q", tt.grade, got, tt.want)
		}
	}
}

func TestTextWidth(t *testing.T) {
	if got := textWidth(""); got != 0 {
		t.Errorf("textWidth(empty) = %d, want 0", got)
	}
	// Uppercase letters are the widest (8px each).
	if got := textWidth("AB"); got != 16 {
		t.Errorf("textWidth(%q) = %d, want 16", "AB", got)
	}
	// Narrow glyphs (i, l, j, ., space) are 3px each.
	if got := textWidth("il.j "); got != 15 {
		t.Errorf("textWidth(%q) = %d, want 15", "il.j ", got)
	}
	// '+' and '%' are 8px; default glyphs are 6px.
	if got := textWidth("+%a"); got != 22 {
		t.Errorf("textWidth(%q) = %d, want 22", "+%a", got)
	}
	// Longer strings must produce strictly larger widths.
	if textWidth("A+") >= textWidth("A+ 96.5%") {
		t.Errorf("expected longer message to be wider")
	}
}

func TestGenerateBadgeSVG(t *testing.T) {
	svg := GenerateBadgeSVG(GradeAPlus)

	for _, want := range []string{
		"<svg xmlns=\"http://www.w3.org/2000/svg\"",
		"go report",
		"A+",
		"#44cc11", // A+ color
		"</svg>",
	} {
		if !strings.Contains(svg, want) {
			t.Errorf("GenerateBadgeSVG(A+) missing %q\n%s", want, svg)
		}
	}

	// The message color must track the grade.
	if !strings.Contains(GenerateBadgeSVG(GradeF), "#e05d44") {
		t.Error("GenerateBadgeSVG(F) should use the red color")
	}

	// Output must be well-formed enough to parse round-trip length markers.
	if strings.Count(svg, "<text") != 4 {
		t.Errorf("expected 4 <text> elements, got %d", strings.Count(svg, "<text"))
	}
}

func TestGenerateReportMarkdown(t *testing.T) {
	result := ChecksResult{
		Average: 0.965,
		Grade:   GradeAPlus,
		Files:   42,
		Issues:  1,
		Checks: []Score{
			{Name: "gofmt", Percentage: 1.0},
			{
				Name:       "go_vet",
				Percentage: 0.9,
				FileSummaries: []FileSummary{
					{Filename: "a.go", FileURL: "https://example.com/a.go", Errors: []Error{{LineNumber: 12, ErrorString: "bad thing"}}},
					{FileURL: "https://example.com/pkg"},
					{Filename: "b.go"},
				},
			},
		},
	}

	md := GenerateReportMarkdown(result)
	for _, want := range []string{
		"# Go Report Card",
		"**Grade: A+** (96.5%)",
		"| Files | 42 |",
		"| Issues | 1 |",
		"| gofmt | 100% |",
		"| go_vet | 90% |",
		"## Issues",
		"### go_vet",
		"[`a.go`](https://example.com/a.go)",
		"Line 12: bad thing",
		"- See [https://example.com/pkg](https://example.com/pkg)",
		"- `b.go`",
	} {
		if !strings.Contains(md, want) {
			t.Errorf("report missing %q\n%s", want, md)
		}
	}
}

func TestGenerateReportMarkdownNoIssues(t *testing.T) {
	md := GenerateReportMarkdown(ChecksResult{
		Average: 1,
		Grade:   GradeAPlus,
		Checks:  []Score{{Name: "gofmt", Percentage: 1}},
	})
	if !strings.Contains(md, "No issues found. Nice work!") {
		t.Errorf("expected clean report to say no issues\n%s", md)
	}
	if strings.Contains(md, "## Issues") {
		t.Errorf("clean report should not contain an Issues section\n%s", md)
	}
}
