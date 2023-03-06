package ec2

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

// AvailabilityZoneInputMapperGet Maps source calls to the correct input for the AZ API
func AvailabilityZoneInputMapperGet(scope, query string) (*ec2.DescribeAvailabilityZonesInput, error) {
	return &ec2.DescribeAvailabilityZonesInput{
		ZoneNames: []string{
			query,
		},
	}, nil
}

// AvailabilityZoneInputMapperList Maps source calls to the correct input for the AZ API
func AvailabilityZoneInputMapperList(scope string) (*ec2.DescribeAvailabilityZonesInput, error) {
	return &ec2.DescribeAvailabilityZonesInput{}, nil
}

// AvailabilityZoneOutputMapper Maps API output to items
func AvailabilityZoneOutputMapper(scope string, _ *ec2.DescribeAvailabilityZonesInput, output *ec2.DescribeAvailabilityZonesOutput) ([]*sdp.Item, error) {
	if output == nil {
		return nil, errors.New("empty output")
	}

	items := make([]*sdp.Item, len(output.AvailabilityZones))
	var err error
	var attrs *sdp.ItemAttributes

	for i, az := range output.AvailabilityZones {
		attrs, err = sources.ToAttributesCase(az)

		if err != nil {
			return nil, err
		}

		item := sdp.Item{
			Type:            "ec2-availability-zone",
			UniqueAttribute: "zoneName",
			Scope:           scope,
			Attributes:      attrs,
		}

		// Link to region
		if az.RegionName != nil {
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
				Type:   "ec2-region",
				Method: sdp.RequestMethod_GET,
				Query:  *az.RegionName,
				Scope:  scope,
			})
		}

		items[i] = &item
	}

	return items, nil
}

// NewAvailabilityZoneSource Creates a new source for aws-availabilityzone resources
func NewAvailabilityZoneSource(config aws.Config, accountID string, limit *LimitBucket) *sources.DescribeOnlySource[*ec2.DescribeAvailabilityZonesInput, *ec2.DescribeAvailabilityZonesOutput, *ec2.Client, *ec2.Options] {
	return &sources.DescribeOnlySource[*ec2.DescribeAvailabilityZonesInput, *ec2.DescribeAvailabilityZonesOutput, *ec2.Client, *ec2.Options]{
		Config:    config,
		Client:    ec2.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "ec2-availability-zone",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeAvailabilityZonesInput) (*ec2.DescribeAvailabilityZonesOutput, error) {
			<-limit.C // Wait for late limiting
			return client.DescribeAvailabilityZones(ctx, input)
		},
		InputMapperGet:  AvailabilityZoneInputMapperGet,
		InputMapperList: AvailabilityZoneInputMapperList,
		OutputMapper:    AvailabilityZoneOutputMapper,
	}
}
