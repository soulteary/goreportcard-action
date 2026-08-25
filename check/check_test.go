package check

import (
	"os/exec"
	"testing"
)

// toolInstalled reports whether a required external linter is available on
// $PATH. The report-card checks (misspell, gocyclo, ineffassign, ...) shell out
// to third-party binaries that are optional in local and CI environments, so a
// unit test must not assume they are present.
func toolInstalled(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func TestRun(t *testing.T) {
	cr, err := Run("testdata/testrepo@v0.1.0", false)
	if err != nil {
		t.Fatal(err)
	}

	// The License check always contributes one issue for the testrepo (it has
	// no LICENSE file, reported with an empty filename). Every other external
	// tool only adds to the count when its binary is installed, so compute the
	// expectation from what is actually available on $PATH.
	want := 1
	if toolInstalled("misspell") {
		// testrepo/a.go contains a deliberate misspelling of "obvious".
		want++
	}

	if cr.Issues != want {
		t.Errorf("got cr.Issues = %d, want %d", cr.Issues, want)
	}
}
