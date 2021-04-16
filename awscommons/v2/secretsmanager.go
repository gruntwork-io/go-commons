package awscommons

import (
	goerrors "errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"

	"github.com/gruntwork-io/go-commons/errors"
)

// GetSecretsManagerMetadata returns the metadata of the Secrets Manager entry with the given ID.
func GetSecretsManagerMetadata(opts *Options, secretID string) (*secretsmanager.DescribeSecretOutput, error) {
	client, err := NewSecretsManagerClient(opts)
	if err != nil {
		return nil, err
	}

	secret, err := client.DescribeSecret(opts.Context, &secretsmanager.DescribeSecretInput{SecretId: aws.String(secretID)})
	return secret, errors.WithStackTrace(err)
}

// SecretsManagerEntryExists returns whether or not the SecretsManager Entry with the given ARN exists. This will return
// an error if it exists, but is not accessible due to permission issues, or if there is an authentication issue.
func SecretsManagerEntryExists(opts *Options, arn string) (bool, error) {
	_, err := GetSecretsManagerMetadata(opts, arn)
	if err != nil {
		// If the err is 404, then return that it doesn't exist without error. Otherwise, return the error.
		rawErr := errors.Unwrap(err)
		var resourceNotFoundErr *types.ResourceNotFoundException
		if goerrors.As(rawErr, &resourceNotFoundErr) {
			return false, nil
		}
		return false, errors.WithStackTrace(err)
	}
	// At this point, we know the secret exists so return that.
	return true, nil
}

// GetSecretsManagerSecretString will return the secret value stored at the given Secrets Manager ARN.
func GetSecretsManagerSecretString(opts *Options, arn string) (string, error) {
	client, err := NewSecretsManagerClient(opts)
	if err != nil {
		return "", err
	}

	secretVal, err := client.GetSecretValue(opts.Context, &secretsmanager.GetSecretValueInput{SecretId: aws.String(arn)})
	if err != nil {
		return "", errors.WithStackTrace(err)
	}
	return aws.ToString(secretVal.SecretString), nil
}

// NewSecretsManagerClient will return a new AWS SDK client for interacting with AWS Secrets Manager.
func NewSecretsManagerClient(opts *Options) (*secretsmanager.Client, error) {
	cfg, err := NewDefaultConfig(opts)
	if err != nil {
		return nil, errors.WithStackTrace(err)
	}
	return secretsmanager.NewFromConfig(cfg), nil
}
