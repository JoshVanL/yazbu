package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config is the top level config to configure backups.
type Config struct {
	// Buckets is the configuration for the target S3 compatible buckets.
	Buckets []Bucket `yaml:"buckets"`

	// Filesystems is the set of ZFS dataset filesystems to backup.
	Filesystems []string `yaml:"filesystems"`

	// Cadence describes the number of backups to keep for each filesystem.
	// Generally backups begin to decay over time, resulting in less frequency of
	// backups the further in the past from the current time.
	Cadence Cadence `yaml:"cadence"`
}

// Bucket if the location and authentication configuration to write and read
// backups from.
type Bucket struct {
	// Name is the name of the S3 bucket to store backups.
	Name string `yaml:"name"`

	// Region is the region where the S3 bucket is located.
	// example:
	// "auto"
	Region string `yaml:"region"`

	// Endpoint is S3 compatible URL where backups are written and read from. For
	// example:
	// https://storage.googleapis.com
	// https://s3.eu-central-003.backblazeb2.com
	Endpoint string `yaml:"endpoint"`

	// StorageClass is the class of storage to write backup files with.
	StorageClass string `yaml:"storageClass"`

	// AccessKey is the access key to authenticate to the S3 endpoint.
	AccessKey string `yaml:"accessKey"`

	// SecretKey is the secret key to authenticate to the S3 endpoint.
	SecretKey string `yaml:"secretKey"`
}

// Cadence describes the cadence of backups, and how older backups are deleted
// as they decay over time. It is recommended that the user setup the decay
// rate at each window to decrease the number of backups over time.
// When a window contains more backups than the maximum target for that window,
// the middle most backup will be deleted.
type Cadence struct {
	// IncrementalPerLastFull is the number of incremental backups to store
	// between each full backup. All incremental backups are deleted once a full
	// backup is taken.
	// Default 7.
	IncrementalPerLastFull *uint `yaml:"incrementalPerLastFull"`

	// FullLast45Days describes the number of full backups to store for the last
	// 45 days from the current time.
	// (45 days).
	// Default 10 (4.5 day/backup).
	FullLast45Days *uint `yaml:"fullLast45Days"`

	// FullLast45To182Days describes the number of full backups to store for the
	// 45->182 days from the current time in the past.
	// (137 days).
	// Default 10 (13.7 day/backup).
	Full45To182Days *uint `yaml:"full45To182Days"`

	// Full182To365Days describes the number of full backups to store between
	// 182->365 days from the current time in the past.
	// (183 days).
	// Default 5 (36.6 day/backup).
	Full182To365Days *uint `yaml:"full182To365Days"`

	// FullPer365Over365Days describes the number of full backups to store after
	// 365 days from the current time in the past. Every 365 day window is
	// treated as a separate bucket of backups. The last 365 days from the
	// current time in the past are not included (i.e. all previous windows
	// defined above).
	// (365 days).
	// Default 4 (91.25 day/backup).
	FullPer365Over365Days *uint `yaml:"fullPer365Over365Days"`
}

// ReadFile reads the given config path location, and returns the parsed
// config.
func ReadFile(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file %q: %w", path, err)
	}

	var config Config
	if err := yaml.NewDecoder(f).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config file %q: %w", path, err)
	}

	if err := config.validate(); err != nil {
		return nil, err
	}

	return &config, nil
}

// DefaultValues sets the default values of the config, if required values are
// not set.
func (c *Config) DefaultValues() *Config {
	defaultIfNil(&c.Cadence.IncrementalPerLastFull, 7)
	defaultIfNil(&c.Cadence.FullLast45Days, 10)
	defaultIfNil(&c.Cadence.Full45To182Days, 10)
	defaultIfNil(&c.Cadence.Full182To365Days, 5)
	defaultIfNil(&c.Cadence.FullPer365Over365Days, 4)
	return c
}

// defaultIfNil sets the default of the given pointer, if the value is nil.
func defaultIfNil(p **uint, def uint) {
	if *p == nil {
		*p = &def
	}
}

// validate validates the given configuration is correct and usable.
func (c *Config) validate() error {
	var errs []string
	if len(c.Buckets) == 0 {
		errs = append(errs, "must specify at least one bucket")
	}

	if len(c.Filesystems) == 0 {
		errs = append(errs, "must specify at least one filesystem")
	}

	bucketEndpoints := make(map[string]struct{})
	for i, bucket := range c.Buckets {
		if len(bucket.Name) == 0 {
			errs = append(errs, fmt.Sprintf("%d: bucket name must be defined", i))
		}

		if len(bucket.Endpoint) == 0 {
			errs = append(errs, fmt.Sprintf("%d: bucket endpoint must be defined", i))
		}

		bucketEndpoint := path.Join(bucket.Endpoint, bucket.Name)
		if _, ok := bucketEndpoints[bucketEndpoint]; ok {
			errs = append(errs, fmt.Sprintf("%d: bucket endpoint can only be configured at most once: %q", i, bucketEndpoint))
		}
		bucketEndpoints[bucketEndpoint] = struct{}{}

		if len(bucket.StorageClass) == 0 {
			errs = append(errs, fmt.Sprintf("%d: bucket storageClass must be defined", i))
		}

		if len(bucket.Region) == 0 {
			errs = append(errs, fmt.Sprintf("%d: bucket region must be defined", i))
		}
	}

	mustNotNil := func(name string, p *uint) {
		if p == nil {
			errs = append(errs, fmt.Sprintf("%s must be set", name))
		}
	}
	mustNotNil("cadence.incrementalPerLastFull", c.Cadence.IncrementalPerLastFull)
	mustNotNil("cadence.fullLast45Days", c.Cadence.FullLast45Days)
	mustNotNil("cadence.full45To182Days", c.Cadence.Full45To182Days)
	mustNotNil("cadence.full182To365Days", c.Cadence.Full182To365Days)
	mustNotNil("cadence.fullPer365Over365Days", c.Cadence.FullPer365Over365Days)

	if c.Cadence.FullLast45Days != nil && *c.Cadence.FullLast45Days < 1 {
		errs = append(errs, "cadence.fullLast45Days must be at least 1 or higher")
	}

	if len(errs) > 0 {
		return fmt.Errorf("config: [%s]", strings.Join(errs, ", "))
	}

	return nil
}

// ToJSON returns the Cadence as a JSON string.
func (c Cadence) ToJSON() string {
	out, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}

	return string(out)
}
