package lock

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"testing"
	"time"

	"github.com/gruntwork-io/go-commons/logging"
	"github.com/stretchr/testify/assert"
)

func TestBlockingAcquireLock(t *testing.T) {
	t.Parallel()

	lockString := "guardduty-blocking-acquire-lock-test"
	maxRetries := 2
	lockTable := "test-dynamo-lock-eu-jam"
	awsRegion := "eu-central-1"
	sleepBetweenRetries := 1 * time.Second

	log := logging.GetLogger("test")

	defer assertLockReleased(t, lockString, lockTable, awsRegion)
	defer ReleaseLock(log, lockString, lockTable, awsRegion)

	log.Infof("Acquiring first lock")
	err := BlockingAcquireLock(log, lockString, lockTable, awsRegion, maxRetries, sleepBetweenRetries)
	assert.NoError(t, err)

	log.Infof("Acquiring second lock")
	err = BlockingAcquireLock(log, lockString, lockTable, awsRegion, maxRetries, sleepBetweenRetries)

	if err == nil {
		assert.Fail(t, "Acquiring of second lock succeeded, but it was expected to fail.")
	}

}

func assertLockReleased(t *testing.T, lockString string, lockTable string, awsRegion string) {
	dynamodbSvc, dbErr := NewDynamoDb(awsRegion)
	assert.NoError(t, dbErr)

	getItemParams := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"LockID": {S: aws.String(lockString)},
		},
		TableName: aws.String(lockTable),
	}

	item, err := dynamodbSvc.GetItem(getItemParams)
	assert.NoError(t, err)

	assert.Empty(t, item.Item)
}
