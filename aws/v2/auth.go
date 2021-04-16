package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/gruntwork-io/go-commons/errors"
)

const (
	DefaultRegion = "us-east-1"
)

// Options represents all the parameters necessary for setting up authentication credentials with AWS.
type Options struct {
	Region string

	Context context.Context
}

// NewOptions will create a new aws.Options struct that provides reasonable defaults for unspecified values.
func NewOptions(region string) *Options {
	return &Options{
		Region:  region,
		Context: context.TODO(),
	}
}

// NewDefaultConfig will retrieve a new authenticated AWS config using SDK default credentials. This config can be used
// to setup new AWS service clients.
func NewDefaultConfig(opts *Options) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(opts.Context, config.WithRegion(opts.Region))
	return cfg, errors.WithStackTrace(err)
}
