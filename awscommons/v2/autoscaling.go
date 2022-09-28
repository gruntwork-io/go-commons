package awscommons

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	autoscaling_types "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	"github.com/gruntwork-io/go-commons/collections"
	"github.com/gruntwork-io/go-commons/errors"
	"github.com/gruntwork-io/go-commons/retry"
)

// GetAsgByName finds the Auto Scaling Group matching the given name. Returns an error if it cannot find a match.
func GetAsgByName(opts *Options, asgName string) (*autoscaling_types.AutoScalingGroup, error) {
	client, err := NewAutoScalingClient(opts)
	if err != nil {
		return nil, errors.WithStackTrace(err)
	}

	input := &autoscaling.DescribeAutoScalingGroupsInput{AutoScalingGroupNames: []string{asgName}}
	output, err := client.DescribeAutoScalingGroups(opts.Context, input)
	if err != nil {
		return nil, errors.WithStackTrace(err)
	}
	groups := output.AutoScalingGroups
	if len(groups) == 0 {
		return nil, errors.WithStackTrace(NewLookupError("ASG", asgName, "detailed data"))
	}
	return &groups[0], nil
}

// ScaleUp sets the desired capacity, waits until all the instances are available, and returns the new instance IDs.
func ScaleUp(
	opts *Options,
	asgName string,
	originalInstanceIds []string,
	desiredCapacity int32,
	maxRetries int,
	sleepBetweenRetries time.Duration,
) ([]string, error) {
	client, err := NewAutoScalingClient(opts)
	if err != nil {
		return nil, err
	}

	err = setAsgCapacity(client, opts, asgName, desiredCapacity)
	if err != nil {
		if opts.Logger != nil {
			opts.Logger.Errorf("Failed to set ASG desired capacity to %d", desiredCapacity)
			opts.Logger.Errorf("If the capacity is set in AWS, undo by lowering back to the original desired capacity. If the desired capacity is not yet set, triage the error message below and try again.")
		}
		return nil, err
	}

	err = waitForCapacity(opts, asgName, maxRetries, sleepBetweenRetries)
	if err != nil {
		if opts.Logger != nil {
			opts.Logger.Errorf("Timed out waiting for ASG to reach desired capacity.")
			opts.Logger.Errorf("Undo by terminating all the new instances and trying again.")
		}
		return nil, err
	}

	newInstanceIds, err := getLaunchedInstanceIds(opts, asgName, originalInstanceIds)
	if err != nil {
		if opts.Logger != nil {
			opts.Logger.Errorf("Error retrieving information about the ASG.")
			opts.Logger.Errorf("Undo by terminating all the new instances and trying again.")
		}
		return nil, err
	}

	return newInstanceIds, nil
}

// getLaunchedInstanceIds returns a list of the newly launched instance IDs in the ASG, given a list of the old instance
// IDs before any change was made.
func getLaunchedInstanceIds(opts *Options, asgName string, existingInstanceIds []string) ([]string, error) {
	asg, err := GetAsgByName(opts, asgName)
	if err != nil {
		return nil, err
	}
	allInstances := asg.Instances
	allInstanceIds := idsFromAsgInstances(allInstances)
	newInstanceIds := []string{}
	for _, instanceId := range allInstanceIds {
		if !collections.ListContainsElement(existingInstanceIds, instanceId) {
			newInstanceIds = append(newInstanceIds, instanceId)
		}
	}
	return newInstanceIds, nil
}

// setAsgCapacity sets the desired capacity on the auto scaling group. This will not wait for the ASG to expand or
// shrink to that size. See waitForCapacity.
func setAsgCapacity(client *autoscaling.Client, opts *Options, asgName string, desiredCapacity int32) error {
	if opts.Logger != nil {
		opts.Logger.Debugf("Updating ASG %s desired capacity to %d.", asgName, desiredCapacity)
	}

	input := &autoscaling.SetDesiredCapacityInput{
		AutoScalingGroupName: aws.String(asgName),
		DesiredCapacity:      aws.Int32(desiredCapacity),
	}
	_, err := client.SetDesiredCapacity(opts.Context, input)
	if err != nil {
		return errors.WithStackTrace(err)
	}

	if opts.Logger != nil {
		opts.Logger.Debugf("Requested ASG %s desired capacity to be %d.", asgName, desiredCapacity)
	}
	return nil
}

// SetAsgMaxSize sets the max size on the auto scaling group. Note that updating the max size does not typically
// change the cluster size.
func SetAsgMaxSize(client *autoscaling.Client, opts *Options, asgName string, maxSize int32) error {
	if opts.Logger != nil {
		opts.Logger.Debugf("Updating ASG %s max size to %d.", asgName, maxSize)
	}

	input := &autoscaling.UpdateAutoScalingGroupInput{
		AutoScalingGroupName: aws.String(asgName),
		MaxSize:              aws.Int32(maxSize),
	}
	_, err := client.UpdateAutoScalingGroup(opts.Context, input)
	if err != nil {
		return errors.WithStackTrace(err)
	}

	if opts.Logger != nil {
		opts.Logger.Debugf("Requested ASG %s max size to be %d.", asgName, maxSize)
	}
	return nil
}

// waitForCapacity waits for the desired capacity to be reached.
func waitForCapacity(
	opts *Options,
	asgName string,
	maxRetries int,
	sleepBetweenRetries time.Duration,
) error {
	if opts.Logger != nil {
		opts.Logger.Debugf("Waiting for ASG %s to reach desired capacity.", asgName)
	}

	err := retry.DoWithRetry(
		opts.Logger.Logger,
		"Waiting for desired capacity to be reached.",
		maxRetries, sleepBetweenRetries,
		func() error {
			if opts.Logger != nil {
				opts.Logger.Debugf("Checking ASG %s capacity.", asgName)
			}
			asg, err := GetAsgByName(opts, asgName)
			if err != nil {
				// TODO: Should we retry this lookup or fail right away?
				return retry.FatalError{Underlying: err}
			}

			currentCapacity := int32(len(asg.Instances))
			desiredCapacity := *asg.DesiredCapacity

			if currentCapacity == desiredCapacity {
				if opts.Logger != nil {
					opts.Logger.Debugf("ASG %s met desired capacity!", asgName)
				}
				return nil
			}

			if opts.Logger != nil {
				opts.Logger.Debugf("ASG %s not yet at desired capacity %d (current %d).", asgName, desiredCapacity, currentCapacity)
				opts.Logger.Debugf("Waiting for %s...", sleepBetweenRetries)
			}
			return errors.WithStackTrace(fmt.Errorf("still waiting for desired capacity to be reached"))
		},
	)

	if err != nil {
		return NewCouldNotMeetASGCapacityError(
			asgName,
			"Error waiting for ASG desired capacity to be reached.",
		)
	}

	return nil
}

// DetachInstances requests AWS to detach the instances, removing them from the ASG. It will also
// request to auto decrement the desired capacity.
func DetachInstances(opts *Options, asgName string, idList []string) error {
	if opts.Logger != nil {
		opts.Logger.Debugf("Detaching %d instances from ASG %s", len(idList), asgName)
	}

	client, err := NewAutoScalingClient(opts)
	if err != nil {
		return errors.WithStackTrace(err)
	}

	// AWS has a 20 instance limit for this, so we detach in groups of 20 ids
	for _, smallIDList := range collections.BatchListIntoGroupsOf(idList, 20) {
		input := &autoscaling.DetachInstancesInput{
			AutoScalingGroupName:           aws.String(asgName),
			InstanceIds:                    smallIDList,
			ShouldDecrementDesiredCapacity: aws.Bool(true),
		}
		_, err := client.DetachInstances(opts.Context, input)
		if err != nil {
			return errors.WithStackTrace(err)
		}
	}

	if opts.Logger != nil {
		opts.Logger.Debugf("Detached %d instances from ASG %s", len(idList), asgName)
	}
	return nil
}

// idsFromAsgInstances returns a list of the instance IDs given a list of instance representations from the ASG API.
func idsFromAsgInstances(instances []autoscaling_types.Instance) []string {
	idList := []string{}
	for _, inst := range instances {
		idList = append(idList, aws.ToString(inst.InstanceId))
	}
	return idList
}

// NewAutoscalingClient returns a new AWS SDK client for interacting with AWS Autoscaling.
func NewAutoScalingClient(opts *Options) (*autoscaling.Client, error) {
	cfg, err := NewDefaultConfig(opts)
	if err != nil {
		return nil, errors.WithStackTrace(err)
	}
	return autoscaling.NewFromConfig(cfg), nil
}
