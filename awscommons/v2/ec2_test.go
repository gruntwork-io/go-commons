package awscommons

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	expectedEnabledRegionsInGruntworkTestAccount = []string{
		"ap-northeast-1",
		"ap-northeast-2",
		"ap-northeast-3",
		"ap-south-1",
		"ap-southeast-1",
		"ap-southeast-2",
		"ca-central-1",
		"eu-central-1",
		"eu-north-1",
		"eu-west-1",
		"eu-west-2",
		"eu-west-3",
		"sa-east-1",
		"us-east-1",
		"us-east-2",
		"us-west-1",
		"us-west-2",
	}
)

func TestGetAllEnabledRegions(t *testing.T) {
	t.Parallel()

	opts := NewOptions(DefaultRegion)
	regions, err := GetAllEnabledRegions(opts)
	require.NoError(t, err)
	sort.Strings(regions)

	// This may seem brittle to compare the region list with a static list, but with AWS defaulting new regions to opt
	// out, the list of enabled regions will be fairly static in our test accounts, and thus this will be a more robust
	// check than other alternatives.
	assert.Equal(t, expectedEnabledRegionsInGruntworkTestAccount, regions)
}
