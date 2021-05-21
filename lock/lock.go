package lock

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/gruntwork-io/gruntwork-cli/errors"
)

var (
	ProjectAwsRegion = "eu-central-1"
	ProjectLockTableName = "test-dynamo-lock-eu-jam"
	ProjectLockRetryTimeout = time.Minute
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

// We assume that the DynamoDB table will be created prior to using the locking mechanism
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
func AcquireLock(log *logrus.Logger, lockString string) error {
	log.Infof("Attempting to acquire lock %s in table %s in region %s\n",
		lockString,
		ProjectLockTableName,
		ProjectAwsRegion,
	)

	dynamodbSvc, err := NewDynamoDb()
	if err != nil {
		log.Errorf("Error authenticating to AWS: %s\n", err)
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
		log.Errorf(
			"Error acquiring lock %s in table %s in region %s (already locked?): %s\n",
			lockString,
			ProjectLockTableName,
			ProjectAwsRegion,
			err,
		)
		return errors.WithStackTrace(err)
	}
	log.Infof("Acquired lock\n")
	return nil
}

// BlockingAcquireLock will attempt to acquire the lock defined by the provided lock string in the configured lock table
// for the configured region. This will retry on failure, until reaching timeout.
func BlockingAcquireLock(log *logrus.Logger, lockString string) error {
	log.Infof(
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
		for AcquireLock(log, lockString) != nil {
			log.Infof("Failed to acquire lock %s. Retrying in 5 seconds...\n", lockString)
			time.Sleep(time.Second * 5)
		}
		doneChannel <- true
	}()
	select {
	case <-doneChannel:
		log.Infof("Successfully acquired lock %s\n", lockString)
		return nil
	case <-ctx.Done():
		log.Errorf("Timed out attempting to acquire lock %s\n", lockString)
		return LockTimeoutExceeded{LockTable: ProjectLockTableName, LockString: lockString, Timeout: ProjectLockRetryTimeout}
	}
}

// ReleaseLock will attempt to release the lock defined by the provided lock string in the configured lock table for the
// configured region.
func ReleaseLock(log *logrus.Logger, lockString string) error {
	log.Infof(
		"Attempting to release lock %s in table %s in region %s\n",
		lockString,
		ProjectLockTableName,
		ProjectAwsRegion,
	)

	dynamodbSvc, err := NewDynamoDb()
	if err != nil {
		log.Errorf("Error authenticating to AWS: %s\n", err)
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
		log.Errorf(
			"Error releasing lock %s in table %s in region %s: %s\n",
			lockString,
			ProjectLockTableName,
			ProjectAwsRegion,
			err,
		)
		return errors.WithStackTrace(err)
	}
	log.Infof("Released lock\n")
	return nil
}