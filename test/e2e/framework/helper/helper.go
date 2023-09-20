package helper

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"gopkg.in/yaml.v3"

	"github.com/joshvanl/yazbu/config"
	testconfig "github.com/joshvanl/yazbu/e2e/framework/config"
)

type Helper struct {
	config *testconfig.Config
}

func New(config *testconfig.Config) *Helper {
	return &Helper{config}
}

func (h *Helper) DeleteBucket(ctx context.Context, endpoint, bucket string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/%s", endpoint, bucket), nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	var list s3.ListObjectsOutput
	if err := xml.NewDecoder(resp.Body).Decode(&list); err != nil {
		return err
	}

	for _, object := range list.Contents {
		req, err := http.NewRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf("%s/%s/%s", endpoint, bucket, *object.Key), nil)
		if err != nil {
			return err
		}
		if _, err := http.DefaultClient.Do(req); err != nil {
			return err
		}
	}

	url, err := url.Parse(endpoint)
	if err != nil {
		return err
	}

	url.Path = bucket
	req, err = http.NewRequestWithContext(ctx, http.MethodDelete, url.String(), strings.NewReader(""))
	if err != nil {
		return err
	}
	_, err = http.DefaultClient.Do(req)
	return err
}

func (h *Helper) CreateBucket(ctx context.Context, endpoint, bucket string) error {
	url, err := url.Parse(endpoint)
	if err != nil {
		return err
	}
	url.Path = bucket
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url.String(), strings.NewReader(""))
	if err != nil {
		return err
	}
	if _, err := http.DefaultClient.Do(req); err != nil {
		return err
	}
	return nil
}

func (h *Helper) S3(endpoint string) (*s3.S3, error) {
	url, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	sess, err := session.NewSession(&aws.Config{
		Endpoint: aws.String(url.Host), Region: aws.String("region"),
		DisableSSL:  aws.Bool(true),
		Credentials: credentials.NewStaticCredentials("remote-identity", "remote-credential", ""),
	})
	if err != nil {
		return nil, err
	}
	return s3.New(sess), nil
}

func (h *Helper) YazbuList(ctx context.Context, config *config.Config) (string, error) {
	cfdYaml, err := yaml.Marshal(config)
	if err != nil {
		return "", err
	}

	f, err := os.CreateTemp(os.TempDir(), "yazbu-config-list-")
	if err != nil {
		return "", err
	}

	if _, err := io.Copy(f, bytes.NewReader(cfdYaml)); err != nil {
		return "", err
	}

	cmd := exec.Command(h.config.YazbuBin, "list", "--config", f.Name())
	var buf bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &buf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &buf)

	if err := cmd.Run(); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (h *Helper) YazbuDefaultConfig() *config.Config {
	return (&config.Config{
		Buckets: []config.Bucket{
			config.Bucket{
				Name:         h.config.BucketName1,
				Endpoint:     h.config.Endpoint1,
				Region:       "region",
				AccessKey:    h.config.AccessKey1,
				SecretKey:    h.config.SecretKey1,
				StorageClass: "STANDARD",
			},
			config.Bucket{
				Name:         h.config.BucketName2,
				Endpoint:     h.config.Endpoint1,
				Region:       "region",
				AccessKey:    h.config.AccessKey2,
				SecretKey:    h.config.SecretKey2,
				StorageClass: "STANDARD",
			},
			config.Bucket{
				Name:         h.config.BucketName1,
				Endpoint:     h.config.Endpoint2,
				Region:       "region",
				StorageClass: "STANDARD",
				AccessKey:    h.config.AccessKey1,
				SecretKey:    h.config.SecretKey1,
			},
			config.Bucket{
				Name:         h.config.BucketName2,
				Endpoint:     h.config.Endpoint2,
				Region:       "region",
				StorageClass: "STANDARD",
				AccessKey:    h.config.AccessKey2,
				SecretKey:    h.config.SecretKey2,
			},
		},
		Filesystems: []string{
			h.config.Filesystem1,
			h.config.Filesystem2,
		},
		Cadence: config.Cadence{},
	}).DefaultValues()
}
