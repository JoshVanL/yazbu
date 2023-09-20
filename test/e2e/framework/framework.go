package framework

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/joshvanl/yazbu/e2e/framework/config"
	"github.com/joshvanl/yazbu/e2e/framework/helper"
)

type Framework struct {
	BaseName string

	config *config.Config
	helper *helper.Helper
}

func NewDefaultFramework(baseName string) *Framework {
	return NewFramework(baseName, config.GetConfig())
}

func NewFramework(baseName string, config *config.Config) *Framework {
	f := &Framework{
		BaseName: baseName,
		config:   config,
	}

	JustBeforeEach(f.BeforeEach)
	JustAfterEach(f.AfterEach)

	return f
}

func (f *Framework) BeforeEach() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	f.helper = helper.New(f.config)

	client1, err := f.helper.S3(config.GetConfig().Endpoint1)
	Expect(err).NotTo(HaveOccurred())
	client2, err := f.helper.S3(config.GetConfig().Endpoint2)
	Expect(err).NotTo(HaveOccurred())

	list, err := client1.ListBuckets(new(s3.ListBucketsInput))
	Expect(err).NotTo(HaveOccurred())
	Expect(list.Buckets).Should(BeEmpty())
	list, err = client2.ListBuckets(new(s3.ListBucketsInput))
	Expect(err).NotTo(HaveOccurred())
	Expect(list.Buckets).Should(BeEmpty())

	for _, pair := range []struct {
		endpoint string
		bucket   string
	}{
		{config.GetConfig().Endpoint1, config.GetConfig().BucketName1},
		{config.GetConfig().Endpoint1, config.GetConfig().BucketName2},
		{config.GetConfig().Endpoint2, config.GetConfig().BucketName1},
		{config.GetConfig().Endpoint2, config.GetConfig().BucketName2},
	} {
		By(fmt.Sprintf("Creating bucket %s on endpoint %s", pair.bucket, pair.endpoint))
		Expect(
			f.helper.CreateBucket(ctx, pair.endpoint, pair.bucket),
		).NotTo(HaveOccurred())
	}

	list, err = client1.ListBuckets(new(s3.ListBucketsInput))
	Expect(err).NotTo(HaveOccurred())
	Expect(list.Buckets).Should(HaveLen(2))
	list, err = client2.ListBuckets(new(s3.ListBucketsInput))
	Expect(err).NotTo(HaveOccurred())
	Expect(list.Buckets).Should(HaveLen(2))
}

func (f *Framework) AfterEach() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	for _, pair := range []struct {
		endpoint string
		bucket   string
	}{
		{endpoint: config.GetConfig().Endpoint1, bucket: config.GetConfig().BucketName1},
		{endpoint: config.GetConfig().Endpoint1, bucket: config.GetConfig().BucketName2},
		{endpoint: config.GetConfig().Endpoint2, bucket: config.GetConfig().BucketName1},
		{endpoint: config.GetConfig().Endpoint2, bucket: config.GetConfig().BucketName2},
	} {
		By(fmt.Sprintf("Deleting bucket %s on endpoint %s", pair.bucket, pair.endpoint))
		Expect(
			f.helper.DeleteBucket(ctx, pair.endpoint, pair.bucket),
		).NotTo(HaveOccurred())
	}

	client1, err := f.helper.S3(config.GetConfig().Endpoint1)
	Expect(err).NotTo(HaveOccurred())
	client2, err := f.helper.S3(config.GetConfig().Endpoint2)
	Expect(err).NotTo(HaveOccurred())

	list, err := client1.ListBuckets(new(s3.ListBucketsInput))
	Expect(err).NotTo(HaveOccurred())
	Expect(list.Buckets).Should(BeEmpty())
	list, err = client2.ListBuckets(new(s3.ListBucketsInput))
	Expect(err).NotTo(HaveOccurred())
	Expect(list.Buckets).Should(BeEmpty())
}

func (f *Framework) Config() *config.Config {
	return f.config
}

func (f *Framework) Helper() *helper.Helper {
	return f.helper
}

func CasesDescribe(text string, body func()) bool {
	return Describe("[yazbu] "+text, body)
}
