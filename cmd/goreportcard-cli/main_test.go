package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// sampleRepo copies a tiny Go source tree into a temp dir whose path does not
// contain any of the checker's skipped directory names (e.g. "testdata"),
// which would otherwise cause every file to be filtered out.
func sampleRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	src := "package main\n\nimport \"fmt\"\n\nfunc main() {\n\t// obivous typo\n\tfmt.Println(\"invalid %s\")\n}\n"
	if err := os.WriteFile(filepath.Join(dir, "a.go"), []byte(src), 0o644); err != nil {
		t.Fatalf("writing sample repo: %v", err)
	}
	return dir
}

func TestDotPrintf(t *testing.T) {
	var buf bytes.Buffer
	dotPrintf(&buf, 24, "Grade", "%s %.1f%%", "A+", 96.5)
	got := buf.String()
	if !strings.HasPrefix(got, "Grade ") || !strings.HasSuffix(got, "A+ 96.5%\n") {
		t.Errorf("unexpected dotPrintf output: %q", got)
	}
	if !strings.Contains(got, "...") {
		t.Errorf("expected dot fill, got %q", got)
	}

	// When the content is wider than fullLen, at least two dots still appear.
	buf.Reset()
	dotPrintf(&buf, 4, "averyverylongname", "%d", 100)
	if !strings.Contains(buf.String(), "..") {
		t.Errorf("expected minimum two dots, got %q", buf.String())
	}
}

func TestRunText(t *testing.T) {
	var buf bytes.Buffer
	code := run(options{dir: sampleRepo(t)}, &buf)
	if code != 0 {
		t.Fatalf("run exit code = %d, want 0", code)
	}
	out := buf.String()
	for _, want := range []string{"Grade", "Files", "Issues", "gofmt"} {
		if !strings.Contains(out, want) {
			t.Errorf("run output missing %q\n%s", want, out)
		}
	}
}

func TestRunJSON(t *testing.T) {
	var buf bytes.Buffer
	code := run(options{dir: sampleRepo(t), jsn: true}, &buf)
	if code != 0 {
		t.Fatalf("run exit code = %d, want 0", code)
	}
	var parsed map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("run -j did not emit valid JSON: %v\n%s", err, buf.String())
	}
	if _, ok := parsed["GradeFromPercentage"]; !ok {
		t.Errorf("JSON output missing grade field: %s", buf.String())
	}
}

func TestRunWritesArtifacts(t *testing.T) {
	tmp := t.TempDir()
	svgPath := filepath.Join(tmp, "badge.svg")
	reportPath := filepath.Join(tmp, "report.md")

	var buf bytes.Buffer
	code := run(options{
		dir:    sampleRepo(t),
		svg:    svgPath,
		report: reportPath,
	}, &buf)
	if code != 0 {
		t.Fatalf("run exit code = %d, want 0", code)
	}

	svg, err := os.ReadFile(svgPath)
	if err != nil {
		t.Fatalf("reading badge: %v", err)
	}
	if !strings.Contains(string(svg), "<svg") {
		t.Errorf("badge does not look like SVG: %s", svg)
	}

	report, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("reading report: %v", err)
	}
	if !strings.Contains(string(report), "# Go Report Card") {
		t.Errorf("report missing heading: %s", report)
	}
}

func TestRunThreshold(t *testing.T) {
	var buf bytes.Buffer
	// A threshold above any achievable score must force a non-zero exit.
	code := run(options{dir: sampleRepo(t), th: 101}, &buf)
	if code != 1 {
		t.Errorf("run with impossible threshold exit code = %d, want 1", code)
	}
}

func TestRunError(t *testing.T) {
	var buf bytes.Buffer
	code := run(options{dir: t.TempDir()}, &buf) // empty dir -> no .go files
	if code != 1 {
		t.Errorf("run on empty dir exit code = %d, want 1", code)
	}
}
