package retry

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// DoWithRetry runs the specified action. If it returns a FatalError, return that error immediately.
// If it returns any other type of error, sleep for sleepBetweenRetries and try again, up to a maximum of
// maxRetries retries. If maxRetries is exceeded, return a MaxRetriesExceeded error.
func DoWithRetry(
	logger *logrus.Entry,
	actionDescription string,
	maxRetries int,
	sleepBetweenRetries time.Duration,
	action func() error,
) error {
	_, err := DoWithRetryInterface(
		logger,
		actionDescription,
		maxRetries,
		sleepBetweenRetries,
		func() (interface{}, error) { return nil, action() },
	)
	return err
}

// DoWithRetryInterface runs the specified action. If it returns a value, return that value. If it returns a FatalError,
// return that error immediately. If it returns any other type of error, sleep for sleepBetweenRetries and try again, up
// to a maximum of maxRetries retries. If maxRetries is exceeded, return a MaxRetriesExceeded error.
func DoWithRetryInterface(
	logger *logrus.Entry,
	actionDescription string,
	maxRetries int,
	sleepBetweenRetries time.Duration,
	action func() (interface{}, error),
) (interface{}, error) {
	var output interface{}
	var err error

	for i := 0; i <= maxRetries; i++ {
		logger.Infof(actionDescription)

		output, err = action()
		if err == nil {
			return output, nil
		}

		if _, isFatalErr := err.(FatalError); isFatalErr {
			logger.Infof("Returning due to fatal error: %v", err)
			return output, err
		}

		logger.Infof("%s returned an error: %s. Attempt %d of %d. Sleeping for %s and will retry.", actionDescription, err.Error(), i+1, maxRetries, sleepBetweenRetries)
		time.Sleep(sleepBetweenRetries)
	}

	return output, MaxRetriesExceeded{Description: actionDescription, MaxRetries: maxRetries}
}

// MaxRetriesExceeded is an error that occurs when the maximum amount of retries is exceeded.
type MaxRetriesExceeded struct {
	Description string
	MaxRetries  int
}

func (err MaxRetriesExceeded) Error() string {
	return fmt.Sprintf("'%s' unsuccessful after %d retries", err.Description, err.MaxRetries)
}

// FatalError is a marker interface for errors that should not be retried.
type FatalError struct {
	Underlying error
}

func (err FatalError) Error() string {
	return fmt.Sprintf("FatalError{Underlying: %v}", err.Underlying)
}
