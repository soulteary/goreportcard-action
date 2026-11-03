package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	if os.Getenv("GRC_CLI_HELPER") == "1" {
		if strings.TrimSuffix(filepath.Base(os.Args[0]), ".exe") == "wc" {
			fmt.Println("5")
		} else {
			fmt.Println("a.go:1:1: issue")
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}
func cliTools(t *testing.T) {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(exe)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	for _, name := range []string{"go", "gofmt", "gocyclo", "ineffassign", "misspell", "wc"} {
		if runtime.GOOS == "windows" {
			name += ".exe"
		}
		if err = os.WriteFile(filepath.Join(dir, name), b, 0700); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("GRC_CLI_HELPER", "1")
}
func TestCLIEntryAndArtifactErrors(t *testing.T) {
	cliTools(t)
	root := sampleRepo(t)
	for _, args := range [][]string{{"-h"}, {"-d", root, "-v"}, {"-d", root, "-j"}, {"-d", root, "-t", "0"}} {
		var out, errout bytes.Buffer
		if code := execute(args, &out, &errout); code != 0 {
			t.Fatalf("%v: %d %s%s", args, code, &out, &errout)
		}
	}
	if execute([]string{"-unknown"}, io.Discard, io.Discard) != 2 {
		t.Fatal("invalid flag accepted")
	}
	for _, opts := range []options{{dir: root, svg: filepath.Join(root, "missing", "b.svg")}, {dir: root, report: filepath.Join(root, "missing", "r.md")}} {
		if run(opts, io.Discard) != 1 {
			t.Fatal("write error ignored")
		}
	}
	var out bytes.Buffer
	if run(options{dir: root, verbose: true}, &out) != 0 || !strings.Contains(out.String(), "Line 1:") {
		t.Fatal(out.String())
	}
}
