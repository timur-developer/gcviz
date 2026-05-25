package config

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestLoad_NoAltScreen_FromEnv(t *testing.T) {
	t.Setenv("GCVIZ_NO_ALT_SCREEN", "true")

	cmd := &cobra.Command{Use: "gcviz"}
	cmd.Flags().Bool("no-alt-screen", false, "")

	cfg, err := Load(cmd, nil)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if !cfg.NoAltScreen {
		t.Fatalf("NoAltScreen=false, want true")
	}
}

func TestLoad_NoAltScreen_FromFlag(t *testing.T) {
	cmd := &cobra.Command{Use: "gcviz"}
	cmd.Flags().Bool("no-alt-screen", false, "")
	if err := cmd.Flags().Set("no-alt-screen", "true"); err != nil {
		t.Fatalf("set flag error: %v", err)
	}

	cfg, err := Load(cmd, nil)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if !cfg.NoAltScreen {
		t.Fatalf("NoAltScreen=false, want true")
	}
}

