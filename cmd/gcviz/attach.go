package main

import (
	"time"

	"github.com/spf13/cobra"
)

func newAttachCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attach",
		Short: "Attach to a running service",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := Load(cmd, args)
			return err
		},
	}

	cmd.Flags().String("url", "", "Runtime metrics endpoint URL")
	cmd.Flags().Duration("poll-interval", time.Second, "Polling interval for runtime metrics")

	return cmd
}
