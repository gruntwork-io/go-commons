package awscommons

import (
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUploadObjectString(t *testing.T) {
	t.Parallel()

	bucket := "gruntwork-go-commons-test-" + strings.ToLower(random.UniqueId())
	key := "test/obj"
	contents := random.UniqueId()
	region := aws.GetRandomStableRegion(t, nil, nil)
	opts := NewOptions(region)

	defer func() {
		aws.EmptyS3Bucket(t, region, bucket)
		aws.DeleteS3Bucket(t, region, bucket)
	}()
	aws.CreateS3Bucket(t, region, bucket)

	require.NoError(t, UploadObjectString(opts, bucket, key, contents))

	actualContents := aws.GetS3ObjectContents(t, region, bucket, key)
	assert.Equal(t, contents, actualContents)
}
