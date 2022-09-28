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

// MultipleLookupErrors represents multiple errors found while looking up a resource
type MultipleLookupErrors struct {
	errors []error
}

func (err MultipleLookupErrors) Error() string {
	messages := []string{
		fmt.Sprintf("%d errors found during lookup:", len(err.errors)),
	}

	for _, individualErr := range err.errors {
		messages = append(messages, individualErr.Error())
	}
	return strings.Join(messages, "\n")
}

func (err MultipleLookupErrors) AddError(newErr error) {
	err.errors = append(err.errors, newErr)
}

func (err MultipleLookupErrors) IsEmpty() bool {
	return len(err.errors) == 0
}

func NewMultipleLookupErrors() MultipleLookupErrors {
	return MultipleLookupErrors{[]error{}}
}

// LookupError represents an error related to looking up data on an object.
type LookupError struct {
	objectProperty string
	objectType     string
	objectId       string
}

func (err LookupError) Error() string {
	return fmt.Sprintf("Failed to look up %s for %s with id %s.", err.objectProperty, err.objectType, err.objectId)
}

// NewLookupError constructs a new LookupError object that can be used to return an error related to a look up error.
func NewLookupError(objectType string, objectId string, objectProperty string) LookupError {
	return LookupError{objectProperty: objectProperty, objectType: objectType, objectId: objectId}
}

// CouldNotMeetASGCapacityError represents an error related to waiting for ASG to reach desired capacity.
type CouldNotMeetASGCapacityError struct {
	asgName string
	message string
}

func (err CouldNotMeetASGCapacityError) Error() string {
	return fmt.Sprintf(
		"Could not reach desired capacity of ASG %s: %s",
		err.asgName,
		err.message,
	)
}

func NewCouldNotMeetASGCapacityError(asgName string, message string) CouldNotMeetASGCapacityError {
	return CouldNotMeetASGCapacityError{asgName, message}
}
