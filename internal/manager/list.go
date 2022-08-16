package manager

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/joshvanl/yazbu/internal/backup"
	"github.com/joshvanl/yazbu/internal/client"
)

// ListDBs lists all databases for all buckets.
func (m *Manager) ListDBs(ctx context.Context) (map[string][]backup.DB, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		errs      []string
		wg        sync.WaitGroup
		lock      sync.Mutex
		clientDBs []backup.DB
	)

	wg.Add(len(m.clients))
	for _, cl := range m.clients {
		go func(cl *client.Client) {
			defer wg.Done()

			dbs, err := cl.ListDBs(ctx)

			lock.Lock()
			defer lock.Unlock()

			if err != nil {
				errs = append(errs, err.Error())
				cancel()
				return
			}

			clientDBs = append(clientDBs, dbs...)
		}(cl)
	}
	wg.Wait()

	if len(errs) > 0 {
		return nil, fmt.Errorf("ListDBs: [%s]", strings.Join(errs, ", "))
	}

	fsBackupList := make(map[string][]backup.DB)
	for _, db := range clientDBs {
		fsBackupList[db.Filesystem] = append(fsBackupList[db.Filesystem], db)
	}

	for fs, dbs := range fsBackupList {
		sort.SliceStable(dbs, func(i, j int) bool {
			if dbs[i].Endpoint == dbs[j].Endpoint {
				return dbs[i].Bucket < dbs[j].Bucket
			}
			return dbs[i].Endpoint < dbs[j].Endpoint
		})

		fsBackupList[fs] = dbs
	}

	return fsBackupList, nil
}
