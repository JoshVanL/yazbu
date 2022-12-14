package client

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/joshvanl/yazbu/internal/backup"
)

// executeCadence will delete all backup entries which need to be deleted,
// according to the cadence.
func (f *fsclient) executeCadence(ctx context.Context, db backup.DB) error {
	f.log.Info("checking database to delete stale backups based on configured cadence...")

	markedForDeletion, err := f.markedForDeletion(db)
	if err != nil {
		return err
	}

	for _, entry := range markedForDeletion {
		log := f.log.WithValues("name", entry.ID, "timestamp", entry.Timestamp, "type", entry.Type)
		log.Info("deleting backup")
		if _, err := f.s3.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(f.bucket),
			Key:    aws.String(entry.S3Key),
		}); err != nil {
			return err
		}
		log.Info("backup deleted")
	}

	return nil
}

// markedForDeletion will return a list of all backups which need to be deleted
// according to the cadence.
func (f *fsclient) markedForDeletion(db backup.DB) ([]backup.Entry, error) {
	now := f.clock.Now()

	var (
		incrementals []backup.Entry
		lastFull     backup.Entry
		fullLast45   []backup.Entry
		full45to182  []backup.Entry
		full182to365 []backup.Entry
		full365Plus  = make(map[int][]backup.Entry)
	)

	for _, entry := range db.Entries {
		switch {
		case entry.Type == backup.TypeIncremental:
			incrementals = append(incrementals, entry)
			// Always continue to next, regardless of time. We will clean up orphaned
			// incremental backups later.
			continue

		case entry.Timestamp.After(now.Add(-time.Hour * 24 * 45)):
			fullLast45 = append(fullLast45, entry)

		case entry.Timestamp.After(now.Add(-time.Hour * 24 * 182)):
			full45to182 = append(full45to182, entry)

		case entry.Timestamp.After(now.Add(-time.Hour * 24 * 365)):
			full182to365 = append(full182to365, entry)

		default:
			// Get the year by mod this backup belongs to.
			full365Plus[entry.Timestamp.Year()] = append(full365Plus[entry.Timestamp.Year()], entry)
		}

		// Get the latest full backup.
		if lastFull.Timestamp.IsZero() || entry.Timestamp.After(lastFull.Timestamp) {
			lastFull = entry
		}
	}

	var markedForDeletion []backup.Entry

	var incrementalsSinceLast []backup.Entry
	for _, entry := range incrementals {
		if entry.Timestamp.Before(lastFull.Timestamp) {
			markedForDeletion = append(markedForDeletion, entry)
		} else {
			incrementalsSinceLast = append(incrementalsSinceLast, entry)
		}
	}

	// Ensure incrementals are in order, and not corrupted.
	sort.SliceStable(incrementals, func(i, j int) bool {
		return incrementals[i].ID < incrementals[j].ID
	})
	var i int
	for _, entry := range incrementalsSinceLast {
		if i == 0 {
			i = entry.ID
			continue
		}
		if entry.ID != i+1 || entry.Parent != i {
			return nil, fmt.Errorf(
				"something's gone wrong. Incremental backups are corrupted (missing a step between %d and %d)- returning error to be safe. Please fix manually (probably by manually deleting all incremental backups. Sorry friend.)",
				i, entry.ID)
		}
		i++
	}

	for _, set := range []struct {
		name    string
		entries []backup.Entry
		max     int
	}{
		{"fullLast45", fullLast45, int(db.Cadence.FullLast45Days)},
		{"full45to182", full45to182, int(db.Cadence.Full45To182Days)},
		{"full182to365", full182to365, int(db.Cadence.Full182To365Days)},
	} {
		for len(set.entries) > set.max {
			n := len(set.entries) / 2
			f.log.Info("deleting full backup from "+set.name,
				"id", set.entries[n].ID,
				"timestamp", set.entries[n].Timestamp,
				"backups", len(set.entries),
				"max", set.max,
			)
			markedForDeletion = append(markedForDeletion, set.entries[n])
			set.entries = append(set.entries[:n], set.entries[n+1:]...)
		}
	}

	for year, entries := range full365Plus {
		fmt.Printf("%d %v\n", year, entries)
		for len(entries) > int(db.Cadence.FullPer365Over365Days) {
			n := len(entries) / 2
			f.log.Info("deleting full backup from full365Plus",
				"id", full365Plus[year][n].ID,
				"timestamp", full365Plus[year][n].Timestamp,
				"backups", len(full365Plus[year]),
				"max", int(db.Cadence.FullPer365Over365Days),
				"year", now.Year()-year,
			)
			markedForDeletion = append(markedForDeletion, entries[n])
			entries = append(entries[:n], entries[n+1:]...)
		}
	}

	sort.SliceStable(markedForDeletion, func(i, j int) bool {
		return markedForDeletion[i].ID < markedForDeletion[j].ID
	})
	return markedForDeletion, nil
}
