package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func PlacementGroupInputMapperGet(scope string, query string) (*ec2.DescribePlacementGroupsInput, error) {
	return &ec2.DescribePlacementGroupsInput{
		GroupIds: []string{
			query,
		},
	}, nil
}

func PlacementGroupInputMapperList(scope string) (*ec2.DescribePlacementGroupsInput, error) {
	return &ec2.DescribePlacementGroupsInput{}, nil
}

func PlacementGroupOutputMapper(scope string, output *ec2.DescribePlacementGroupsOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, ng := range output.PlacementGroups {
		var err error
		var attrs *sdp.ItemAttributes
		attrs, err = sources.ToAttributesCase(ng)

		if err != nil {
			return nil, &sdp.ItemRequestError{
				ErrorType:   sdp.ItemRequestError_OTHER,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		item := sdp.Item{
			Type:            "ec2-placement-group",
			UniqueAttribute: "groupId",
			Scope:           scope,
			Attributes:      attrs,
		}

		items = append(items, &item)
	}

	return items, nil
}

func NewPlacementGroupSource(config aws.Config, accountID string) *EC2Source[*ec2.DescribePlacementGroupsInput, *ec2.DescribePlacementGroupsOutput] {
	return &EC2Source[*ec2.DescribePlacementGroupsInput, *ec2.DescribePlacementGroupsOutput]{
		Config:    config,
		AccountID: accountID,
		ItemType:  "ec2-placement-group",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribePlacementGroupsInput, optFns ...func(*ec2.Options)) (*ec2.DescribePlacementGroupsOutput, error) {
			return client.DescribePlacementGroups(ctx, input)
		},
		InputMapperGet:  PlacementGroupInputMapperGet,
		InputMapperList: PlacementGroupInputMapperList,
		OutputMapper:    PlacementGroupOutputMapper,
	}
}