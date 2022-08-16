package manager

import (
	"context"
	"time"

	"github.com/joshvanl/yazbu/internal/backup"
)

// executeCadence will delete all backup entries which need to be deleted,
// according to the cadence.
func (m *Manager) executeCadence(ctx context.Context, db *backup.DB) error {
	m.log.Info("checking database backups to delete stale backups based on configured cadence")

	now := time.Now()

	var (
		incrementalLast []backup.Entry
		lastFull        backup.Entry
		fullLast45      []backup.Entry
		full45to182     []backup.Entry
		full182to365    []backup.Entry
		full365Plus     = make(map[int][]backup.Entry)
	)

	for _, entry := range db.Entries {
		switch {
		case entry.Type == backup.TypeIncremental:
			incrementalLast = append(incrementalLast, entry)
			continue

		case entry.Timestamp.After(now.Add(-time.Hour * 24 * 45)):
			fullLast45 = append(fullLast45, entry)

		case entry.Timestamp.After(now.Add(-time.Hour * 24 * 182)):
			full45to182 = append(full45to182, entry)

		case entry.Timestamp.After(now.Add(-time.Hour * 24 * 365)):
			full182to365 = append(full182to365, entry)

		default:
			diff := entry.Timestamp.Sub(now)
			year := int(diff % time.Hour * 24 * 365)
			full365Plus[year] = append(full365Plus[year], entry)
		}

		if entry.Type == backup.TypeFull && (lastFull.Timestamp.IsZero() || entry.Timestamp.After(lastFull.Timestamp)) {
			lastFull = entry
		}
	}

	return nil
}
