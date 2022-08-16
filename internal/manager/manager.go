package manager

import (
	"fmt"
	"strings"

	"github.com/go-logr/logr"

	"github.com/joshvanl/yazbu/internal/client"
	"github.com/joshvanl/yazbu/internal/config"
	"github.com/joshvanl/yazbu/internal/util"
)

// Manager is the database manager for a set of buckets over a set of
// filesystems.
type Manager struct {
	// log is the manager logger
	log logr.Logger

	// filesystems is the set of ZFS dataset filesystems to backup.
	filesystems []string

	// clients is the set of real S3 clients to backup data.
	clients []*client.Client
}

// New creates a new Database manager for backups. Assumes the given config is
// correct.
// Force is used to signal that cadence should be overridden if the remote
// Cadence is configured differently to the local. Should only be used by users
// if they know what they are doing!
func New(log logr.Logger, io util.IO, cfg config.Config, force bool) (*Manager, error) {
	log = log.WithName("manager")

	var (
		clients []*client.Client
		errs    []string
	)

	// Create a client for each S3 endpoint bucket.
	for _, bucket := range cfg.Buckets {
		cl, err := client.New(client.Options{
			Log:         log,
			IO:          io,
			Filesystems: cfg.Filesystems,
			Cadence:     cfg.Cadence,
			Bucket:      bucket,
			Force:       force,
		})
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}
		clients = append(clients, cl)
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("%s", strings.Join(errs, ", "))
	}

	return &Manager{
		log:         log,
		filesystems: cfg.Filesystems,
		clients:     clients,
	}, nil
}
