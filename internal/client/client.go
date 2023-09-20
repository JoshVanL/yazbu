package client

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/go-logr/logr"
	"k8s.io/utils/clock"

	"github.com/joshvanl/yazbu/internal/backup"
	"github.com/joshvanl/yazbu/config"
	"github.com/joshvanl/yazbu/internal/util"
	"github.com/joshvanl/yazbu/internal/zfs"
)

const (
	// fileBackup is the filename or object name which contains the database
	// data.
	keyFileBackup = "backup.db"
)

// fsrunner is a filesystem operation runner.
type fsrunner func(context.Context) error

// Options defines the options for creating Clients.
type Options struct {
	// Log is the logger for the Client.
	Log logr.Logger

	// Filesystems are the ZFS filesystems to backup.
	Filesystems []string

	// Cadence is the cadence of backups to be kept over time.
	Cadence config.Cadence

	// Bucket contains the configuration for the S3 bucket.
	Bucket config.Bucket

	// IO references files to write to the terminal.
	IO util.IO

	// Force instructs the client to overwrite the existing cadence if it differs
	// from the local config. Dangerous, and should only be done when the user
	// knows what they are doing.
	Force bool
}

// Client is the zfs backup client for a single S3 bucket.
type Client struct {
	// log is the client logger.
	log logr.Logger

	// cadence is the cadence of the backup.
	cadence config.Cadence

	// s3 is the s3 generic client.
	s3 *s3.S3

	// uploader is the s3 client to upload files.
	uploader *s3manager.Uploader

	// bucket is the name of the S3 bucket for this client.
	bucket string

	// storeageClass is the storage class to use for backup files. The database
	// file will remain as "STANDARD".
	storageClass string

	// fsclients the set of filesystem clients for this bucket, indexed by the
	// filesystem.
	fsclients map[string]*fsclient
}

// New creates a new client for this bucket. Constructs filesystem clients for
// all filesystems defined.
func New(opts Options) (*Client, error) {
	log := opts.Log.WithName(opts.Bucket.Endpoint).WithName(opts.Bucket.Name).WithName("client")

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(opts.Bucket.Region),
		Endpoint:    aws.String(opts.Bucket.Endpoint),
		Credentials: credentials.NewStaticCredentials(opts.Bucket.AccessKey, opts.Bucket.SecretKey, ""),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create s3 client session for %q: %w", opts.Bucket.Name, err)
	}

	c := &Client{
		log:          log.WithName(opts.Bucket.Endpoint).WithName(opts.Bucket.Name),
		cadence:      opts.Cadence,
		s3:           s3.New(sess),
		uploader:     s3manager.NewUploader(sess),
		bucket:       opts.Bucket.Name,
		storageClass: opts.Bucket.StorageClass,
		fsclients:    make(map[string]*fsclient),
	}

	for _, fs := range opts.Filesystems {
		c.fsclients[fs] = &fsclient{
			log:        log.WithName(fs),
			io:         opts.IO,
			Client:     c,
			filesystem: fs,
			dbKey:      filepath.Join(opts.Bucket.Name, fs, keyFileBackup),
			force:      opts.Force,
			clock:      clock.RealClock{},
		}
	}

	return c, nil
}

// ListDBs lists the databases in the bucket for each filesystem.
func (c *Client) ListDBs(ctx context.Context) ([]backup.DB, error) {
	var (
		errs []string
		wg   sync.WaitGroup
		lock sync.Mutex
		dbs  []backup.DB
	)

	wg.Add(len(c.fsclients))
	for _, fs := range c.fsclients {
		go func(fs *fsclient) {
			defer wg.Done()

			db, err := fs.getDB(ctx)
			lock.Lock()
			defer lock.Unlock()
			if err != nil {
				errs = append(errs, fmt.Sprintf("%s: %s", fs.filesystem, err.Error()))
				return
			}
			dbs = append(dbs, db)
		}(fs)
	}
	wg.Wait()

	if len(errs) > 0 {
		return nil, fmt.Errorf("ListDBs %q: [%s]", c.bucket, strings.Join(errs, ", "))
	}

	return dbs, nil
}

// BackupWriteFull writes a full backup to the bucket for every filesystem.
func (c *Client) BackupWriteFull(ctx context.Context, key string, size uint64, rc zfs.ZFSReader) error {
	var runners []fsrunner
	for _, fs := range c.fsclients {
		db, err := fs.getDB(ctx)
		if err != nil {
			return err
		}

		runner, err := fs.writeFull(ctx, db, key, size, rc)
		if err != nil {
			return err
		}

		runners = append(runners, func(ctx context.Context) error {
			if err := runner(ctx); err != nil {
				return err
			}
			return fs.executeCadence(ctx, db)
		})
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		errs []string
		wg   sync.WaitGroup
		lock sync.Mutex
	)

	wg.Add(len(runners))
	for _, runner := range runners {
		go func(runner fsrunner) {
			defer wg.Done()
			if err := runner(ctx); err != nil {
				lock.Lock()
				defer lock.Unlock()
				errs = append(errs, err.Error())
				cancel()
			}
		}(runner)
	}
	wg.Wait()

	if len(errs) > 0 {
		return fmt.Errorf("BackupWriteFull %q: [%s]", c.bucket, strings.Join(errs, ", "))
	}

	return nil
}
