package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path"
	"reflect"
	"sort"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/go-logr/logr"

	"github.com/joshvanl/yazbu/internal/backup"
	"github.com/joshvanl/yazbu/internal/client/progress"
	"github.com/joshvanl/yazbu/internal/util"
	"github.com/joshvanl/yazbu/internal/zfs"
)

// fsclient is a filesystem client, responsible for running database and backup
// operations on a single S3 bucket and filesystem.
type fsclient struct {
	// log is the logger for fsclient.
	log logr.Logger

	// io references files to output to the terminal.
	io util.IO

	// Client is the S3 client to use for operations.
	*Client

	// filesystem is the filesystem to backup to S3 buckets.
	filesystem string

	// dbKey is the filepath or "key" to the database file object.
	dbKey string

	// force will overwrite the cadence if their is a difference.
	force bool

	// lock gates concurrent access to the database file.
	lock sync.Mutex
}

// writeFull writes the full backup of the filesystem to the S3 bucket. Updates
// the database file with
func (f *fsclient) writeFull(ctx context.Context, db backup.DB, key string, size uint64, rc zfs.ZFSReader) (fsrunner, error) {
	log := f.log.WithName(key)

	f.lock.Lock()
	defer f.lock.Unlock()

	return func(ctx context.Context) error {
		log.Info("writing full backup")

		reader, err := rc(ctx, log)
		if err != nil {
			return err
		}

		progress := progress.New(path.Join(f.bucket, f.filesystem, key), size, reader)

		if _, err := f.uploader.Upload(&s3manager.UploadInput{
			Bucket:       aws.String(f.bucket),
			Key:          aws.String(key),
			Body:         progress,
			StorageClass: aws.String(f.storageClass),
		}); err != nil {
			return fmt.Errorf("failed to create full backup %q: %w", key, err)
		}

		var parent int
		for _, entry := range db.Entries {
			if entry.ID > parent {
				parent = entry.ID
			}
		}

		db.Entries = append(db.Entries, backup.Entry{
			ID:        parent + 1,
			Parent:    parent,
			Timestamp: time.Now(),
			Type:      backup.TypeFull,
			S3Key:     key,
			Size:      size,
		})

		log.Info("updating database file", "db_file", f.dbKey)

		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(&db); err != nil {
			return err
		}

		if _, err := f.uploader.Upload(&s3manager.UploadInput{
			Bucket:       aws.String(f.bucket),
			Key:          aws.String(f.dbKey),
			Body:         &buf,
			StorageClass: aws.String("STANDARD"),
		}); err != nil {
			return fmt.Errorf("failed to create db file %q: %w", f.dbKey, err)
		}
		return nil
	}, nil
}

// getDB returns the database file from the bucket.
func (f *fsclient) getDB(ctx context.Context) (backup.DB, error) {
	if err := f.ensureDBFile(ctx); err != nil {
		return backup.DB{}, err
	}

	out, err := f.s3.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: aws.String(f.bucket),
		Key:    aws.String(f.dbKey),
	})
	if err != nil {
		return backup.DB{}, fmt.Errorf("failed to get bucket database file %q: %w", f.bucket, err)
	}

	db, err := backup.Parse(out.Body)
	if err != nil {
		return backup.DB{}, err
	}

	if !reflect.DeepEqual(db, backup.DB{}) && !reflect.DeepEqual(db.Cadence, f.cadence) {
		if !f.force {
			return backup.DB{}, fmt.Errorf(
				`local cadence mismatches with remote, use --force to ignore and overwrite.
WARNING: doing so is incredibly dangerous since you may end up deleting old backups you want. Make sure you know what you are doing before you do this.
local:
%s

remote:
%s
`,
				f.cadence.ToJSON(), db.Cadence.ToJSON())
		} else {
			fmt.Fprintf(f.io.Err, "WARNING: remote cadence does not match that locally. Will be overwritten\nlocal:\n%s\nremote:\n%s\n", f.cadence.ToJSON(), db.Cadence.ToJSON())
		}
	}

	sort.SliceStable(db.Entries, func(i, j int) bool {
		return db.Entries[i].ID < db.Entries[j].ID
	})

	return backup.DB{
		Endpoint:   f.s3.Endpoint,
		Bucket:     f.bucket,
		Filesystem: f.filesystem,
		Cadence:    db.Cadence,
		Entries:    db.Entries,
	}, nil
}

// tidyDB removes old entries from the database file which no longer exist.
func (f *fsclient) tidyDB(ctx context.Context, db backup.DB) error {
	return nil
}

// ensureDBFiles ensures that the database file exists in the bucket
// filesystem.
func (f *fsclient) ensureDBFile(ctx context.Context) error {
	_, err := f.s3.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(f.bucket),
		Key:    aws.String(f.dbKey),
	})

	// s3.ErrCodeNoSuchKey does not work, aws is missing this error code so we
	// hardwire a string.
	if aerr, ok := err.(awserr.Error); ok && aerr.Code() == "NotFound" {
		f.log.Info("db file does not exist, writing", "db_file", f.dbKey)

		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(backup.DB{}); err != nil {
			return err
		}

		if _, err := f.uploader.Upload(&s3manager.UploadInput{
			Bucket:       aws.String(f.bucket),
			Key:          aws.String(f.dbKey),
			Body:         &buf,
			ContentType:  aws.String("application/json"),
			StorageClass: aws.String("STANDARD"),
		}); err != nil {
			return fmt.Errorf("failed to create db file %q: %w", f.dbKey, err)
		}

		return nil
	}

	return err
}
