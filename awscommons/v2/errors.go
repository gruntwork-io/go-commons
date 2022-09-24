package awscommons

import (
	"fmt"
	"strings"
)

// MultipleDrainContainerInstanceErrors represents multiple errors found while terminating instances
type MultipleDrainContainerInstanceErrors struct {
	errors []error
}

func (err MultipleDrainContainerInstanceErrors) Error() string {
	messages := []string{
		fmt.Sprintf("%d errors found while draining container instances:", len(err.errors)),
	}

	for _, individualErr := range err.errors {
		messages = append(messages, individualErr.Error())
	}
	return strings.Join(messages, "\n")
}

func (err MultipleDrainContainerInstanceErrors) AddError(newErr error) {
	err.errors = append(err.errors, newErr)
}

func (err MultipleDrainContainerInstanceErrors) IsEmpty() bool {
	return len(err.errors) == 0
}

func NewMultipleDrainContainerInstanceErrors() MultipleDrainContainerInstanceErrors {
	return MultipleDrainContainerInstanceErrors{[]error{}}
}
