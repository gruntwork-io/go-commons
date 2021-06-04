package lock

import (
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/require"
	"testing"
	"time"

	"github.com/gruntwork-io/go-commons/logging"
	"github.com/stretchr/testify/assert"
)

var options = Options {
	logging.GetLogger("test"),
	"test-dynamodb-lock-string-" + random.UniqueId(),
	"test-dynamodb-lock-table",
	"eu-central-1",
	2,
	1 * time.Second,
	true,
}

func TestAcquireLockWithRetries(t *testing.T) {
	t.Parallel()

	defer assertLockReleased(t, &options)
	defer ReleaseLock(&options)

	options.Logger.Infof("Acquiring first lock...\n")
	err := AcquireLock(&options)
	require.NoError(t, err)

	options.Logger.Infof("Acquiring second lock...\n")
	err = AcquireLock(&options)

	if err == nil {
		require.Error(t, err, "Acquiring of second lock succeeded, but it was expected to fail.\n")
	}
}

func assertLockReleased(t *testing.T, options *Options) {
	item, err := GetLockStatus(options)
	require.NoError(t, err)

	assert.Empty(t, item.Item)
}
