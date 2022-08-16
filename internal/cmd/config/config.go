package config

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/joshvanl/yazbu/internal/util"
)

var configFile string

// New returns a new config command.
func New(ctx context.Context, io util.IO) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Create a new empty config file.",
	}

	cmd.PersistentFlags().StringVarP(&configFile, "config", "c", "~/.config/yazbu/config.yaml", "File path location to the yaml config.")

	cmd.AddCommand(newInit(ctx, io))
	cmd.AddCommand(newValidate(ctx, io))

	return cmd
}

func configFileAbs() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	path := configFile
	if strings.HasPrefix(path, "~/") {
		path = filepath.Join(homeDir, path[2:])
	}

	return path, nil
}
