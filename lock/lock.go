package lock

import (
	"fmt"
	"github.com/gruntwork-io/go-commons/retry"
	"github.com/sirupsen/logrus"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/gruntwork-io/gruntwork-cli/errors"
)

type LockTimeoutExceeded struct {
	LockTable  string
	LockString string
	Timeout    time.Duration
}

func (err LockTimeoutExceeded) Error() string {
	return fmt.Sprintf("Timeout trying to acquire lock %s in table %s (timeout was %s)", err.LockString, err.LockTable, err.Timeout)
}


// We assume that the DynamoDB table will be created prior to using this functionality. 
// NewAuthenticatedSession gets an AWS Session, checking that the user has credentials properly configured in their environment.
func NewAuthenticatedSession(awsRegion string) (*session.Session, error) {
	sess, err := session.NewSession(aws.NewConfig().WithRegion(awsRegion))
	if err != nil {
		return nil, errors.WithStackTrace(err)
	}

	if _, err = sess.Config.Credentials.Get(); err != nil {
		return nil, errors.WithStackTrace(err)
	}

	return sess, nil
}

// We assume that the DynamoDB table will be created prior to using the locking mechanism
// NewDynamoDb returns an authenticated client object for accessing DynamoDb
func NewDynamoDb(awsRegion string) (*dynamodb.DynamoDB, error) {
	sess, err := NewAuthenticatedSession(awsRegion)
	if err != nil {
		return nil, err
	}
	dynamodbSvc := dynamodb.New(sess)
	return dynamodbSvc, nil
}

// AcquireLock will attempt to acquire the lock defined by the provided lock string in the configured lock table for the
// configured region.
func AcquireLock(log *logrus.Logger, lockString string, lockTable string, awsRegion string) error {
	log.Infof("Attempting to acquire lock %s in table %s in region %s\n",
		lockString,
		lockTable,
		awsRegion,
	)

	dynamodbSvc, err := NewDynamoDb(awsRegion)
	if err != nil {
		log.Errorf("Error authenticating to AWS: %s\n", err)
		return err
	}

	putParams := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"LockID": {S: aws.String(lockString)},
		},
		TableName:           aws.String(lockTable),
		ConditionExpression: aws.String("attribute_not_exists(LockID)"),
	}
	_, err = dynamodbSvc.PutItem(putParams)
	if err != nil {
		log.Errorf(
			"Error acquiring lock %s in table %s in region %s (already locked?): %s\n",
			lockString,
			lockTable,
			awsRegion,
			err,
		)
		return errors.WithStackTrace(err)
	}
	log.Infof("Acquired lock\n")
	return nil
}

// BlockingAcquireLock will attempt to acquire the lock defined by the provided lock string in the configured lock table
// for the configured region. This will retry on failure, until reaching timeout.
func BlockingAcquireLock(
	log *logrus.Logger,
	lockString string,
	lockTable string,
	awsRegion string,
	maxRetries int,
	sleepBetweenRetries time.Duration,
	) error {
	log.Infof(
		"Attempting to acquire lock %s in table %s in region %s, retrying on failure for up to %d times %s",
		lockString,
		lockTable,
		awsRegion,
		maxRetries,
		sleepBetweenRetries,
	)

	return retry.DoWithRetry(
		log,
		fmt.Sprintf("Trying to acquire DynamoDB lock %s at table %s", lockString, lockTable),
		maxRetries,
		sleepBetweenRetries,
		func() error {
			return AcquireLock(log, lockString, lockTable, awsRegion)
		},
		)
}

// ReleaseLock will attempt to release the lock defined by the provided lock string in the configured lock table for the
// configured region.
func ReleaseLock(log *logrus.Logger, lockString string, lockTable string, awsRegion string) error {
	log.Infof(
		"Attempting to release lock %s in table %s in region %s\n",
		lockString,
		lockTable,
		awsRegion,
	)

	dynamodbSvc, err := NewDynamoDb(awsRegion)
	if err != nil {
		log.Errorf("Error authenticating to AWS: %s\n", err)
		return err
	}

	params := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"LockID": {S: aws.String(lockString)},
		},
		TableName: aws.String(lockTable),
	}
	_, err = dynamodbSvc.DeleteItem(params)

	if err != nil {
		log.Errorf(
			"Error releasing lock %s in table %s in region %s: %s\n",
			lockString,
			lockTable,
			awsRegion,
			err,
		)
		return errors.WithStackTrace(err)
	}
	log.Infof("Released lock\n")
	return nil
}