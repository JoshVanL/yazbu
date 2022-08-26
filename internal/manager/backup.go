package manager

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/joshvanl/yazbu/internal/client"
	"github.com/joshvanl/yazbu/internal/zfs"
)

// BackupFull create a full ZFS backup for each filesystem, and writes those
// backups to all S3 endpoints, updating their respective databases.
func (m *Manager) BackupFull(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	m.log.Info("performing full backup")

	var (
		errs []string
		wg   sync.WaitGroup
		lock sync.Mutex
	)

	wg.Add(len(m.filesystems))
	for _, fs := range m.filesystems {
		go func(fs string) {
			defer wg.Done()

			if err := m.backupFullFS(ctx, fs); err != nil {
				lock.Lock()
				defer lock.Unlock()
				errs = append(errs, err.Error())
				cancel()
			}
		}(fs)
	}
	wg.Wait()

	if len(errs) > 0 {
		return fmt.Errorf("backupFull: [%s]", strings.Join(errs, ", "))
	}

	return nil
}

// backupFullFS creates a backup in all buckets, for the given filesystem.
func (m *Manager) backupFullFS(ctx context.Context, fs string) error {
	snapshot, size, err := zfs.SnapshotCreate(ctx, m.log, fs)
	if err != nil {
		return fmt.Errorf("failed to create full snapshot: %w", err)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		errs []string
		wg   sync.WaitGroup
		lock sync.Mutex
	)

	wg.Add(len(m.clients))
	for _, cl := range m.clients {
		go func(cl *client.Client) {
			defer wg.Done()

			f := func() error {
				rc, err := zfs.SnapshotSendFull(ctx, m.log, snapshot)
				if err != nil {
					return fmt.Errorf("failed to send snapshot: %w", err)
				}

				split := strings.Split(snapshot, "@")
				if err := cl.BackupWriteFull(ctx,
					filepath.Join(split[0], fmt.Sprintf("%s.full", split[1])),
					size, rc,
				); err != nil {
					return err
				}

				return nil
			}

			if err := f(); err != nil {
				lock.Lock()
				defer lock.Unlock()
				errs = append(errs, err.Error())
				cancel()
			}
		}(cl)
	}
	wg.Wait()

	if len(errs) > 0 {
		return fmt.Errorf("backupFullFS %q: [%s]", fs, strings.Join(errs, ", "))
	}

	return nil
}
