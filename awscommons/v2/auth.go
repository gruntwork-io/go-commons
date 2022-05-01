package awscommons

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
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

// NewAssumedRoleConfig will retrieve a new authenticated AWS config that uses the SDK default credentials to assume the
// given role. This config can then be used to setup new AWS service clients.
func NewAssumedRoleConfig(opts *Options, roleARN, sessionName string) (aws.Config, error) {
	cfg, err := NewDefaultConfig(opts)
	if err != nil {
		return cfg, err
	}

	// Use sts:AssumeRole to get temporary credentials, and update the assigned credentials on the config object.
	client := sts.NewFromConfig(cfg)
	provider := stscreds.NewAssumeRoleProvider(client, roleARN)
	cfg.Credentials = aws.NewCredentialsCache(provider)
	return cfg, nil
}

// GetAuthEnvVars will return environment variables that can be set to authenticate to AWS using the provided
// authenticated SDK config.
func GetAuthEnvVars(opts *Options, config aws.Config) (map[string]string, error) {
	creds, err := config.Credentials.Retrieve(opts.Context)
	if err != nil {
		return nil, err
	}

	authEnv := map[string]string{
		"AWS_ACCESS_KEY_ID":     creds.AccessKeyID,
		"AWS_SECRET_ACCESS_KEY": creds.SecretAccessKey,
		"AWS_SESSION_TOKEN":     creds.SessionToken,
		"AWS_SECURITY_TOKEN":    creds.SessionToken,
	}
	if creds.CanExpire {
		authEnv["AWS_SESSION_EXPIRATION"] = creds.Expires.Format(time.RFC3339)
	}
	return authEnv, nil
}
