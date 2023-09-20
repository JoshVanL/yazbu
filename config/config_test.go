package config

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_DefaultValues(t *testing.T) {
	uintToPtr := func(u uint) *uint {
		return &u
	}

	tests := map[string]struct {
		config    Config
		expConfig Config
	}{
		"if no values set, expect default values to be set": {
			config: Config{
				Cadence: Cadence{},
			},
			expConfig: Config{
				Cadence: Cadence{
					IncrementalPerLastFull: uintToPtr(7),
					FullLast45Days:         uintToPtr(10),
					Full45To182Days:        uintToPtr(10),
					Full182To365Days:       uintToPtr(5),
					FullPer365Over365Days:  uintToPtr(4),
				},
			},
		},

		"if a mixture, expect those set to nil to be defaulted": {
			config: Config{
				Cadence: Cadence{
					IncrementalPerLastFull: uintToPtr(4),
					Full45To182Days:        uintToPtr(50),
					FullPer365Over365Days:  uintToPtr(0),
				},
			},
			expConfig: Config{
				Cadence: Cadence{
					IncrementalPerLastFull: uintToPtr(4),
					FullLast45Days:         uintToPtr(10),
					Full45To182Days:        uintToPtr(50),
					Full182To365Days:       uintToPtr(5),
					FullPer365Over365Days:  uintToPtr(0),
				},
			},
		},

		"if all values set, expect no default values to be set": {
			config: Config{
				Cadence: Cadence{
					IncrementalPerLastFull: uintToPtr(4),
					FullLast45Days:         uintToPtr(25),
					Full45To182Days:        uintToPtr(50),
					Full182To365Days:       uintToPtr(1),
					FullPer365Over365Days:  uintToPtr(0),
				},
			},
			expConfig: Config{
				Cadence: Cadence{
					IncrementalPerLastFull: uintToPtr(4),
					FullLast45Days:         uintToPtr(25),
					Full45To182Days:        uintToPtr(50),
					Full182To365Days:       uintToPtr(1),
					FullPer365Over365Days:  uintToPtr(0),
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			test.config.DefaultValues()
			assert.Equal(t, test.expConfig, test.config)
		})
	}
}

func Test_validate(t *testing.T) {
	var zero uint = 0
	var one uint = 1

	tests := map[string]struct {
		config Config
		expErr error
	}{
		"if everything is wrong, then expect error": {
			config: Config{
				Buckets:     []Bucket{},
				Filesystems: []string{},
				Cadence:     Cadence{},
			},
			expErr: errors.New("config: [must specify at least one bucket, must specify at least one filesystem, cadence.incrementalPerLastFull must be set, cadence.fullLast45Days must be set, cadence.full45To182Days must be set, cadence.full182To365Days must be set, cadence.fullPer365Over365Days must be set]"),
		},
		"if buckets have bad config": {
			config: Config{
				Buckets:     []Bucket{Bucket{}, Bucket{}},
				Filesystems: []string{"rpool/foo"},
				Cadence: Cadence{
					FullLast45Days:         &one,
					IncrementalPerLastFull: &zero,
					Full45To182Days:        &zero,
					Full182To365Days:       &zero,
					FullPer365Over365Days:  &zero,
				},
			},
			expErr: errors.New("config: [0: bucket name must be defined, 0: bucket endpoint must be defined, 0: bucket storageClass must be defined, 0: bucket region must be defined, 1: bucket name must be defined, 1: bucket endpoint must be defined, 1: bucket endpoint can only be configured at most once: \"\", 1: bucket storageClass must be defined, 1: bucket region must be defined]"),
		},
		"if duplicates of bucket endpoints, then error": {
			config: Config{
				Buckets: []Bucket{
					Bucket{Name: "foo", Endpoint: "foo", Region: "region", StorageClass: "standard"},
					Bucket{Name: "bar", Endpoint: "foo", Region: "region", StorageClass: "standard"},
					Bucket{Name: "bar", Endpoint: "foo", Region: "region", StorageClass: "standard"},
					Bucket{Name: "foo", Endpoint: "foo", Region: "region", StorageClass: "standard"},
				},
				Filesystems: []string{"rpool/foo"},
				Cadence: Cadence{
					FullLast45Days:         &one,
					IncrementalPerLastFull: &zero,
					Full45To182Days:        &zero,
					Full182To365Days:       &zero,
					FullPer365Over365Days:  &zero,
				},
			},
			expErr: errors.New("config: [2: bucket endpoint can only be configured at most once: \"foo/bar\", 3: bucket endpoint can only be configured at most once: \"foo/foo\"]"),
		},
		"if last 45 day cadence is 0, expect error": {
			config: Config{
				Buckets:     []Bucket{Bucket{Name: "foo"}},
				Filesystems: []string{"rpool/foo"},
				Cadence: Cadence{
					FullLast45Days:         &zero,
					IncrementalPerLastFull: &zero,
					Full45To182Days:        &zero,
					Full182To365Days:       &zero,
					FullPer365Over365Days:  &zero,
				},
			},
			expErr: errors.New("config: [0: bucket endpoint must be defined, 0: bucket storageClass must be defined, 0: bucket region must be defined, cadence.fullLast45Days must be at least 1 or higher]"),
		},
		"if validation is ok, expect no error": {
			config: Config{
				Buckets:     []Bucket{Bucket{Name: "foo", Endpoint: "foo", Region: "region", StorageClass: "standard"}},
				Filesystems: []string{"rpool/foo"},
				Cadence: Cadence{
					FullLast45Days:         &one,
					IncrementalPerLastFull: &zero,
					Full45To182Days:        &zero,
					Full182To365Days:       &zero,
					FullPer365Over365Days:  &zero,
				},
			},
			expErr: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := test.config.validate()
			require.Equal(t, test.expErr != nil, err != nil)
			if test.expErr != nil {
				assert.Equal(t, test.expErr.Error(), err.Error())
			}
		})
	}
}
