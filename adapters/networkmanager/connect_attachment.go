package networkmanager

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"
)

func connectAttachmentGetFunc(ctx context.Context, client *networkmanager.Client, _, query string) (*types.ConnectAttachment, error) {
	out, err := client.GetConnectAttachment(ctx, &networkmanager.GetConnectAttachmentInput{
		AttachmentId: &query,
	})
	if err != nil {
		return nil, err
	}

	return out.ConnectAttachment, nil
}

func connectAttachmentItemMapper(_, scope string, ca *types.ConnectAttachment) (*sdp.Item, error) {
	attributes, err := adapters.ToAttributesWithExclude(ca)

	if err != nil {
		return nil, err
	}

	if ca == nil || ca.Attachment == nil {
		return nil, sdp.NewQueryError(errors.New("attachment is nil for connect attachment"))
	}

	// The uniqueAttributeValue for this is a nested value of AttachmentId:
	attributes.Set("AttachmentId", *ca.Attachment.AttachmentId)

	item := sdp.Item{
		Type:            "networkmanager-connect-attachment",
		UniqueAttribute: "AttachmentId",
		Attributes:      attributes,
		Scope:           scope,
	}

	if ca.Attachment.CoreNetworkId != nil {
		item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				// +overmind:link networkmanager-core-network
				Type:   "networkmanager-core-network",
				Method: sdp.QueryMethod_GET,
				Query:  *ca.Attachment.CoreNetworkId,
				Scope:  scope,
			},
			BlastPropagation: &sdp.BlastPropagation{
				In:  true,
				Out: false,
			},
		})
	}

	if ca.Attachment.CoreNetworkArn != nil {
		if arn, err := adapters.ParseARN(*ca.Attachment.CoreNetworkArn); err == nil {
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					// +overmind:link networkmanager-core-network
					Type:   "networkmanager-core-network",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *ca.Attachment.CoreNetworkArn,
					Scope:  adapters.FormatScope(arn.AccountID, arn.Region),
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: false,
				},
			})
		}
	}

	item.Tags = tagsToMap(ca.Attachment.Tags)

	return &item, nil
}

//go:generate docgen ../../docs-data
// +overmind:type networkmanager-connect-attachment
// +overmind:descriptiveType Networkmanager Connect Attachment
// +overmind:get Get a Networkmanager Connect Attachment by id
// +overmind:group AWS
// +overmind:terraform:queryMap aws_networkmanager_core_network.id

func NewConnectAttachmentAdapter(client *networkmanager.Client, accountID, region string) *adapters.GetListAdapter[*types.ConnectAttachment, *networkmanager.Client, *networkmanager.Options] {
	return &adapters.GetListAdapter[*types.ConnectAttachment, *networkmanager.Client, *networkmanager.Options]{
		Client:          client,
		AccountID:       accountID,
		Region:          region,
		ItemType:        "networkmanager-connect-attachment",
		AdapterMetadata: ConnectAttachmentMetadata(),
		GetFunc: func(ctx context.Context, client *networkmanager.Client, scope string, query string) (*types.ConnectAttachment, error) {
			return connectAttachmentGetFunc(ctx, client, scope, query)
		},
		ItemMapper: connectAttachmentItemMapper,
		ListFunc: func(ctx context.Context, client *networkmanager.Client, scope string) ([]*types.ConnectAttachment, error) {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_NOTFOUND,
				ErrorString: "list not supported for networkmanager-connect-attachment, use get",
			}
		},
	}
}

func ConnectAttachmentMetadata() sdp.AdapterMetadata {
	return sdp.AdapterMetadata{
		Type:            "networkmanager-connect-attachment",
		DescriptiveName: "Networkmanager Connect Attachment",
		SupportedQueryMethods: &sdp.AdapterSupportedQueryMethods{
			Get: true,
		},
		TerraformMappings: []*sdp.TerraformMapping{
			{TerraformQueryMap: "aws_networkmanager_core_network.id"},
		},
		PotentialLinks: []string{"networkmanager-core-network"},
		Category:       sdp.AdapterCategory_ADAPTER_CATEGORY_NETWORK,
	}
}
