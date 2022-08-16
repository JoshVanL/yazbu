package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/joshvanl/yazbu/internal/config"
	"github.com/joshvanl/yazbu/internal/util"
)

// New returns a new config init command.
func newInit(ctx context.Context, io util.IO) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create a new empty config file. Buckets and Filesystems will still need to be manually edited.",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := configFileAbs()
			if err != nil {
				return err
			}

			_, err = os.Stat(path)
			if err == nil {
				return fmt.Errorf("config file already exists: %s", path)
			}
			if !errors.Is(err, os.ErrNotExist) {
				return err
			}

			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return err
			}

			f, err := os.Create(path)
			if err != nil {
				return err
			}

			enc := yaml.NewEncoder(f)
			enc.SetIndent(2)
			if err := enc.Encode(new(config.Config).DefaultValues()); err != nil {
				return err
			}

			fmt.Fprintf(io.Out, "Written new config file at %s\n", path)

			return nil
		},
	}

	return cmd
}
