package main

import (
	"github.com/spf13/cobra"
	"github.com/timur-developer/gcviz/internal/config"
)

var version = "dev"

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "gcviz",
		Short:        "TUI visualizer for Go GC",
		Long:         "gcviz is a TUI visualizer for Go GC behavior.",
		SilenceUsage: true,
	}

	cmd.SetVersionTemplate("gcviz version {{.Version}}\n")
	cmd.Version = version

	cmd.PersistentFlags().Int("window-size", config.DefaultWindowSize, "Number of recent samples to keep in memory")
	cmd.PersistentFlags().String("snapshot-path", config.DefaultSnapshotDir(), "Path to write snapshot files")
	cmd.PersistentFlags().Bool("no-alt-screen", false, "Disable terminal alternate screen buffer")
	cmd.PersistentFlags().Int64("stw-warn-us", config.DefaultSTWWarnUs, "STW warning threshold (microseconds)")
	cmd.PersistentFlags().Int64("stw-bad-us", config.DefaultSTWBadUs, "STW bad threshold (microseconds)")

	cmd.AddCommand(newRunCmd(), newAttachCmd(), newLabCmd(), newDiffCmd())

	return cmd
}
