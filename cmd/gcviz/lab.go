package main

import (
	"github.com/spf13/cobra"
)

func newLabCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lab",
		Short: "Run a built-in demo workload",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := Load(cmd, args)
			return err
		},
	}

	cmd.Flags().String("preset", "", "Lab preset name")

	return cmd
}
