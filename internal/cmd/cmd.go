package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/joshvanl/yazbu/internal/cmd/util/commands"
	"github.com/joshvanl/yazbu/internal/util"
)

func New(ctx context.Context, io util.IO) *cobra.Command {
	cmds := &cobra.Command{
		Use:   "yazbu",
		Short: "TODO",
		Long:  "TODO",
	}

	for _, registerCmd := range commands.Commands() {
		cmds.AddCommand(registerCmd(ctx, io))
	}

	return cmds
}
