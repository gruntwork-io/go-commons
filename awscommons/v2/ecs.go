package awscommons

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecsTypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/gruntwork-io/go-commons/collections"
	"github.com/gruntwork-io/go-commons/errors"
	"github.com/gruntwork-io/go-commons/retry"
)

// GetContainerInstanceArns gets the container instance ARNs of all the EC2 instances in an ECS Cluster.
// ECS container instance ARNs are different from EC2 instance IDs!
// An ECS container instance is an EC2 instance that runs the ECS container agent and has been registered into
// an ECS cluster.
// Example identifiers:
// - EC2 instance ID: i-08e8cfc073db135a9
// - container instance ID: 2db66342-5f69-4782-89a3-f9b707f979ab
// - container instance ARN: arn:aws:ecs:us-east-1:012345678910:container-instance/2db66342-5f69-4782-89a3-f9b707f979ab
func GetContainerInstanceArns(opts *Options, clusterName string) ([]string, error) {
	client, err := NewECSClient(opts)
	if err != nil {
		return nil, err
	}

	if opts.Logger != nil {
		opts.Logger.Debugf("Looking up Container Instance ARNs for ECS cluster %s", clusterName)
	}

	input := &ecs.ListContainerInstancesInput{Cluster: aws.String(clusterName)}
	arns := []string{}
	// Handle pagination by repeatedly making the API call while there is a next token set.
	for {
		result, err := client.ListContainerInstances(opts.Context, input)
		if err != nil {
			return nil, errors.WithStackTrace(err)
		}
		arns = append(arns, result.ContainerInstanceArns...)
		if result.NextToken == nil {
			break
		}
		input.NextToken = result.NextToken
	}

	return arns, nil
}

// StartDrainingContainerInstances puts ECS container instances in DRAINING state so that all ECS Tasks running on
// them are migrated to other container instances. Batches into chunks of 10 because of AWS API limitations.
// (An error occurred InvalidParameterException when calling the UpdateContainerInstancesState
// operation: instanceIds can have at most 10 items.)
func StartDrainingContainerInstances(opts *Options, clusterName string, containerInstanceArns []string) error {
	client, err := NewECSClient(opts)
	if err != nil {
		return err
	}

	batchSize := 10
	numBatches := int(math.Ceil(float64(len(containerInstanceArns) / batchSize)))

	errList := NewMultipleDrainContainerInstanceErrors()
	for batchIdx, batchedArnList := range collections.BatchListIntoGroupsOf(containerInstanceArns, batchSize) {
		batchedArns := aws.StringSlice(batchedArnList)

		if opts.Logger != nil {
			opts.Logger.Debugf("Putting batch %d/%d of container instances in cluster %s into DRAINING state", batchIdx, numBatches, clusterName)
		}
		input := &ecs.UpdateContainerInstancesStateInput{
			Cluster:            aws.String(clusterName),
			ContainerInstances: aws.ToStringSlice(batchedArns),
			Status:             "DRAINING",
		}
		_, err := client.UpdateContainerInstancesState(opts.Context, input)
		if err != nil {
			errList.AddError(err)
			if opts.Logger != nil {
				opts.Logger.Errorf("Encountered error starting to drain container instances in batch %d: %s", batchIdx, err)
				opts.Logger.Errorf("Container Instance ARNs: %s", strings.Join(batchedArnList, ","))
			}
			continue
		}

		if opts.Logger != nil {
			opts.Logger.Debugf("Started draining %d container instances from batch %d", len(batchedArnList), batchIdx)
		}
	}

	if !errList.IsEmpty() {
		return errors.WithStackTrace(errList)
	}

	if opts.Logger != nil {
		opts.Logger.Debugf("Successfully started draining all %d container instances", len(containerInstanceArns))
	}
	return nil
}

// WaitForContainerInstancesToDrain waits until there are no more ECS Tasks running on any of the ECS container
// instances. Batches container instances in groups of 100 because of AWS API limitations.
func WaitForContainerInstancesToDrain(opts *Options, clusterName string, containerInstanceArns []string, start time.Time, timeout time.Duration, maxRetries int, sleepBetweenRetries time.Duration) error {
	client, err := NewECSClient(opts)
	if err != nil {
		return err
	}

	if opts.Logger != nil {
		opts.Logger.Debugf("Checking if all ECS Tasks have been drained from the ECS Container Instances in Cluster %s.", clusterName)
	}

	batchSize := 100
	numBatches := int(math.Ceil(float64(len(containerInstanceArns) / batchSize)))

	err = retry.DoWithRetry(
		opts.Logger,
		"Wait for Container Instances to be Drained",
		maxRetries, sleepBetweenRetries,
		func() error {
			responses := []*ecs.DescribeContainerInstancesOutput{}
			for batchIdx, batchedArnList := range collections.BatchListIntoGroupsOf(containerInstanceArns, batchSize) {
				batchedArns := aws.StringSlice(batchedArnList)

				if opts.Logger != nil {
					opts.Logger.Debugf("Fetching description of batch %d/%d of ECS Instances in Cluster %s.", batchIdx, numBatches, clusterName)
				}
				input := &ecs.DescribeContainerInstancesInput{
					Cluster:            aws.String(clusterName),
					ContainerInstances: aws.ToStringSlice(batchedArns),
				}
				result, err := client.DescribeContainerInstances(opts.Context, input)
				if err != nil {
					return errors.WithStackTrace(err)
				}
				responses = append(responses, result)
			}

			// If we exceeded the timeout, halt with error.
			if timeoutExceeded(start, timeout) {
				return retry.FatalError{Underlying: fmt.Errorf("maximum drain timeout of %s seconds has elapsed and instances are still draining", timeout)}
			}

			// Yay, all done.
			if drained, _ := allInstancesFullyDrained(responses); drained == true {
				if opts.Logger != nil {
					opts.Logger.Debugf("All container instances have been drained in Cluster %s!", clusterName)
				}
				return nil
			}

			// If there's no error, retry.
			if err == nil {
				return errors.WithStackTrace(fmt.Errorf("container instances still draining"))
			}

			// Else, there's an error, halt and fail.
			return retry.FatalError{Underlying: err}
		})
	return errors.WithStackTrace(err)
}

// timeoutExceeded returns true if the amount of time since start has exceeded the timeout.
func timeoutExceeded(start time.Time, timeout time.Duration) bool {
	timeElapsed := time.Now().Sub(start)
	return timeElapsed > timeout
}

// NewECSClient returns a new AWS SDK client for interacting with AWS ECS.
func NewECSClient(opts *Options) (*ecs.Client, error) {
	cfg, err := NewDefaultConfig(opts)
	if err != nil {
		return nil, errors.WithStackTrace(err)
	}
	return ecs.NewFromConfig(cfg), nil
}

func allInstancesFullyDrained(responses []*ecs.DescribeContainerInstancesOutput) (bool, error) {
	for _, response := range responses {
		instances := response.ContainerInstances
		if len(instances) == 0 {
			return false, errors.WithStackTrace(fmt.Errorf("querying DescribeContainerInstances returned no instances"))
		}

		for _, instance := range instances {
			if !instanceFullyDrained(instance) {
				return false, nil
			}
		}
	}
	return true, nil
}

func instanceFullyDrained(instance ecsTypes.ContainerInstance) bool {
	instanceArn := instance.ContainerInstanceArn

	if *instance.Status == "ACTIVE" {
		if opts.Logger != nil {
			opts.Logger.Debugf("The ECS Container Instance %s is still in ACTIVE status", *instanceArn)
		}
		return false
	}
	if instance.PendingTasksCount > 0 {
		if opts.Logger != nil {
			opts.Logger.Debugf("The ECS Container Instance %s still has pending tasks", *instanceArn)
		}
		return false
	}
	if instance.RunningTasksCount > 0 {
		if opts.Logger != nil {
			opts.Logger.Debugf("The ECS Container Instance %s still has running tasks", *instanceArn)
		}
		return false
	}

	return true
}
