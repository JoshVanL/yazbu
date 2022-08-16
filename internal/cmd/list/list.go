package list

import (
	"context"
	"fmt"
	"os"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"

	"github.com/joshvanl/yazbu/internal/cmd/util/options"
	"github.com/joshvanl/yazbu/internal/cmd/util/table"
	"github.com/joshvanl/yazbu/internal/util"
)

// list is the list command.
type list struct {
	util.IO

	// options is the command options.
	options *options.Options
}

// New returns a new list command.
func New(ctx context.Context, io util.IO) *cobra.Command {
	b := list{IO: io}

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "TODO",
		Long:    "TODO",
		Example: "",
		RunE: func(cmd *cobra.Command, args []string) error {
			fsDBs, err := b.options.Manager.ListDBs(ctx)
			if err != nil {
				fmt.Fprintf(io.Err, "%s\n", err)
				os.Exit(1)
			}

			tbl := table.NewBuilder([]string{"dataset", "endpoint", "bucket", "id", "parent", "type", "path", "size", "timestamp"})

			for fs, dbs := range fsDBs {
				if len(dbs) == 0 || len(dbs[0].Entries) == 0 {
					continue
				}

				for i, db := range dbs {
					if len(db.Entries) == 0 {
						continue
					}

					{
						entry := db.Entries[0]
						if i == 0 {
							tbl.AddRow(fs, db.Endpoint, db.Bucket, entry.ID, entry.Parent, entry.Type, entry.S3Key, humanize.Bytes(entry.Size), entry.Timestamp.UTC().String())
						} else {
							tbl.AddRow("", db.Endpoint, db.Bucket, entry.ID, entry.Parent, entry.Type, entry.S3Key, humanize.Bytes(entry.Size), entry.Timestamp.UTC().String())
						}
					}

					for _, entry := range db.Entries[1:] {
						tbl.AddRow("", "", "", entry.ID, entry.Parent, entry.Type, entry.S3Key, humanize.Bytes(entry.Size), entry.Timestamp.UTC().String())
					}

				}
			}

			return tbl.Build(io.Out)
		},
	}

	b.options = options.New(ctx, io, cmd)

	return cmd
}
