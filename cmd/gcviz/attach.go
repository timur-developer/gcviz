package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/timur-developer/gcviz/internal/source/collector"
	"github.com/timur-developer/gcviz/internal/ui"
)

func newAttachCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attach <url>",
		Short: "Attach to a running service",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := Load(cmd, args)
			if err != nil {
				return err
			}
			if cfg.Attach.URL == "" {
				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "missing url")
				return ExitError{Code: 2, Err: errors.New("missing url")}
			}

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			c := collector.NewCollector(cfg.Attach.URL, cfg.Attach.PollInterval, nil)
			if err := c.Start(ctx); err != nil {
				return err
			}

			model := ui.NewModel(ctx, cancel, cfg.WindowSize)
			prog := tea.NewProgram(model, tea.WithAltScreen())

			go func() {
				for ev := range c.Events() {
					prog.Send(ui.GCEventMsg{Event: ev, At: time.Now()})
				}
			}()
			go func() {
				for range c.Errors() {
				}
			}()

			progErrCh := make(chan error, 1)
			go func() {
				_, err := prog.Run()
				progErrCh <- err
			}()

			waitErr := c.Wait()
			cancel()
			uiErr := <-progErrCh

			if uiErr != nil && !errors.Is(uiErr, tea.ErrProgramKilled) {
				return uiErr
			}
			return waitErr
		},
	}

	cmd.Flags().String("url", "", "Runtime metrics endpoint URL")
	cmd.Flags().Duration("poll-interval", time.Second, "Polling interval for runtime metrics")

	return cmd
}
