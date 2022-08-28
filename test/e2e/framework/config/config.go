package config

import (
	"errors"
	"flag"
	"strings"
)

type Config struct {
	YazbuBin string

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
	fs.StringVar(&c.YazbuBin, "yazbu-bin", "", "yazbu-bin")

	fs.StringVar(&c.Endpoint1, "endpoint-1", "", "endpoint1")
	fs.StringVar(&c.Endpoint2, "endpoint-2", "", "endpoint2")
	fs.StringVar(&c.Filesystem1, "filesystem-1", "", "filesystem1")
	fs.StringVar(&c.Filesystem2, "filesystem-2", "", "filesystem2")
	fs.StringVar(&c.BucketName1, "bucketname-1", "", "bucketname-1")
	fs.StringVar(&c.BucketName2, "bucketname-2", "", "bucketname-2")
}

func (c *Config) Validate() error {
	var errs []string

	for _, pair := range []struct {
		name  string
		value string
	}{
		{"yazbu-bin", c.YazbuBin},
		{"endpoint-1", c.Endpoint1},
		{"endpoint-2", c.Endpoint2},
		{"filesystem-1", c.Filesystem1},
		{"filesystem-2", c.Filesystem2},
		{"bucketname-1", c.BucketName1},
		{"bucketname-2", c.BucketName2},
	} {
		if len(pair.value) == 0 {
			errs = append(errs, pair.name+" is required")
		}
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n"))
	}

	return nil
}
