package awscommons

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"

	"github.com/gruntwork-io/go-commons/errors"
)

// GetAllEnabledRegions will return the list of AWS regions (e.g., us-east-1) that are enabled and available to use in
// the account.
func GetAllEnabledRegions(opts *Options) ([]string, error) {
	client, err := NewEC2Client(opts)
	if err != nil {
		return nil, err
	}

	regionsResp, err := client.DescribeRegions(opts.Context, &ec2.DescribeRegionsInput{})
	if err != nil {
		return nil, errors.WithStackTrace(err)
	}

	regions := []string{}
	for _, region := range regionsResp.Regions {
		// TODO: any filters on opt in status needed?
		regions = append(regions, aws.ToString(region.RegionName))
	}
	return regions, nil
}

// NewEC2Client will return a new AWS SDK client for interacting with AWS EC2.
func NewEC2Client(opts *Options) (*ec2.Client, error) {
	cfg, err := NewDefaultConfig(opts)
	if err != nil {
		return nil, errors.WithStackTrace(err)
	}
	return ec2.NewFromConfig(cfg), nil
}
