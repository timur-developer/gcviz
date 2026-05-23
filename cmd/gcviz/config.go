package main

import (
	"github.com/spf13/cobra"

	"github.com/timur-developer/gcviz/internal/config"
)

type Config = config.Config

type RunConfig = config.RunConfig

type AttachConfig = config.AttachConfig

type LabConfig = config.LabConfig

type DiffConfig = config.DiffConfig

func Default() Config {
	return config.Default()
}

func Load(cmd *cobra.Command, args []string) (Config, error) {
	return config.Load(cmd, args)
}
