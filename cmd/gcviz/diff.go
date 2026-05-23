package main

import (
	"github.com/spf13/cobra"
)

func newDiffCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Compare two snapshot files",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := Load(cmd, args)
			return err
		},
	}

	cmd.Flags().String("a", "", "Path to snapshot A")
	cmd.Flags().String("b", "", "Path to snapshot B")

	return cmd
}
