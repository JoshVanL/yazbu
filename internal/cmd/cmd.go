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
		Short: "Yet Another ZFS BackerUper.",
		Long:  "yazbu is a Yet Another ZFS BackerUper. Designed to be a simple and easy to use tool for backing up ZFS filesystems. Backups decay over time using a configurable decay rate.",
	}

	for _, registerCmd := range commands.Commands() {
		cmds.AddCommand(registerCmd(ctx, io))
	}

	return cmds
}
