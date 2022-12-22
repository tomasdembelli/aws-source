package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func RegionInputMapperGet(scope string, query string) (*ec2.DescribeRegionsInput, error) {
	return &ec2.DescribeRegionsInput{
		RegionNames: []string{
			query,
		},
	}, nil
}

func RegionInputMapperList(scope string) (*ec2.DescribeRegionsInput, error) {
	return &ec2.DescribeRegionsInput{}, nil
}

func RegionOutputMapper(scope string, output *ec2.DescribeRegionsOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, ni := range output.Regions {
		var err error
		var attrs *sdp.ItemAttributes
		attrs, err = sources.ToAttributesCase(ni)

		if err != nil {
			return nil, &sdp.ItemRequestError{
				ErrorType:   sdp.ItemRequestError_OTHER,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		item := sdp.Item{
			Type:            "ec2-region",
			UniqueAttribute: "regionName",
			Scope:           scope,
			Attributes:      attrs,
		}

		items = append(items, &item)
	}

	return items, nil
}

func NewRegionSource(config aws.Config, accountID string) *EC2Source[*ec2.DescribeRegionsInput, *ec2.DescribeRegionsOutput] {
	return &EC2Source[*ec2.DescribeRegionsInput, *ec2.DescribeRegionsOutput]{
		Config:    config,
		AccountID: accountID,
		ItemType:  "ec2-region",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeRegionsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeRegionsOutput, error) {
			return client.DescribeRegions(ctx, input)
		},
		InputMapperGet:  RegionInputMapperGet,
		InputMapperList: RegionInputMapperList,
		OutputMapper:    RegionOutputMapper,
	}
}