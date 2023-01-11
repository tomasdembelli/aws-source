package ecs

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

var CapacityProviderIncludeFields = []types.CapacityProviderField{
	types.CapacityProviderFieldTags,
}

func CapacityProviderOutputMapper(scope string, output *ecs.DescribeCapacityProvidersOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	var attributes *sdp.ItemAttributes
	var err error
	var item sdp.Item

	for _, provider := range output.CapacityProviders {
		attributes, err = sources.ToAttributesCase(provider)

		if err != nil {
			return nil, err
		}

		item = sdp.Item{
			Type:            "ecs-capacity-provider",
			UniqueAttribute: "name",
			Attributes:      attributes,
			Scope:           scope,
		}

		if provider.AutoScalingGroupProvider != nil {
			if provider.AutoScalingGroupProvider.AutoScalingGroupArn != nil {
				if a, err := sources.ParseARN(*provider.AutoScalingGroupProvider.AutoScalingGroupArn); err == nil {
					item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
						Type:   "autoscaling-auto-scaling-group",
						Method: sdp.RequestMethod_SEARCH,
						Query:  *provider.AutoScalingGroupProvider.AutoScalingGroupArn,
						Scope:  sources.FormatScope(a.AccountID, a.Region),
					})
				}
			}
		}

		items = append(items, &item)
	}

	return items, nil
}

func NewCapacityProviderSource(config aws.Config, accountID string) *sources.DescribeOnlySource[*ecs.DescribeCapacityProvidersInput, *ecs.DescribeCapacityProvidersOutput, ECSClient, *ecs.Options] {
	return &sources.DescribeOnlySource[*ecs.DescribeCapacityProvidersInput, *ecs.DescribeCapacityProvidersOutput, ECSClient, *ecs.Options]{
		ItemType:  "ecs-capacity-provider",
		Config:    config,
		AccountID: accountID,
		Client:    ecs.NewFromConfig(config),
		DescribeFunc: func(ctx context.Context, client ECSClient, input *ecs.DescribeCapacityProvidersInput) (*ecs.DescribeCapacityProvidersOutput, error) {
			return client.DescribeCapacityProviders(ctx, input)
		},
		InputMapperGet: func(scope, query string) (*ecs.DescribeCapacityProvidersInput, error) {
			return &ecs.DescribeCapacityProvidersInput{
				CapacityProviders: []string{
					query,
				},
				Include: CapacityProviderIncludeFields,
			}, nil
		},
		InputMapperList: func(scope string) (*ecs.DescribeCapacityProvidersInput, error) {
			return &ecs.DescribeCapacityProvidersInput{
				Include: CapacityProviderIncludeFields,
			}, nil
		},
		PaginatorBuilder: func(client ECSClient, params *ecs.DescribeCapacityProvidersInput) sources.Paginator[*ecs.DescribeCapacityProvidersOutput, *ecs.Options] {
			return NewDescribeCapacityProvidersPaginator(client, params)
		},
		OutputMapper: CapacityProviderOutputMapper,
	}
}

// Incredibly annoyingly the go package doesn't provide a paginator builder for
// DescribeCapacityProviders despite the fact that it's paginated, so I'm going
// to create one myself below

// DescribeCapacityProvidersPaginator is a paginator for DescribeCapacityProviders
type DescribeCapacityProvidersPaginator struct {
	client    ECSClient
	params    *ecs.DescribeCapacityProvidersInput
	nextToken *string
	firstPage bool
}

// NewDescribeCapacityProvidersPaginator returns a new DescribeCapacityProvidersPaginator
func NewDescribeCapacityProvidersPaginator(client ECSClient, params *ecs.DescribeCapacityProvidersInput) *DescribeCapacityProvidersPaginator {
	if params == nil {
		params = &ecs.DescribeCapacityProvidersInput{}
	}

	return &DescribeCapacityProvidersPaginator{
		client:    client,
		params:    params,
		firstPage: true,
		nextToken: params.NextToken,
	}
}

// HasMorePages returns a boolean indicating whether more pages are available
func (p *DescribeCapacityProvidersPaginator) HasMorePages() bool {
	return p.firstPage || (p.nextToken != nil && len(*p.nextToken) != 0)
}

// NextPage retrieves the next DescribeCapacityProviders page.
func (p *DescribeCapacityProvidersPaginator) NextPage(ctx context.Context, optFns ...func(*ecs.Options)) (*ecs.DescribeCapacityProvidersOutput, error) {
	if !p.HasMorePages() {
		return nil, fmt.Errorf("no more pages available")
	}

	params := *p.params
	params.NextToken = p.nextToken

	result, err := p.client.DescribeCapacityProviders(ctx, &params, optFns...)
	if err != nil {
		return nil, err
	}
	p.firstPage = false

	prevToken := p.nextToken
	p.nextToken = result.NextToken

	if prevToken != nil &&
		p.nextToken != nil &&
		*prevToken == *p.nextToken {
		p.nextToken = nil
	}

	return result, nil
}