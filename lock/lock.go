package lock

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/gruntwork-io/gruntwork-cli/errors"
)

// We assume that the DynamoDB table will be created prior to using this functionality. 

// NewAuthenticatedSession gets an AWS Session, checking that the user has credentials properly configured in their environment.
func NewAuthenticatedSession() (*session.Session, error) {
	sess, err := session.NewSession(aws.NewConfig().WithRegion(ProjectAwsRegion))
	if err != nil {
		return nil, errors.WithStackTrace(err)
	}

	if _, err = sess.Config.Credentials.Get(); err != nil {
		return nil, errors.WithStackTrace(err)
	}

	return sess, nil
}

// NewDynamoDb returns an authenticated client object for accessing DynamoDb
func NewDynamoDb() (*dynamodb.DynamoDB, error) {
	sess, err := NewAuthenticatedSession()
	if err != nil {
		return nil, err
	}
	dynamodbSvc := dynamodb.New(sess)
	return dynamodbSvc, nil
}

// AcquireLock will attempt to acquire the lock defined by the provided lock string in the configured lock table for the
// configured region.
func AcquireLock(lockString string	) error {
	logger := GetProjectLogger()
	logger.Infof(
		"Attempting to acquire lock %s in table %s in region %s",
		lockString,
		ProjectLockTableName,
		ProjectAwsRegion,
	)

	dynamodbSvc, err := NewDynamoDb()
	if err != nil {
		logger.Errorf("Error authenticating to AWS: %s", err)
		return err
	}

	putParams := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"LockID": {S: aws.String(lockString)},
		},
		TableName:           aws.String(ProjectLockTableName),
		ConditionExpression: aws.String("attribute_not_exists(LockID)"),
	}
	_, err = dynamodbSvc.PutItem(putParams)
	if err != nil {
		logger.Warnf(
			"Error acquiring lock %s in table %s in region %s (already locked?): %s",
			lockString,
			ProjectLockTableName,
			ProjectAwsRegion,
			err,
		)
		return errors.WithStackTrace(err)
	}
	logger.Info("Acquired lock")
	return nil
}

// BlockingAcquireLock will attempt to acquire the lock defined by the provided lock string in the configured lock table
// for the configured region. This will retry on failure, until reaching timeout.
func BlockingAcquireLock(lockString string) error {
	logger := GetProjectLogger()
	logger.Infof(
		"Attempting to acquire lock %s in table %s in region %s, retrying on failure for up to %s",
		lockString,
		ProjectLockTableName,
		ProjectAwsRegion,
		ProjectLockRetryTimeout,
	)

	// Timeout logic inspired by terratest
	// See: https://github.com/gruntwork-io/terratest/blob/master/modules/retry/retry.go
	ctx, cancel := context.WithTimeout(context.Background(), ProjectLockRetryTimeout)
	defer cancel()

	doneChannel := make(chan bool, 1)

	go func() {
		for AcquireLock(lockString) != nil {
			logger.Warnf("Failed to acquire lock %s. Retrying in 5 seconds...", lockString)
			time.Sleep(5 * time.Second)
		}
		doneChannel <- true
	}()
	select {
	case <-doneChannel:
		logger.Infof("Successfully acquired lock %s", lockString)
		return nil
	case <-ctx.Done():
		logger.Errorf("Timed out attempting to acquire lock %s", lockString)
		return LockTimeoutExceeded{LockTable: ProjectLockTableName, LockString: lockString, Timeout: ProjectLockRetryTimeout}
	}
}

// ReleaseLock will attempt to release the lock defined by the provided lock string in the configured lock table for the
// configured region.
func ReleaseLock(lockString string) error {
	logger := GetProjectLogger()
	logger.Infof(
		"Attempting to release lock %s in table %s in region %s",
		lockString,
		ProjectLockTableName,
		ProjectAwsRegion,
	)

	dynamodbSvc, err := NewDynamoDb()
	if err != nil {
		logger.Errorf("Error authenticating to AWS: %s", err)
		return err
	}

	params := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"LockID": {S: aws.String(lockString)},
		},
		TableName: aws.String(ProjectLockTableName),
	}
	_, err = dynamodbSvc.DeleteItem(params)

	if err != nil {
		logger.Errorf(
			"Error releasing lock %s in table %s in region %s: %s",
			lockString,
			ProjectLockTableName,
			ProjectAwsRegion,
			err,
		)
		return errors.WithStackTrace(err)
	}
	logger.Info("Released lock")
	return nil
}