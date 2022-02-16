package awscommons

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/gruntwork-io/go-commons/errors"
)

// UploadObjectString will upload the provided string to the given S3 bucket as an object under the specified key.
func UploadObjectString(opts *Options, bucket, key, contents string) error {
	client, err := NewS3Client(opts)
	if err != nil {
		return err
	}

	input := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   strings.NewReader(contents),
	}
	_, err = client.PutObject(opts.Context, input)
	return errors.WithStackTrace(err)
}

// NewS3Client will return a new AWS SDK client for interacting with AWS S3.
func NewS3Client(opts *Options) (*s3.Client, error) {
	cfg, err := NewDefaultConfig(opts)
	if err != nil {
		return nil, errors.WithStackTrace(err)
	}
	return s3.NewFromConfig(cfg), nil
}
