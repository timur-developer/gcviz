package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/timur-developer/gcviz/internal/source/runner"
	"github.com/timur-developer/gcviz/internal/ui"
)

func newRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run <target> [args...]",
		Short: "Run target under GC observation",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := Load(cmd, args)
			if err != nil {
				return err
			}

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			r := runner.NewRunner(cfg.Run.Target, cfg.Run.Args, nil)
			if err := r.Start(ctx); err != nil {
				return err
			}

			model := ui.NewModel(ctx, cancel, cfg.WindowSize)
			prog := tea.NewProgram(model, tea.WithAltScreen())

			go func() {
				for ev := range r.Events() {
					prog.Send(ui.GCEventMsg{Event: ev, At: time.Now()})
				}
			}()
			go func() {
				for range r.Stderr() {
				}
			}()
			go func() {
				for range r.ParseErrors() {
				}
			}()

			progErrCh := make(chan error, 1)
			go func() {
				_, err := prog.Run()
				progErrCh <- err
			}()

			waitErr := r.Wait()
			cancel()
			uiErr := <-progErrCh

			if uiErr != nil && !errors.Is(uiErr, tea.ErrProgramKilled) {
				return uiErr
			}
			return waitErr
		},
	}

	cmd.Flags().String("target", "", "Path to target binary")

	return cmd
}
