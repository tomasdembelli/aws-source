package cloudfront

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func KeyGroupItemMapper(_, scope string, awsItem *types.KeyGroup) (*sdp.Item, error) {
	attributes, err := sources.ToAttributesWithExclude(awsItem)

	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "cloudfront-key-group",
		UniqueAttribute: "Id",
		Attributes:      attributes,
		Scope:           scope,
	}

	return &item, nil
}

//go:generate docgen ../../docs-data
// +overmind:type cloudfront-key-group
// +overmind:descriptiveType CloudFront Key Group
// +overmind:get Get a CloudFront Key Group by ID
// +overmind:list List CloudFront Key Groups
// +overmind:search Search CloudFront Key Groups by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_cloudfront_key_group.id

func NewKeyGroupSource(client *cloudfront.Client, accountID string) *sources.GetListSource[*types.KeyGroup, *cloudfront.Client, *cloudfront.Options] {
	return &sources.GetListSource[*types.KeyGroup, *cloudfront.Client, *cloudfront.Options]{
		ItemType:        "cloudfront-key-group",
		Client:          client,
		AccountID:       accountID,
		Region:          "", // Cloudfront resources aren't tied to a region
		AdapterMetadata: KeyGroupMetadata(),
		GetFunc: func(ctx context.Context, client *cloudfront.Client, scope, query string) (*types.KeyGroup, error) {
			out, err := client.GetKeyGroup(ctx, &cloudfront.GetKeyGroupInput{
				Id: &query,
			})

			if err != nil {
				return nil, err
			}

			return out.KeyGroup, nil
		},
		ListFunc: func(ctx context.Context, client *cloudfront.Client, scope string) ([]*types.KeyGroup, error) {
			out, err := client.ListKeyGroups(ctx, &cloudfront.ListKeyGroupsInput{})

			if err != nil {
				return nil, err
			}

			keyGroups := make([]*types.KeyGroup, 0, len(out.KeyGroupList.Items))

			for _, item := range out.KeyGroupList.Items {
				keyGroups = append(keyGroups, item.KeyGroup)
			}

			return keyGroups, nil
		},
		ItemMapper: KeyGroupItemMapper,
	}
}

func KeyGroupMetadata() sdp.AdapterMetadata {
	return sdp.AdapterMetadata{
		Type:            "cloudfront-key-group",
		DescriptiveName: "CloudFront Key Group",
		SupportedQueryMethods: &sdp.AdapterSupportedQueryMethods{
			Get:               true,
			List:              true,
			Search:            true,
			GetDescription:    "Get a CloudFront Key Group by ID",
			ListDescription:   "List CloudFront Key Groups",
			SearchDescription: "Search CloudFront Key Groups by ARN",
		},
		Category: sdp.AdapterCategory_ADAPTER_CATEGORY_CONFIGURATION,
		TerraformMappings: []*sdp.TerraformMapping{
			{TerraformQueryMap: "aws_cloudfront_key_group.id"},
		},
	}
}
