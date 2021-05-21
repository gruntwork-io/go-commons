package lock

import (
"fmt"
"testing"
"time"

"github.com/stretchr/testify/assert"
"github.com/gruntwork-io/go-commons/logging"
)



//set up one run with a lock for resource X
//set up another run with the same lock for resource X
// -> is expected to wait
// -> is expected to fail creating a new lock as the first one is still in use
// the first lock should release

func TestBlockingAcquireLock(t *testing.T) {
	t.Parallel()

	lockString := "guardduty-blocking-acquire-lock-test"

	log := logging.GetLogger("")

	defer ReleaseLock(log, lockString)
	//acquire the lock first time for resource X
	err := BlockingAcquireLock(log, lockString)

	time.Sleep(30 * time.Seconds)
	if err != nil {
		fmt.Println(err)
	}

	//acquire the lock for resource X again shortly after
	err := BlockingAcquireLock(log, lockString)

}
