package client

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/go-logr/stdr"
	"github.com/stretchr/testify/assert"
	clocktesting "k8s.io/utils/clock/testing"

	"github.com/joshvanl/yazbu/internal/backup"
)

func Test_markedForDeletion(t *testing.T) {
	epoch := time.Date(2020, 5, 1, 0, 0, 0, 0, time.UTC)
	fakeclock := clocktesting.NewFakeClock(epoch)

	tests := map[string]struct {
		db     backup.DB
		exp    []backup.Entry
		expErr bool
	}{
		"no entries should return no marked entries": {
			db: backup.DB{
				Cadence: backup.Cadence{
					IncrementalPerLastFull: 1,
				},
				Entries: []backup.Entry{},
			},
			exp:    nil,
			expErr: false,
		},
		"if incremental backups with a full backup after all incremental timestamps, should return all incremental backups": {
			db: backup.DB{
				Cadence: backup.Cadence{
					IncrementalPerLastFull: 2,
					FullLast45Days:         2,
				},
				Entries: []backup.Entry{
					backup.Entry{ID: 1, Parent: 0, Type: backup.TypeFull, Timestamp: epoch.Add(-4)},
					backup.Entry{ID: 2, Parent: 1, Type: backup.TypeIncremental, Timestamp: epoch.Add(-3)},
					backup.Entry{ID: 3, Parent: 2, Type: backup.TypeIncremental, Timestamp: epoch.Add(-2)},
					backup.Entry{ID: 4, Parent: 1, Type: backup.TypeFull, Timestamp: epoch.Add(-1)},
				},
			},
			exp: []backup.Entry{
				backup.Entry{ID: 2, Parent: 1, Type: backup.TypeIncremental, Timestamp: epoch.Add(-3)},
				backup.Entry{ID: 3, Parent: 2, Type: backup.TypeIncremental, Timestamp: epoch.Add(-2)},
			},
			expErr: false,
		},
		"if more incremental backups than incrementalPerLastFull do nothing": {
			db: backup.DB{
				Cadence: backup.Cadence{
					IncrementalPerLastFull: 1,
				},
				Entries: []backup.Entry{
					backup.Entry{ID: 1, Parent: 0, Type: backup.TypeIncremental},
					backup.Entry{ID: 2, Parent: 1, Type: backup.TypeIncremental},
					backup.Entry{ID: 3, Parent: 2, Type: backup.TypeIncremental},
				},
			},
			exp:    nil,
			expErr: false,
		},
		"if incremental backups have the wrong parent, should return error": {
			db: backup.DB{
				Cadence: backup.Cadence{
					IncrementalPerLastFull: 1,
				},
				Entries: []backup.Entry{
					backup.Entry{ID: 1, Parent: 0, Type: backup.TypeIncremental},
					backup.Entry{ID: 2, Parent: 0, Type: backup.TypeIncremental},
					backup.Entry{ID: 3, Parent: 2, Type: backup.TypeIncremental},
				},
			},
			exp:    nil,
			expErr: true,
		},
		"if incremental backups have the ID increment, expect error": {
			db: backup.DB{
				Cadence: backup.Cadence{
					IncrementalPerLastFull: 1,
				},
				Entries: []backup.Entry{
					backup.Entry{ID: 1, Parent: 0, Type: backup.TypeIncremental},
					backup.Entry{ID: 2, Parent: 1, Type: backup.TypeIncremental},
					backup.Entry{ID: 4, Parent: 2, Type: backup.TypeIncremental},
				},
			},
			exp:    nil,
			expErr: true,
		},
		"if has less than the max full last 45 days, return nothing": {
			db: backup.DB{
				Cadence: backup.Cadence{
					IncrementalPerLastFull: 1,
					FullLast45Days:         2,
				},
				Entries: []backup.Entry{
					backup.Entry{ID: 4, Parent: 3, Type: backup.TypeFull, Timestamp: epoch.Add(-2)},
					backup.Entry{ID: 5, Parent: 4, Type: backup.TypeIncremental, Timestamp: epoch.Add(-1)},
				},
			},
			exp:    nil,
			expErr: false,
		},
		"if same than the max full last 45 days, return nothing": {
			db: backup.DB{
				Cadence: backup.Cadence{
					IncrementalPerLastFull: 1,
					FullLast45Days:         2,
				},
				Entries: []backup.Entry{
					backup.Entry{ID: 4, Parent: 3, Type: backup.TypeFull, Timestamp: epoch.Add(-3)},
					backup.Entry{ID: 5, Parent: 4, Type: backup.TypeFull, Timestamp: epoch.Add(-2)},
					backup.Entry{ID: 6, Parent: 5, Type: backup.TypeIncremental, Timestamp: epoch.Add(-1)},
				},
			},
			exp:    nil,
			expErr: false,
		},
		"if more than the max full last 45 days, return middle entry": {
			db: backup.DB{
				Cadence: backup.Cadence{
					IncrementalPerLastFull: 1,
					FullLast45Days:         2,
				},
				Entries: []backup.Entry{
					backup.Entry{ID: 4, Parent: 3, Type: backup.TypeFull, Timestamp: epoch.Add(-3)},
					backup.Entry{ID: 5, Parent: 4, Type: backup.TypeFull, Timestamp: epoch.Add(-2)},
					backup.Entry{ID: 6, Parent: 5, Type: backup.TypeFull, Timestamp: epoch.Add(-1)},
					backup.Entry{ID: 7, Parent: 6, Type: backup.TypeIncremental, Timestamp: epoch.Add(-1)},
				},
			},
			exp:    []backup.Entry{backup.Entry{ID: 5, Parent: 4, Type: backup.TypeFull, Timestamp: epoch.Add(-2)}},
			expErr: false,
		},
		"if many more than the max full last 45 days, return middle entries": {
			db: backup.DB{
				Cadence: backup.Cadence{
					IncrementalPerLastFull: 1,
					FullLast45Days:         2,
				},
				Entries: []backup.Entry{
					backup.Entry{ID: 4, Parent: 3, Type: backup.TypeFull, Timestamp: epoch.Add(-3)},
					backup.Entry{ID: 5, Parent: 4, Type: backup.TypeFull, Timestamp: epoch.Add(-2)},
					backup.Entry{ID: 6, Parent: 5, Type: backup.TypeFull, Timestamp: epoch.Add(-1)},
					backup.Entry{ID: 7, Parent: 6, Type: backup.TypeFull, Timestamp: epoch.Add(-1)},
					backup.Entry{ID: 8, Parent: 7, Type: backup.TypeIncremental, Timestamp: epoch.Add(-1)},
				},
			},
			exp: []backup.Entry{
				backup.Entry{ID: 5, Parent: 4, Type: backup.TypeFull, Timestamp: epoch.Add(-2)},
				backup.Entry{ID: 6, Parent: 5, Type: backup.TypeFull, Timestamp: epoch.Add(-1)},
			},
			expErr: false,
		},
		"if has less than the max full 45 to 182 days, return nothing": {
			db: backup.DB{
				Cadence: backup.Cadence{
					IncrementalPerLastFull: 1,
					FullLast45Days:         2,
					Full45To182Days:        3,
				},
				Entries: []backup.Entry{
					backup.Entry{ID: 3, Parent: 2, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 46)},
					backup.Entry{ID: 4, Parent: 3, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 45)},
					backup.Entry{ID: 5, Parent: 4, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 44)},
					backup.Entry{ID: 6, Parent: 5, Type: backup.TypeIncremental, Timestamp: epoch.Add(-1)},
				},
			},
			exp:    nil,
			expErr: false,
		},
		"if has same as the max full 45 to 182 days, return nothing": {
			db: backup.DB{
				Cadence: backup.Cadence{
					IncrementalPerLastFull: 1,
					FullLast45Days:         2,
					Full45To182Days:        3,
				},
				Entries: []backup.Entry{
					backup.Entry{ID: 2, Parent: 1, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 47)},
					backup.Entry{ID: 3, Parent: 2, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 46)},
					backup.Entry{ID: 4, Parent: 3, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 45)},
					backup.Entry{ID: 5, Parent: 4, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 44)},
					backup.Entry{ID: 6, Parent: 5, Type: backup.TypeIncremental, Timestamp: epoch.Add(-1)},
				},
			},
			exp:    nil,
			expErr: false,
		},
		"if has more than the max full 45 to 182 days, return middle entry": {
			db: backup.DB{
				Cadence: backup.Cadence{
					IncrementalPerLastFull: 1,
					FullLast45Days:         2,
					Full45To182Days:        3,
				},
				Entries: []backup.Entry{
					backup.Entry{ID: 1, Parent: 0, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 48)},
					backup.Entry{ID: 2, Parent: 1, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 47)},
					backup.Entry{ID: 3, Parent: 2, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 46)},
					backup.Entry{ID: 4, Parent: 3, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 45)},
					backup.Entry{ID: 5, Parent: 4, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 44)},
					backup.Entry{ID: 6, Parent: 5, Type: backup.TypeIncremental, Timestamp: epoch.Add(-1)},
				},
			},
			exp:    []backup.Entry{backup.Entry{ID: 3, Parent: 2, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 46)}},
			expErr: false,
		},
		"if has many more than the max full 45 to 182 days, return middle entries": {
			db: backup.DB{
				Cadence: backup.Cadence{
					IncrementalPerLastFull: 1,
					FullLast45Days:         2,
					Full45To182Days:        3,
				},
				Entries: []backup.Entry{
					backup.Entry{ID: 1, Parent: 0, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 50)},
					backup.Entry{ID: 2, Parent: 1, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 49)},
					backup.Entry{ID: 3, Parent: 2, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 48)},
					backup.Entry{ID: 4, Parent: 3, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 47)},
					backup.Entry{ID: 5, Parent: 4, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 46)},
					backup.Entry{ID: 6, Parent: 5, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 45)},
					backup.Entry{ID: 7, Parent: 6, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 44)},
					backup.Entry{ID: 8, Parent: 7, Type: backup.TypeIncremental, Timestamp: epoch.Add(-1)},
				},
			},
			exp: []backup.Entry{
				backup.Entry{ID: 3, Parent: 2, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 48)},
				backup.Entry{ID: 4, Parent: 3, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 47)},
				backup.Entry{ID: 5, Parent: 4, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 46)},
			},
			expErr: false,
		},
		"if has less than the max full 182 to 365 days, return nothing": {
			db: backup.DB{
				Cadence: backup.Cadence{
					IncrementalPerLastFull: 1,
					FullLast45Days:         2,
					Full45To182Days:        3,
					Full182To365Days:       4,
				},
				Entries: []backup.Entry{
					backup.Entry{ID: 2, Parent: 1, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 184)},
					backup.Entry{ID: 3, Parent: 2, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 183)},
					backup.Entry{ID: 4, Parent: 3, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 182)},
					backup.Entry{ID: 5, Parent: 4, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 181)},
					backup.Entry{ID: 6, Parent: 5, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 45)},
					backup.Entry{ID: 7, Parent: 6, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 2)},
					backup.Entry{ID: 8, Parent: 7, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 1)},
					backup.Entry{ID: 9, Parent: 8, Type: backup.TypeIncremental, Timestamp: epoch.Add(-1)},
				},
			},
			exp:    nil,
			expErr: false,
		},
		"if has same as the max full 182 to 365 days, return nothing": {
			db: backup.DB{
				Cadence: backup.Cadence{
					IncrementalPerLastFull: 1,
					FullLast45Days:         2,
					Full45To182Days:        3,
					Full182To365Days:       4,
				},
				Entries: []backup.Entry{
					backup.Entry{ID: 1, Parent: 0, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 185)},
					backup.Entry{ID: 2, Parent: 1, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 184)},
					backup.Entry{ID: 3, Parent: 2, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 183)},
					backup.Entry{ID: 4, Parent: 3, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 182)},
					backup.Entry{ID: 5, Parent: 4, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 181)},
					backup.Entry{ID: 6, Parent: 5, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 45)},
					backup.Entry{ID: 7, Parent: 6, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 2)},
					backup.Entry{ID: 8, Parent: 7, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 1)},
					backup.Entry{ID: 9, Parent: 8, Type: backup.TypeIncremental, Timestamp: epoch.Add(-1)},
				},
			},
			exp:    nil,
			expErr: false,
		},
		"if has more than the max full 182 to 365 days, return middle": {
			db: backup.DB{
				Cadence: backup.Cadence{
					IncrementalPerLastFull: 1,
					FullLast45Days:         2,
					Full45To182Days:        3,
					Full182To365Days:       4,
				},
				Entries: []backup.Entry{
					backup.Entry{ID: 1, Parent: 0, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 186)},
					backup.Entry{ID: 2, Parent: 1, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 185)},
					backup.Entry{ID: 3, Parent: 2, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 184)},
					backup.Entry{ID: 4, Parent: 3, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 183)},
					backup.Entry{ID: 5, Parent: 4, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 182)},
					backup.Entry{ID: 6, Parent: 5, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 181)},
					backup.Entry{ID: 7, Parent: 6, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 45)},
					backup.Entry{ID: 8, Parent: 7, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 2)},
					backup.Entry{ID: 9, Parent: 8, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 1)},
					backup.Entry{ID: 10, Parent: 9, Type: backup.TypeIncremental, Timestamp: epoch.Add(-1)},
				},
			},
			exp:    []backup.Entry{backup.Entry{ID: 3, Parent: 2, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 184)}},
			expErr: false,
		},
		"if has many more than the max full 182 to 365 days, return middle entries": {
			db: backup.DB{
				Cadence: backup.Cadence{
					IncrementalPerLastFull: 1,
					FullLast45Days:         2,
					Full45To182Days:        3,
					Full182To365Days:       4,
				},
				Entries: []backup.Entry{
					backup.Entry{ID: 1, Parent: 0, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 188)},
					backup.Entry{ID: 2, Parent: 1, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 187)},
					backup.Entry{ID: 3, Parent: 2, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 186)},
					backup.Entry{ID: 4, Parent: 3, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 185)},
					backup.Entry{ID: 5, Parent: 4, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 184)},
					backup.Entry{ID: 6, Parent: 5, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 183)},
					backup.Entry{ID: 7, Parent: 6, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 182)},
					backup.Entry{ID: 8, Parent: 7, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 181)},
					backup.Entry{ID: 9, Parent: 8, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 45)},
					backup.Entry{ID: 10, Parent: 9, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 2)},
					backup.Entry{ID: 11, Parent: 10, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 1)},
					backup.Entry{ID: 12, Parent: 11, Type: backup.TypeIncremental, Timestamp: epoch.Add(-1)},
				},
			},
			exp: []backup.Entry{
				backup.Entry{ID: 3, Parent: 2, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 186)},
				backup.Entry{ID: 4, Parent: 3, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 185)},
				backup.Entry{ID: 5, Parent: 4, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 184)},
			},
			expErr: false,
		},
		"if has less than the max full per year, return nil": {
			db: backup.DB{
				Cadence: backup.Cadence{
					IncrementalPerLastFull: 1,
					FullLast45Days:         2,
					Full45To182Days:        3,
					Full182To365Days:       4,
					FullPer365Over365Days:  3,
				},
				Entries: []backup.Entry{
					backup.Entry{ID: 1, Parent: 0, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 366 * 2)},
					backup.Entry{ID: 2, Parent: 0, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 365 * 2)},
					backup.Entry{ID: 3, Parent: 0, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 366)},
					backup.Entry{ID: 4, Parent: 0, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 365)},
					backup.Entry{ID: 5, Parent: 0, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 364)},
					backup.Entry{ID: 6, Parent: 0, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 363)},
					backup.Entry{ID: 7, Parent: 0, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 362)},
				},
			},
			exp:    nil,
			expErr: false,
		},
		"if has same as the max full per year, return nothing": {
			db: backup.DB{
				Cadence: backup.Cadence{
					IncrementalPerLastFull: 1,
					FullLast45Days:         2,
					Full45To182Days:        3,
					Full182To365Days:       4,
					FullPer365Over365Days:  3,
				},
				Entries: []backup.Entry{
					backup.Entry{ID: 1, Parent: 0, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 368 * 3)},
					backup.Entry{ID: 2, Parent: 0, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 367 * 3)},
					backup.Entry{ID: 3, Parent: 0, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 366 * 3)},
					backup.Entry{ID: 4, Parent: 0, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 367 * 2)},
					backup.Entry{ID: 5, Parent: 0, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 366 * 2)},
					backup.Entry{ID: 6, Parent: 0, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 365 * 2)},
					backup.Entry{ID: 7, Parent: 0, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 367)},
					backup.Entry{ID: 8, Parent: 0, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 366)},
					backup.Entry{ID: 9, Parent: 0, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 365)},
					backup.Entry{ID: 10, Parent: 0, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 364)},
					backup.Entry{ID: 11, Parent: 0, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 363)},
					backup.Entry{ID: 12, Parent: 0, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 362)},
				},
			},
			exp:    nil,
			expErr: false,
		},
		"if has more than the max full per year, return the middle entries": {
			db: backup.DB{
				Cadence: backup.Cadence{
					IncrementalPerLastFull: 1,
					FullLast45Days:         2,
					Full45To182Days:        3,
					Full182To365Days:       4,
					FullPer365Over365Days:  3,
				},
				Entries: []backup.Entry{
					backup.Entry{ID: 1, Parent: 0, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 369 * 3)},
					backup.Entry{ID: 2, Parent: 0, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 368 * 3)},
					backup.Entry{ID: 3, Parent: 0, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 367 * 3)},
					backup.Entry{ID: 4, Parent: 0, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 366 * 3)},
					backup.Entry{ID: 5, Parent: 0, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 365 * 3)},
					backup.Entry{ID: 6, Parent: 0, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 368 * 2)},
					backup.Entry{ID: 7, Parent: 0, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 367 * 2)},
					backup.Entry{ID: 8, Parent: 0, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 366 * 2)},
					backup.Entry{ID: 9, Parent: 0, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 365 * 2)},
					backup.Entry{ID: 10, Parent: 0, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 369)},
					backup.Entry{ID: 11, Parent: 0, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 368)},
					backup.Entry{ID: 12, Parent: 0, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 367)},
					backup.Entry{ID: 13, Parent: 0, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 366)},
					backup.Entry{ID: 14, Parent: 0, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 365)},
					backup.Entry{ID: 15, Parent: 0, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 364)},
					backup.Entry{ID: 16, Parent: 0, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 363)},
					backup.Entry{ID: 17, Parent: 0, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 362)},
				},
			},
			exp: []backup.Entry{
				backup.Entry{ID: 3, Parent: 0, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 367 * 3)},
				backup.Entry{ID: 4, Parent: 0, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 366 * 3)},
				backup.Entry{ID: 8, Parent: 0, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 366 * 2)},
				backup.Entry{ID: 12, Parent: 0, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 367)},
				backup.Entry{ID: 13, Parent: 0, Type: backup.TypeFull, Timestamp: epoch.Add(-time.Hour * 24 * 366)},
			},
			expErr: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			entries, err := (&fsclient{
				clock: fakeclock,
				log:   stdr.New(log.New(os.Stdout, "", log.Lshortfile)),
			}).markedForDeletion(test.db)
			assert.Equal(t, test.expErr, err != nil, "%v", err)
			assert.Equal(t, test.exp, entries)
		})
	}
}
