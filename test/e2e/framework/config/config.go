package config

import (
	"errors"
	"flag"
	"strings"
)

type Config struct {
	Endpoint1   string
	Endpoint2   string
	Filesystem1 string
	Filesystem2 string
	BucketName1 string
	BucketName2 string
}

var (
	sharedConfig = &Config{}
)

func SetConfig(config *Config) {
	sharedConfig = config
}

func GetConfig() *Config {
	return sharedConfig
}

func (c *Config) AddFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.Endpoint1, "endpoint-1", "", "endpoint1")
	fs.StringVar(&c.Endpoint2, "endpoint-2", "", "endpoint2")
	fs.StringVar(&c.Filesystem1, "filesystem-1", "", "filesystem1")
	fs.StringVar(&c.Filesystem2, "filesystem-2", "", "filesystem2")
	fs.StringVar(&c.BucketName1, "bucketname-1", "", "bucketname-1")
	fs.StringVar(&c.BucketName2, "bucketname-2", "", "bucketname-2")
}

func (c *Config) Validate() error {
	var errs []string

	if len(c.Endpoint1) == 0 {
		errs = append(errs, "endpoint-1 is required")
	}
	if len(c.Endpoint2) == 0 {
		errs = append(errs, "endpoint-2 is required")
	}
	if len(c.Filesystem1) == 0 {
		errs = append(errs, "filesystem-1 is required")
	}
	if len(c.Filesystem2) == 0 {
		errs = append(errs, "filesystem-2 is required")
	}
	if len(c.BucketName1) == 0 {
		errs = append(errs, "bucketname1 is required")
	}
	if len(c.BucketName2) == 0 {
		errs = append(errs, "bucketname2 is required")
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n"))
	}

	return nil
}
