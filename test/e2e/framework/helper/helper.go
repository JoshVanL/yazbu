package helper

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/joshvanl/yazbu/test/e2e/framework/config"
)

type Helper struct {
	config *config.Config
}

func New(config *config.Config) *Helper {
	return &Helper{config}
}

func (h *Helper) DeleteBucket(ctx context.Context, endpoint, bucket string) error {
	url, err := url.Parse(endpoint)
	if err != nil {
		return err
	}
	url.Path = bucket
	req, err := http.NewRequest(http.MethodDelete, url.String(), strings.NewReader(""))
	if err != nil {
		return err
	}
	if _, err := http.DefaultClient.Do(req); err != nil {
		return err
	}
	return nil
}

func (h *Helper) CreateBucket(ctx context.Context, endpoint, bucket string) error {
	url, err := url.Parse(endpoint)
	if err != nil {
		return err
	}
	url.Path = bucket
	req, err := http.NewRequest(http.MethodPut, url.String(), strings.NewReader(""))
	if err != nil {
		return err
	}
	if _, err := http.DefaultClient.Do(req); err != nil {
		return err
	}
	return nil
}

func (h *Helper) S3(endpoint string) *s3.S3 {
	return s3.New(session.New(&aws.Config{
		Endpoint: aws.String(endpoint), Region: aws.String("region"),
		Credentials: credentials.NewStaticCredentials("remote-identity", "remote-credential", ""),
	}))
}
