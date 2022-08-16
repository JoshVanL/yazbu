package options

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
	"github.com/spf13/cobra"

	"github.com/joshvanl/yazbu/internal/config"
	"github.com/joshvanl/yazbu/internal/manager"
	"github.com/joshvanl/yazbu/internal/util"
)

// Options are global options used to configure backups.
type Options struct {
	// configPath is the filepath location to the config file.
	configPath string

	// Log is the shared logger for yazbu.
	Log logr.Logger

	// Config is the parsed and validated config.
	Config *config.Config

	// Manager is the configured manager which are used to perform operations.
	Manager *manager.Manager

	// force indicates that the cadence should be overriden
	force bool
}

// New constructs a new shared Options.
func New(ctx context.Context, io util.IO, cmd *cobra.Command) *Options {
	o := new(Options)

	cmd.PersistentFlags().StringVarP(&o.configPath, "config", "c", "~/.config/yazbu/config.yaml", "File path location to the yaml config.")
	cmd.PersistentFlags().BoolVar(&o.force, "force", false,
		"If the local Cadence is different from the remote discovered one then it will be overridden. WARNING: doing so is incredibly dangerous since you may end up deleting old backups you want. Make sure you know what you are doing before you do this.")

	// Setup a PreRun to populate the Factory. Catch the existing PreRun command
	// if one was defined, and execute it second.
	existingPreRun := cmd.PreRunE
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if err := o.complete(io, cmd); err != nil {
			return err
		}
		if existingPreRun != nil {
			return existingPreRun(cmd, args)
		}

		return nil
	}

	return o
}

// complete defaults and validates the command options.
func (o *Options) complete(io util.IO, cmd *cobra.Command) error {
	var err error

	o.Log = stdr.New(log.New(os.Stdout, "", log.Lshortfile))
	o.Log = o.Log.WithName("yazbu")

	if !cmd.Flag("config").Changed {
		o.configPath = os.Getenv("YAZBU_CONFIG")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	path := o.configPath
	if strings.HasPrefix(path, "~/") {
		path = filepath.Join(homeDir, path[2:])
	}

	o.Config, err = config.ReadFile(path)
	if err != nil {
		return err
	}

	o.Manager, err = manager.New(o.Log, io, *o.Config, o.force)
	if err != nil {
		return err
	}

	return nil
}
