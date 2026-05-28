package main

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestRunCmd_MissingTarget(t *testing.T) {
	cmd := newRunCmd()

	var stderr bytes.Buffer
	cmd.SetOut(io.Discard)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	var ee ExitError
	if !AsExitError(err, &ee) {
		t.Fatalf("expected ExitError, got %v", err)
	}
	if ee.Code != 2 {
		t.Fatalf("expected exit code 2, got %d", ee.Code)
	}
	if !strings.Contains(stderr.String(), "missing target") {
		t.Fatalf("expected missing target message, got %q", stderr.String())
	}
}
