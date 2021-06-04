package lock

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/require"
	"testing"
	"time"

	"github.com/gruntwork-io/go-commons/logging"
	"github.com/stretchr/testify/assert"
)

func TestAcquireLockWithRetries(t *testing.T) {
	t.Parallel()

	var options = Options {
		logging.GetLogger("test"),
		"test-dynamodb-lock-string-" + random.UniqueId(),
		"test-dynamodb-lock-table",
		"eu-central-1",
		2,
		1 * time.Second,
		true,
	}

	defer assertLockReleased(t, &options)
	defer ReleaseLock(&options)

	options.Logger.Infof("Acquiring first lock")
	err := AcquireLock(&options)
	require.NoError(t, err)

	options.Logger.Infof("Acquiring second lock")
	err = AcquireLock(&options)

	if err == nil {
		require.Error(t, err, "Acquiring of second lock succeeded, but it was expected to fail.")
	}
}

func assertLockReleased(t *testing.T, options *Options) {
	client, dbErr := NewDynamoDb(options.AwsRegion)
	assert.NoError(t, dbErr)

	getItemParams := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"LockID": {S: aws.String(options.LockString)},
		},
		TableName: aws.String(options.LockTable),
	}

	item, err := client.GetItem(getItemParams)
	require.NoError(t, err)

	assert.Empty(t, item.Item)
}
