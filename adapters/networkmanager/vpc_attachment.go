package networkmanager

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"
)

func vpcAttachmentGetFunc(ctx context.Context, client *networkmanager.Client, _, query string) (*types.VpcAttachment, error) {
	out, err := client.GetVpcAttachment(ctx, &networkmanager.GetVpcAttachmentInput{
		AttachmentId: &query,
	})
	if err != nil {
		return nil, err
	}

	return out.VpcAttachment, nil
}

func vpcAttachmentItemMapper(_, scope string, awsItem *types.VpcAttachment) (*sdp.Item, error) {
	attributes, err := adapters.ToAttributesWithExclude(awsItem)

	if err != nil {
		return nil, err
	}

	// The uniqueAttributeValue for this is a nested value of AttachmentId:
	if awsItem != nil && awsItem.Attachment != nil {
		attributes.Set("AttachmentId", *awsItem.Attachment.AttachmentId)
	}

	item := sdp.Item{
		Type:            "networkmanager-vpc-attachment",
		UniqueAttribute: "AttachmentId",
		Attributes:      attributes,
		Scope:           scope,
		Tags:            tagsToMap(awsItem.Attachment.Tags),
	}

	if awsItem.Attachment != nil && awsItem.Attachment.CoreNetworkId != nil {
		item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				// +overmind:link networkmanager-core-network
				Type:   "networkmanager-core-network",
				Method: sdp.QueryMethod_GET,
				Query:  *awsItem.Attachment.CoreNetworkId,
				Scope:  scope,
			},
			BlastPropagation: &sdp.BlastPropagation{
				In:  true,
				Out: true,
			},
		})

	}

	return &item, nil
}

//go:generate docgen ../../docs-data
// +overmind:type networkmanager-vpc-attachment
// +overmind:descriptiveType Networkmanager VPC Attachment
// +overmind:get Get a Networkmanager VPC Attachment by id
// +overmind:group AWS
// +overmind:terraform:queryMap aws_networkmanager_vpc_attachment.id

func NewVPCAttachmentAdapter(client *networkmanager.Client, accountID, region string) *adapters.GetListAdapter[*types.VpcAttachment, *networkmanager.Client, *networkmanager.Options] {
	return &adapters.GetListAdapter[*types.VpcAttachment, *networkmanager.Client, *networkmanager.Options]{
		Client:          client,
		Region:          region,
		AccountID:       accountID,
		ItemType:        "networkmanager-vpc-attachment",
		AdapterMetadata: VPCAttachmentMetadata(),
		GetFunc: func(ctx context.Context, client *networkmanager.Client, scope string, query string) (*types.VpcAttachment, error) {
			return vpcAttachmentGetFunc(ctx, client, scope, query)
		},
		ItemMapper: vpcAttachmentItemMapper,
		ListFunc: func(ctx context.Context, client *networkmanager.Client, scope string) ([]*types.VpcAttachment, error) {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_NOTFOUND,
				ErrorString: "list not supported for networkmanager-vpc-attachment, use get",
			}
		},
	}
}

func VPCAttachmentMetadata() sdp.AdapterMetadata {
	return sdp.AdapterMetadata{
		Type:            "networkmanager-vpc-attachment",
		DescriptiveName: "Networkmanager VPC Attachment",
		SupportedQueryMethods: &sdp.AdapterSupportedQueryMethods{
			Get:            true,
			GetDescription: "Get a Networkmanager VPC Attachment by id",
		},
		TerraformMappings: []*sdp.TerraformMapping{
			{TerraformQueryMap: "aws_networkmanager_vpc_attachment.id"},
		},
		PotentialLinks: []string{"networkmanager-core-network"},
		Category:       sdp.AdapterCategory_ADAPTER_CATEGORY_NETWORK,
	}
}
