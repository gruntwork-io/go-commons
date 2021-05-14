package lock

import (
"fmt"
"testing"
"time"

"github.com/stretchr/testify/assert"
"github.com/gruntwork-io/go-commons/logging"
)

func TestBlockingAcquireLock(t *testing.T) {
	t.Parallel()

	lockString := "guardduty1112-ina"

	log := logging.GetLogger("")
	defer ReleaseLock(log, lockString)
	err := BlockingAcquireLock(lockString)
	time.Sleep(2 * time.Minute)
	if err != nil {
		fmt.Println(err)
	}

	//set up one run with a lock for resource X
	//set up another run with the same lock for resource X
		// -> is expected to wait
		// -> is expected to fail creating a new lock as the first one is still in use
	// the first lock should release

}
