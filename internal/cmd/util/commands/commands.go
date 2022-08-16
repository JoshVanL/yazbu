package commands

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/joshvanl/yazbu/internal/cmd/backup"
	"github.com/joshvanl/yazbu/internal/cmd/config"
	"github.com/joshvanl/yazbu/internal/cmd/list"
	"github.com/joshvanl/yazbu/internal/util"
)

// Commands returns the cobra Commands that should be registered for the CLI
// build.
func Commands() []func(context.Context, util.IO) *cobra.Command {
	return []func(context.Context, util.IO) *cobra.Command{
		backup.New,
		list.New,
		config.New,
	}
}
