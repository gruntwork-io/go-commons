package lock

import (
	"testing"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"

	"github.com/gruntwork-io/go-commons/logging"
	"github.com/stretchr/testify/assert"
)

//set up one run with a lock for resource X
//set up another run with the same lock for resource X
// -> is expected to wait
// -> is expected to fail creating a new lock as the first one is still in use
// the first lock should release

//func DoWithRetryableErrorsE(t testing.TestingT, actionDescription string, retryableErrors map[string]string, maxRetries int, sleepBetweenRetries time.Duration,
//action func() (string, error)) (string, error) {
func TestBlockingAcquireLock(t *testing.T) {
	t.Parallel()

	lockString := "guardduty-blocking-acquire-lock-test"

	log := logging.GetLogger("")

	defer assertLockReleased(t, lockString, ProjectLockTableName)
	defer ReleaseLock(log, lockString)

	//err := BlockingAcquireLock(log, lockString, maxRetries, sleep)
	//err := BlockingAcquireLock(log, lockString, 3, 30 * time.Second)
	err := BlockingAcquireLock(log, lockString)
	assert.NoError(t, err)
	//acquire the lock for resource X again shortly after
	//err = BlockingAcquireLock(log, lockString, 2, 5 * time.Second)
	err = BlockingAcquireLock(log, lockString)
	if err == nil {
		// FAILS THE TEST
	}

}

func assertLockReleased(t *testing.T, lockString string, lockTable string) {
	dynamodbSvc, dbErr := NewDynamoDb()
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
