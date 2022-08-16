package zfs

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/go-logr/logr"
)

// ZFSReader is a function that returns a reader for a zfs snapshot, configured
// with the given logger.
type ZFSReader func(context.Context, logr.Logger) (io.ReadCloser, error)

// SnapshotCreate creates a snapshot of the given filesystem. Returns the name
// of the zfs snapshot, and its size.
func SnapshotCreate(ctx context.Context, log logr.Logger, filesystem string) (string, uint64, error) {
	log = log.WithName("zfs_create_snapshot")
	now := time.Now().UTC()

	snapshot := fmt.Sprintf("%s@yazbu_%04d-%02d-%02d_%02d-%02d-%02d",
		filesystem,
		now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second(),
	)

	log.Info("creating snapshot", "snapshot", snapshot)
	cmd := exec.CommandContext(ctx, "zfs", "snapshot", snapshot)
	cmd.Stdout, cmd.Stderr = logWriter(log, logStdout), logWriter(log, logStderr)

	if err := cmd.Run(); err != nil {
		return snapshot, 0, err
	}

	size, err := SnapshotSize(ctx, log, snapshot)
	if err != nil {
		return "", 0, err
	}

	return snapshot, size, nil
}

// SnapshotSendFull sends the given zfs full snapshot to the returned reader.
func SnapshotSend(ctx context.Context, log logr.Logger, snapshot string) (ZFSReader, error) {
	log = log.WithName("zfs_send_full")
	log.Info("sending snapshot", "snapshot", snapshot)

	cmd := exec.CommandContext(ctx, "zfs", "send", "--raw", snapshot)
	cmd.Stderr = logWriter(log, logStderr)

	rc, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	return func(ctx context.Context, log logr.Logger) (io.ReadCloser, error) {
		if err := cmd.Start(); err != nil {
			return nil, fmt.Errorf("failed to start sending full snapshot: %w", err)
		}

		go func() {
			if err := cmd.Wait(); err != nil {
				log.Error(err, "error sending full snapshot")
			}
		}()

		return rc, nil
	}, nil
}

// SnapshotSendInc sends the given zfs incremental snapshot to the returned
// reader.
func SnapshotSendInc(ctx context.Context, log logr.Logger, fromSnapshot, toSnapshot string) (ZFSReader, error) {
	log = log.WithName("zfs_send_inc")

	log.Info("sending incremental snapshot", "from", fromSnapshot, "to", toSnapshot)
	cmd := exec.CommandContext(ctx, "zfs", "send", "--raw", "-i", fromSnapshot, toSnapshot)
	cmd.Stderr = logWriter(log, logStderr)

	rc, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	return func(ctx context.Context, log logr.Logger) (io.ReadCloser, error) {
		if err := cmd.Start(); err != nil {
			return nil, fmt.Errorf("failed to start sending incremental snapshot: %w", err)
		}

		go func() {
			if err := cmd.Wait(); err != nil {
				log.Error(err, "error sending incremental snapshot")
			}
		}()

		return rc, nil
	}, nil
}

// SnapshotSize returns the size of the given zfs snapshot.
func SnapshotSize(ctx context.Context, log logr.Logger, snapshot string) (uint64, error) {
	log = log.WithName("zfs_size")

	cmd := exec.CommandContext(ctx, "zfs", "send", "--raw", "--parsable", "--dryrun", snapshot)
	cmd.Stderr = logWriter(log, logStderr)

	rc, err := cmd.StdoutPipe()
	if err != nil {
		return 0, err
	}

	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("failed to start getting size of snapshot: %w", err)
	}

	b, err := io.ReadAll(rc)
	if err != nil {
		return 0, err
	}

	fields := strings.Fields(string(b))
	size, err := strconv.ParseUint(fields[len(fields)-1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse size of snapshot: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		log.Error(err, "error getting size of snapshot")
	}

	return size, nil
}
