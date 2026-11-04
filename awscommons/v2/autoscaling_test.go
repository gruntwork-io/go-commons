package awscommons

import (
	"fmt"
	"testing"
	"time"

	awsgo "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/require"
)

func TestGetAsgByNameReturnsCorrectAsg(t *testing.T) {
	t.Parallel()

	uniqueID := random.UniqueId()
	name := fmt.Sprintf("%s-%s", t.Name(), uniqueID)
	otherUniqueID := random.UniqueId()
	otherName := fmt.Sprintf("%s-%s", t.Name(), otherUniqueID)

	region := getRandomRegion(t)
	opts := NewOptions(region)

	defer terminateEc2InstancesByName(t, region, []string{name, otherName})
	defer deleteAutoScalingGroup(t, name, region)
	defer deleteAutoScalingGroup(t, otherName, region)
	createTestAutoScalingGroup(t, name, region, 1)
	createTestAutoScalingGroup(t, otherName, region, 1)

	asg, err := GetAsgByName(opts, name)
	require.NoError(t, err)
	require.Equal(t, *asg.AutoScalingGroupName, name)
}

func TestSetAsgCapacityDeploysNewInstances(t *testing.T) {
	t.Parallel()

	uniqueID := random.UniqueId()
	name := fmt.Sprintf("%s-%s", t.Name(), uniqueID)

	region := getRandomRegion(t)
	opts := NewOptions(region)
	client, err := NewAutoScalingClient(opts)
	require.NoError(t, err)

	defer terminateEc2InstancesByName(t, region, []string{name})
	defer deleteAutoScalingGroup(t, name, region)
	createTestAutoScalingGroup(t, name, region, 1)

	asg, err := GetAsgByName(opts, name)
	require.NoError(t, err)
	existingInstances := asg.Instances

	require.NoError(t, setAsgCapacity(client, opts, name, 2))
	require.NoError(t, waitForCapacity(opts, name, 40, 15*time.Second))

	asg, err = GetAsgByName(opts, name)
	require.NoError(t, err)
	allInstances := asg.Instances
	require.Equal(t, len(allInstances), len(existingInstances)+1)

	existingInstanceIds := idsFromAsgInstances(existingInstances)
	newInstanceIds, err := getLaunchedInstanceIds(opts, name, existingInstanceIds)
	require.NoError(t, err)
	require.Equal(t, len(existingInstanceIds), 1)
	require.Equal(t, len(newInstanceIds), 1)
	require.NotEqual(t, existingInstanceIds[0], newInstanceIds[0])
}

func TestSetAsgCapacityRemovesInstances(t *testing.T) {
	t.Parallel()

	uniqueID := random.UniqueId()
	name := fmt.Sprintf("%s-%s", t.Name(), uniqueID)

	region := getRandomRegion(t)
	opts := NewOptions(region)
	client, err := NewAutoScalingClient(opts)
	require.NoError(t, err)

	defer terminateEc2InstancesByName(t, region, []string{name})
	defer deleteAutoScalingGroup(t, name, region)
	createTestAutoScalingGroup(t, name, region, 2)

	asg, err := GetAsgByName(opts, name)
	require.NoError(t, err)
	existingInstances := asg.Instances

	require.NoError(t, setAsgCapacity(client, opts, name, 1))
	require.NoError(t, waitForCapacity(opts, name, 40, 15*time.Second))

	asg, err = GetAsgByName(opts, name)
	require.NoError(t, err)
	allInstances := asg.Instances
	require.Equal(t, len(allInstances), len(existingInstances)-1)
}

// The following functions were adapted from the tests for cloud-nuke

func getRandomRegion(t *testing.T) string {
	// Use the same regions as those that EKS is available
	approvedRegions := []string{"us-west-2", "us-east-1", "us-east-2", "eu-west-1"}
	return aws.GetRandomRegion(t, approvedRegions, []string{})
}

func createTestAutoScalingGroup(t *testing.T, name string, region string, desiredCount int32) {
	instance := createTestEC2Instance(t, region, name)

	opts := NewOptions(region)
	client, err := NewAutoScalingClient(opts)
	require.NoError(t, err)

	input := &autoscaling.CreateAutoScalingGroupInput{
		AutoScalingGroupName: &name,
		InstanceId:           instance.InstanceId,
		DesiredCapacity:      awsgo.Int32(desiredCount),
		MinSize:              awsgo.Int32(1),
		MaxSize:              awsgo.Int32(3),
	}
	_, err = client.CreateAutoScalingGroup(opts.Context, input)
	require.NoError(t, err)

	waiter := autoscaling.NewGroupExistsWaiter(client)
	err = waiter.Wait(opts.Context, &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []string{name},
	}, 10*time.Minute)
	require.NoError(t, err)

	aws.WaitForCapacity(t, name, region, 40, 15*time.Second)
}

func createTestEC2Instance(t *testing.T, region string, name string) ec2Types.Instance {
	opts := NewOptions(region)
	ec2Client, err := NewEC2Client(opts)
	require.NoError(t, err)

	imageID := aws.GetAmazonLinuxAmi(t, region)
	input := &ec2.RunInstancesInput{
		ImageId:      awsgo.String(imageID),
		InstanceType: ec2Types.InstanceTypeT2Micro,
		MinCount:     awsgo.Int32(1),
		MaxCount:     awsgo.Int32(1),
	}
	runResult, err := ec2Client.RunInstances(opts.Context, input)
	require.NoError(t, err)

	require.NotEqual(t, len(runResult.Instances), 0)

	waiter := ec2.NewInstanceExistsWaiter(ec2Client)
	err = waiter.Wait(opts.Context, &ec2.DescribeInstancesInput{
		Filters: []ec2Types.Filter{
			{
				Name:   awsgo.String("instance-id"),
				Values: []string{*runResult.Instances[0].InstanceId},
			},
		},
	}, 10*time.Minute)
	require.NoError(t, err)

	// Add test tag to the created instance
	_, err = ec2Client.CreateTags(opts.Context, &ec2.CreateTagsInput{
		Resources: []string{*runResult.Instances[0].InstanceId},
		Tags: []ec2Types.Tag{
			{
				Key:   awsgo.String("Name"),
				Value: awsgo.String(name),
			},
		},
	})
	require.NoError(t, err)

	// EC2 Instance must be in a running before this function returns
	runningWaiter := ec2.NewInstanceRunningWaiter(ec2Client)
	err = runningWaiter.Wait(opts.Context, &ec2.DescribeInstancesInput{
		Filters: []ec2Types.Filter{
			{
				Name:   awsgo.String("instance-id"),
				Values: []string{*runResult.Instances[0].InstanceId},
			},
		},
	}, 10*time.Minute)
	require.NoError(t, err)

	return runResult.Instances[0]
}

func terminateEc2InstancesByName(t *testing.T, region string, names []string) {
	for _, name := range names {
		instanceIds := aws.GetEc2InstanceIdsByTag(t, region, "Name", name)
		for _, instanceID := range instanceIds {
			aws.TerminateInstance(t, region, instanceID)
		}
	}
}

func deleteAutoScalingGroup(t *testing.T, name string, region string) {
	// We have to scale ASG down to 0 before we can delete it
	scaleAsgToZero(t, name, region)

	opts := NewOptions(region)
	client, err := NewAutoScalingClient(opts)
	require.NoError(t, err)

	input := &autoscaling.DeleteAutoScalingGroupInput{AutoScalingGroupName: awsgo.String(name)}
	_, err = client.DeleteAutoScalingGroup(opts.Context, input)
	require.NoError(t, err)
	waiter := autoscaling.NewGroupNotExistsWaiter(client)
	err = waiter.Wait(opts.Context, &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []string{name},
	}, 10*time.Minute)
	require.NoError(t, err)
}

func scaleAsgToZero(t *testing.T, name string, region string) {
	opts := NewOptions(region)
	client, err := NewAutoScalingClient(opts)
	require.NoError(t, err)

	input := &autoscaling.UpdateAutoScalingGroupInput{
		AutoScalingGroupName: awsgo.String(name),
		DesiredCapacity:      awsgo.Int32(0),
		MinSize:              awsgo.Int32(0),
		MaxSize:              awsgo.Int32(0),
	}
	_, err = client.UpdateAutoScalingGroup(opts.Context, input)
	require.NoError(t, err)
	aws.WaitForCapacity(t, name, region, 40, 15*time.Second)

	// There is an eventual consistency bug where even though the ASG is scaled down, AWS sometimes still views a
	// scaling activity so we add a 5 second pause here to work around it.
	time.Sleep(5 * time.Second)
}
