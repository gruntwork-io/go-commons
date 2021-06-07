package lock

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/gruntwork-io/go-commons/errors"
	"github.com/gruntwork-io/go-commons/retry"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	// Terraform requires the DynamoDB table to have a primary key with this name
	attributeLockId = "LockID"

	// Default is to retry for up to 5 minutes
	maxRetriesWaitingForTableToBeActive = 30
	sleepBetweenTableStatusChecks       = 10 * time.Second

	dynamoDbPayPerRequestBillingMode = "PAY_PER_REQUEST"
)

type Options struct {
	Logger *logrus.Logger
	LockString string
	LockTable string
	AwsRegion string
	MaxRetries int
	SleepBetweenRetries time.Duration
	CreateTableIfMissing bool
}

type TimeoutExceeded struct {
	LockTable  string
	LockString string
	Timeout    time.Duration
}

func (err TimeoutExceeded) Error() string {
	return fmt.Sprintf("Timeout trying to acquire lock %s in table %s (timeout was %s)\n", err.LockString, err.LockTable, err.Timeout)
}

type TableNotActiveError struct {
	LockTable string
}

func (err TableNotActiveError) Error() string {
	return fmt.Sprintf("Table %s is not active\n", err.LockTable)
}

// NewAuthenticatedSession gets an AWS Session, checking that the user has credentials properly configured in their environment.
func NewAuthenticatedSession(awsRegion string) (*session.Session, error) {
	sessionOptions := session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config:            *aws.NewConfig().WithRegion(awsRegion),
	}
	sess, err := session.NewSessionWithOptions(sessionOptions)

	if err != nil {
		return nil, errors.WithStackTrace(err)
	}

	if _, err = sess.Config.Credentials.Get(); err != nil {
		return nil, errors.WithStackTrace(err)
	}

	return sess, nil
}

// NewDynamoDb returns an authenticated client object for accessing DynamoDb
func NewDynamoDb(awsRegion string) (*dynamodb.DynamoDB, error) {
	sess, err := NewAuthenticatedSession(awsRegion)
	if err != nil {
		return nil, err
	}
	dynamodbSvc := dynamodb.New(sess)
	return dynamodbSvc, nil
}

// AcquireLock will attempt to acquire a lock in DynamoDB table while taking the configuration options into account.
// We are using DynamoDB to create a table to help us track the lock status of different resources.
// The acquiring of a lock attempts to create a table if the configuration property `CreateTableIfMissing` is set to true,
// otherwise it assumes that such a table already exists. The intention is that we have 1 table per resource in a single region.
// This would allow the locking mechanism to flexibly decide if a resource is locked or not. For test cases where the AWS resource
// is multi-region, or global, the configuration of which regions to use should reflect that.
func AcquireLock(options *Options) error {
	client, err := NewDynamoDb(options.AwsRegion)
	if err != nil {
		options.Logger.Errorf("Error authenticating to AWS: %s\n", err)
		return err
	}

	if options.CreateTableIfMissing == true {
		err := createLockTableIfNecessary(options, client)
		if err != nil {
			return errors.WithStackTrace(err)
		}
	}

	return acquireLockWithRetries(options, client)
}

// acquireLock will attempt to acquire the lock defined by the provided lock string in the configured lock table for the
// configured region.
func acquireLock(options *Options, client *dynamodb.DynamoDB) error {
	options.Logger.Infof("Attempting to acquire lock %s in table %s in region %s\n",
		options.LockString,
		options.LockTable,
		options.AwsRegion,
	)

	dynamodbSvc, err := NewDynamoDb(options.AwsRegion)
	if err != nil {
		options.Logger.Errorf("Error authenticating to AWS: %s\n", err)
		return err
	}

	putParams := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			attributeLockId: {S: aws.String(options.LockString)},
		},
		TableName:           aws.String(options.LockTable),
		ConditionExpression: aws.String(fmt.Sprintf("attribute_not_exists(%s)", attributeLockId)),
	}
	_, err = dynamodbSvc.PutItem(putParams)
	if err != nil {
		options.Logger.Errorf(
			"Error acquiring lock %s in table %s in region %s (already locked?): %s\n",
			options.LockString,
			options.LockTable,
			options.AwsRegion,
			err,
		)
		return errors.WithStackTrace(err)
	}
	options.Logger.Infof("Acquired lock\n")
	return nil
}

// acquireLockWithRetries will attempt to acquire the lock defined by the provided lock string in the configured lock table
// for the configured region. This will retry on failure, until reaching timeout.
func acquireLockWithRetries(options *Options, client *dynamodb.DynamoDB) error {
	return retry.DoWithRetry(
		options.Logger,
		fmt.Sprintf("Trying to acquire DynamoDB lock %s at table %s\n", options.LockString, options.LockTable),
		options.MaxRetries,
		options.SleepBetweenRetries,
		func() error {
			return acquireLock(options, client)
		},
	)
}

// ReleaseLock will attempt to release the lock defined by the provided lock string in the configured lock table for the
// configured region.
func ReleaseLock(options *Options) error {
	client, err := NewDynamoDb(options.AwsRegion)
	if err != nil {
		options.Logger.Errorf("Error authenticating to AWS: %s\n", err)
		return err
	}

	tableExists, err := lockTableExistsAndIsActive(options.LockTable, client)
	if err != nil {
		options.Logger.Errorf("Error checking if DynamoDB table %s exists and is active\n", options.LockTable)
		return err
	}

	if tableExists != true {
		options.Logger.Errorf("DynamoDB table %s does not exist\n", options.LockTable)
		return err
	}

	return retry.DoWithRetry(
		options.Logger,
		fmt.Sprintf("Trying to release DynamoDB lock %s at table %s\n", options.LockString, options.LockTable),
		options.MaxRetries,
		options.SleepBetweenRetries,
		func() error {
			return releaseLock(options, client)
		},
	)
}

// releaseLock will try to delete the DynamoDB item that serves as the lock object
func releaseLock(options *Options, client *dynamodb.DynamoDB) error {
	options.Logger.Infof(
		"Attempting to release lock %s in table %s in region %s\n",
		options.LockString,
		options.LockTable,
		options.AwsRegion,
	)

	dynamodbSvc, err := NewDynamoDb(options.AwsRegion)
	if err != nil {
		options.Logger.Errorf("Error authenticating to AWS: %s\n", err)
		return err
	}

	params := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"LockID": {S: aws.String(options.LockString)},
		},
		TableName: aws.String(options.LockTable),
	}
	_, err = dynamodbSvc.DeleteItem(params)

	if err != nil {
		options.Logger.Errorf(
			"Error releasing lock %s in table %s in region %s: %s\n",
			options.LockString,
			options.LockTable,
			options.AwsRegion,
			err,
		)
		return errors.WithStackTrace(err)
	}
	options.Logger.Infof("Released lock\n")
	return nil
}

// createLockTableIfNecessary will create the lock table in DynamoDB if it doesn't already exist
func createLockTableIfNecessary(options *Options, client *dynamodb.DynamoDB) error {
	tableExists, err := lockTableExistsAndIsActive(options.LockTable, client)
	if err != nil {
		return err
	}

	if !tableExists {
		options.Logger.Infof("Lock table %s does not exist in DynamoDB. Will need to create it just this first time.", options.LockTable)
		return createLockTable(options, client)
	}

	return nil
}

// create a lock table in DynamoDB and wait until it is in "active" state. If the table already exists, merely wait
// until it is in "active" state.
func createLockTable(options *Options, client *dynamodb.DynamoDB) error {
	options.Logger.Infof("Creating table %s in DynamoDB...\n", options.LockTable)

	attributeDefinitions := []*dynamodb.AttributeDefinition{
		{AttributeName: aws.String(attributeLockId), AttributeType: aws.String(dynamodb.ScalarAttributeTypeS)},
	}

	keySchema := []*dynamodb.KeySchemaElement{
		{AttributeName: aws.String(attributeLockId), KeyType: aws.String(dynamodb.KeyTypeHash)},
	}

	_, err := client.CreateTable(&dynamodb.CreateTableInput{
		TableName:            aws.String(options.LockTable),
		BillingMode:          aws.String(dynamoDbPayPerRequestBillingMode),
		AttributeDefinitions: attributeDefinitions,
		KeySchema:            keySchema,
	})

	if err != nil {
		if isTableAlreadyBeingCreatedOrUpdatedError(err) {
			options.Logger.Infof("Looks like someone created table %s at the same time. Will wait for it to be in active state.\n", options.LockTable)
		} else {
			return errors.WithStackTrace(err)
		}
	}

	return waitForTableToBeActive(options, client)
}

// Return true if the given error is the error message returned by AWS when the resource already exists and is being
// updated by someone else
func isTableAlreadyBeingCreatedOrUpdatedError(err error) bool {
	awsErr, isAwsErr := err.(awserr.Error)
	return isAwsErr && awsErr.Code() == "ResourceInUseException"
}

// Wait for the given DynamoDB table to be in the "active" state. If it's not in "active" state, sleep for the
// specified amount of time, and try again, up to a maximum of maxRetries retries.
func waitForTableToBeActive(options *Options, client *dynamodb.DynamoDB) error {
	return retry.DoWithRetry(options.Logger, fmt.Sprintf("Waiting for Table %s to be active...\n", options.LockTable), maxRetriesWaitingForTableToBeActive, sleepBetweenTableStatusChecks,
		func() error {
			isReady, err := lockTableExistsAndIsActive(options.LockTable, client)
			if err != nil {
				return err
			}

			if isReady {
				options.Logger.Infof("Success! Table %s is now in active state.\n", options.LockTable)
				return nil
			}

			return TableNotActiveError{options.LockTable}
		},
	)
}

// lockTableExistsAndIsActive will return true if the lock table exists in DynamoDB and is in "active" state
func lockTableExistsAndIsActive(tableName string, client *dynamodb.DynamoDB) (bool, error) {
	output, err := client.DescribeTable(&dynamodb.DescribeTableInput{TableName: aws.String(tableName)})
	if err != nil {
		if awsErr, isAwsErr := err.(awserr.Error); isAwsErr && awsErr.Code() == "ResourceNotFoundException" {
			return false, nil
		} else {
			return false, errors.WithStackTrace(err)
		}
	}

	return *output.Table.TableStatus == dynamodb.TableStatusActive, nil
}

// GetLockStatus attempts to acquire the lock and check if the expected item is there. If it is - the status is `locked`,
// if the item with the `LockString` is not there, then the status is `not locked`.
func GetLockStatus(options *Options) (*dynamodb.GetItemOutput, error) {
	client, err := NewDynamoDb(options.AwsRegion)
	if err != nil {
		options.Logger.Errorf("Error authenticating to AWS: %s\n", err)
		return nil, err
	}

	getItemParams := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			attributeLockId: {S: aws.String(options.LockString)},
		},
		TableName: aws.String(options.LockTable),
	}

	item, err := client.GetItem(getItemParams)
	if err != nil {
		options.Logger.Errorf("Error authenticating to AWS: %s\n", err)
		return nil, err
	}

	return item, nil
}
