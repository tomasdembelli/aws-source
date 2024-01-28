package directconnect

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func virtualGatewayOutputMapper(_ context.Context, _ *directconnect.Client, scope string, _ *directconnect.DescribeVirtualGatewaysInput, output *directconnect.DescribeVirtualGatewaysOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, virtualGateway := range output.VirtualGateways {
		attributes, err := sources.ToAttributesCase(virtualGateway, "tags")
		if err != nil {
			return nil, err
		}

		item := sdp.Item{
			Type:            "directconnect-virtual-gateway",
			UniqueAttribute: "virtualGatewayId",
			Attributes:      attributes,
			Scope:           scope,
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type directconnect-virtual-gateway
// +overmind:descriptiveType Virtual Gateway
// +overmind:get Get a virtual gateway by ID TODO: how?
// +overmind:list List all virtual gateways
// +overmind:search Search virtual gateways by TODO ??
// +overmind:group AWS

func NewVirtualGatewaySource(config aws.Config, accountID string, limit *sources.LimitBucket) *sources.DescribeOnlySource[*directconnect.DescribeVirtualGatewaysInput, *directconnect.DescribeVirtualGatewaysOutput, *directconnect.Client, *directconnect.Options] {
	return &sources.DescribeOnlySource[*directconnect.DescribeVirtualGatewaysInput, *directconnect.DescribeVirtualGatewaysOutput, *directconnect.Client, *directconnect.Options]{
		Config:    config,
		Client:    directconnect.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "directconnect-virtual-gateway",
		DescribeFunc: func(ctx context.Context, client *directconnect.Client, input *directconnect.DescribeVirtualGatewaysInput) (*directconnect.DescribeVirtualGatewaysOutput, error) {
			limit.Wait(ctx) // Wait for rate limiting
			return client.DescribeVirtualGateways(ctx, input)
		},
		InputMapperGet: func(scope, query string) (*directconnect.DescribeVirtualGatewaysInput, error) {
			// query must be the virtual gateway id
			// We will list all virtual gateways and filter by id
			return &directconnect.DescribeVirtualGatewaysInput{}, nil
		},
		InputMapperList: func(scope string) (*directconnect.DescribeVirtualGatewaysInput, error) {
			return &directconnect.DescribeVirtualGatewaysInput{}, nil
		},
		OutputMapper: virtualGatewayOutputMapper,
		InputMapperSearch: func(ctx context.Context, client *directconnect.Client, scope, query string) (*directconnect.DescribeVirtualGatewaysInput, error) {
			return &directconnect.DescribeVirtualGatewaysInput{}, nil
		},
	}
}
