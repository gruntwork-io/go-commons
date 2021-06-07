package lock

import (
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
		AwsRegion: "eu-central-1",
		LockTable: "test-dynamodb-lock-table-" + random.UniqueId(),
		LockString: "test-dynamodb-lock-string-" + random.UniqueId(),
		MaxRetries: 2,
		SleepBetweenRetries: 1 * time.Second,
		CreateTableIfMissing: true,
		Logger: logging.GetLogger("test"),
	}

	defer assertLockReleased(t, &options)
	defer ReleaseLock(&options)

	options.Logger.Infof("Acquiring first lock...\n")
	err := AcquireLock(&options)
	require.NoError(t, err)

	options.Logger.Infof("Acquiring second lock...\n")
	err = AcquireLock(&options)

	assert.Error(t, err, "Acquiring of second lock succeeded, but it was expected to fail")
}

func assertLockReleased(t *testing.T, options *Options) {
	item, err := GetLockStatus(options)
	require.NoError(t, err)

	assert.Empty(t, item.Item)
}
