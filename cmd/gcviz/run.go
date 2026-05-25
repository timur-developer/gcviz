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

	"github.com/timur-developer/gcviz/internal/config"
	"github.com/timur-developer/gcviz/internal/source/runner"
	"github.com/timur-developer/gcviz/internal/ui"
)

func newRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run <target> [args...]",
		Short: "Run target under GC observation",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(cmd, args)
			if err != nil {
				return err
			}
			if cfg.Run.Target == "" {
				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "missing target")
				return ExitError{Code: 2, Err: errors.New("missing target")}
			}

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			snapshotDir := cfg.SnapshotPath
			writer := snapshotWriter{dir: snapshotDir}

			stwTh := ui.STWThresholds{WarnUs: cfg.STWWarnUs, BadUs: cfg.STWBadUs}
			envInfo := &ui.TargetEnvInfo{
				GOGC:       os.Getenv("GOGC"),
				GOMEMLIMIT: os.Getenv("GOMEMLIMIT"),
				GODEBUG:    runner.NormalizeGODEBUG(os.Getenv("GODEBUG")),
			}
			model := ui.NewModel(ctx, cancel, cfg.WindowSize, snapshotDir, writer, stwTh, envInfo)
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
						snapErr := writeSnapshotOnExit(snapshotDir, m)
						if snapErr != nil {
							err = ExitError{Code: 1, Err: snapErr}
						}
					}
				}
				progErrCh <- err
			}()

			r := runner.NewRunner(cfg.Run.Target, cfg.Run.Args, nil)
			if err := r.Start(ctx); err != nil {
				cancel()
				uiErr := <-progErrCh
				if uiErr != nil && !errors.Is(uiErr, tea.ErrProgramKilled) {
					return uiErr
				}
				return err
			}

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
