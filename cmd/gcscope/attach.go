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

	"github.com/timur-developer/gcscope/internal/config"
	"github.com/timur-developer/gcscope/internal/source/collector"
	"github.com/timur-developer/gcscope/internal/ui"
)

func newAttachCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attach <url>",
		Short: "Attach to a running service",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(cmd, args)
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

			snapshotDir := cfg.SnapshotPath
			writer := snapshotWriter{dir: snapshotDir}

			stwTh := ui.STWThresholds{WarnUs: cfg.STWWarnUs, BadUs: cfg.STWBadUs}
			model := ui.NewModel(ctx, cancel, cfg.WindowSize, snapshotDir, writer, stwTh, nil)
			var prog *tea.Program
			if cfg.NoAltScreen {
				prog = tea.NewProgram(model)
			} else {
				prog = tea.NewProgram(model, tea.WithAltScreen())
			}

			progErrCh := make(chan error, 1)
			go func() {
				finalModel, err := prog.Run()
				if err == nil {
					if m, ok := finalModel.(ui.Model); ok {
						if cfg.ExitSnapshot {
							snapErr := writeSnapshotOnExit(snapshotDir, m)
							if snapErr != nil {
								err = ExitError{Code: 1, Err: snapErr}
							}
						}
					}
				}
				progErrCh <- err
			}()

			c := collector.NewCollector(cfg.Attach.URL, cfg.Attach.PollInterval, nil)
			if err := c.Start(ctx); err != nil {
				cancel()
				uiErr := <-progErrCh
				if uiErr != nil && !errors.Is(uiErr, tea.ErrProgramKilled) {
					return uiErr
				}
				return err
			}

			go func() {
				for ev := range c.Events() {
					prog.Send(ui.GCEventMsg{Event: ev, At: time.Now()})
				}
			}()
			go func() {
				for range c.Errors() {
				}
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
