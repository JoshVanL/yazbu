package backup

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/joshvanl/yazbu/internal/config"
)

// Type is the type of a backup entry.
type Type string

const (
	// TypeIncremental is a backup of an incremental snapshot. Effectively a diff
	// of the previous snapshot, using ZFS magic.
	TypeIncremental Type = "inc"

	// TypeFull is a backup of a ZFS dataset in its entirety. Can be considered a
	// full copy of that file system.
	TypeFull Type = "full"
)

// DB is a database file of an instance of Entries for a ZFS file system
// dataset.
type DB struct {
	// Endpoint is the URL where the S3 compatible server is location for this
	// database.
	Endpoint string `json:"endpoint,omitempty"`

	// Bucket is the S3 bucket of this database.
	Bucket string `json:"bucket,omitempty"`

	// Filesystem is the path to the zfs dataset these snapshot entries were
	// taken from.
	Filesystem string `json:"filesystem"`

	// Cadence is the cadence for this database.
	Cadence config.Cadence `json:"cadence"`

	// Entries is the set of backup entry instances that belong to this dataset.
	Entries []Entry `json:"entries"`
}

// Entry is a reference to a backup file for a particular ZFS file system
// dataset.
type Entry struct {
	// ID is the unique ID number for this backup entry. ID should be unique for
	// this Filesystem, and should always have a higher number than the parent
	// that proceeded it.
	ID int `json:"id"`

	// Parent is the ID of the Parent Entry who this Entry was written next for
	// the same Filesystem.
	Parent int `json:"parent"`

	// Timestamp is the time at which this backup entry snapshot was taken on the
	// host machine.
	Timestamp time.Time `json:"timestamp"`

	// Type is the type of the snapshot, i.e an incremental or full ZFS snapshot.
	Type Type `json:"backupType"`

	// S3Key is the remote S3 path key where this entry was written to.
	S3Key string `json:"s3Key"`

	// Size is the number of bytes this Entry.
	Size uint64 `json:"size"`
}

// Parse parses a full Database file using the given reader. Returns the
// entries in ascending age order.
func Parse(r io.Reader) (DB, error) {
	var db DB
	if err := json.NewDecoder(r).Decode(&db); err != nil {
		return DB{}, fmt.Errorf("failed to decode backup database file: %w", err)
	}

	sort.SliceStable(db.Entries, func(i, j int) bool {
		return db.Entries[i].ID < db.Entries[j].ID
	})

	return db, nil
}
