package config

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/joshvanl/yazbu/config"
	"github.com/joshvanl/yazbu/internal/util"
)

// New returns a new config validate command.
func newValidate(ctx context.Context, io util.IO) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate the config file.",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := configFileAbs()
			if err != nil {
				return err
			}

			if _, err := config.ReadFile(path); err != nil {
				return err
			}

			fmt.Fprintf(io.Out, "Config file valid at %s\n", path)

			return nil
		},
	}

	return cmd
}
