package awscommons

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sts"
	awsgo "github.com/aws/aws-sdk-go/aws"
	iamv1 "github.com/aws/aws-sdk-go/service/iam"
	stsv1 "github.com/aws/aws-sdk-go/service/sts"
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const defaultRegion = "us-east-1"

func TestAssumeRole(t *testing.T) {
	t.Parallel()

	stsv1Client, err := aws.NewStsClientE(t, defaultRegion)
	require.NoError(t, err)
	identity, err := stsv1Client.GetCallerIdentity(&stsv1.GetCallerIdentityInput{})
	require.NoError(t, err)
	sessionID := awsgo.StringValue(identity.UserId)

	uniqueID := random.UniqueId()
	roleName := fmt.Sprintf("awscommons-assumetest-%s", uniqueID)
	iamv1Client := aws.NewIamClient(t, defaultRegion)

	defer func() {
		_, err := iamv1Client.DeleteRole(&iamv1.DeleteRoleInput{RoleName: awsgo.String(roleName)})
		require.NoError(t, err)
	}()

	assumeRolePolicy := fmt.Sprintf(assumeRolePolicyTemplate, sessionID)
	createIAMRoleOut, err := iamv1Client.CreateRole(&iamv1.CreateRoleInput{
		RoleName:                 awsgo.String(roleName),
		AssumeRolePolicyDocument: awsgo.String(assumeRolePolicy),
	})
	require.NoError(t, err)
	roleARN := awsgo.StringValue(createIAMRoleOut.Role.Arn)

	opts := NewOptions(defaultRegion)
	cfg, err := NewAssumedRoleConfig(opts, roleARN, uniqueID)
	require.NoError(t, err)

	stsClient := sts.NewFromConfig(cfg)
	callerIDOut, err := stsClient.GetCallerIdentity(opts.Context, &sts.GetCallerIdentityInput{})
	require.NoError(t, err)
	assert.Contains(t, awsgo.StringValue(callerIDOut.UserId), roleName)
}

const assumeRolePolicyTemplate = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "%s"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}`
