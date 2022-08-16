package backup

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/joshvanl/yazbu/internal/cmd/util/options"
	"github.com/joshvanl/yazbu/internal/util"
)

// backup is the backup command.
type backup struct {
	util.IO

	// options is the command options.
	options *options.Options
}

// New constructs a new backup command.
func New(ctx context.Context, io util.IO) *cobra.Command {
	b := backup{IO: io}

	cmd := &cobra.Command{
		Use:     "backup",
		Short:   "TODO",
		Long:    "TODO",
		Example: "",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := b.options.Manager.BackupFull(ctx); err != nil {
				fmt.Fprintf(io.Err, "%s\n", err)
				os.Exit(1)
			}
			b.options.Log.Info("backup complete.")
			return nil
		},
	}

	b.options = options.New(ctx, io, cmd)

	return cmd
}
