package sns

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"
)

type endpointClient interface {
	ListEndpointsByPlatformApplication(ctx context.Context, params *sns.ListEndpointsByPlatformApplicationInput, optFns ...func(*sns.Options)) (*sns.ListEndpointsByPlatformApplicationOutput, error)
	GetEndpointAttributes(ctx context.Context, params *sns.GetEndpointAttributesInput, optFns ...func(*sns.Options)) (*sns.GetEndpointAttributesOutput, error)
	ListTagsForResource(context.Context, *sns.ListTagsForResourceInput, ...func(*sns.Options)) (*sns.ListTagsForResourceOutput, error)
}

func getEndpointFunc(ctx context.Context, client endpointClient, scope string, input *sns.GetEndpointAttributesInput) (*sdp.Item, error) {
	output, err := client.GetEndpointAttributes(ctx, input)
	if err != nil {
		return nil, err
	}

	if output.Attributes == nil {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOTFOUND,
			ErrorString: "get endpoint attributes response was nil",
		}
	}

	attributes, err := adapters.ToAttributesWithExclude(output.Attributes)
	if err != nil {
		return nil, err
	}

	err = attributes.Set("EndpointArn", *input.EndpointArn)
	if err != nil {
		return nil, err
	}

	item := &sdp.Item{
		Type:            "sns-endpoint",
		UniqueAttribute: "EndpointArn",
		Attributes:      attributes,
		Scope:           scope,
	}

	if resourceTags, err := tagsByResourceARN(ctx, client, *input.EndpointArn); err == nil {
		item.Tags = tagsToMap(resourceTags)
	}

	return item, nil
}

//go:generate docgen ../../docs-data
// +overmind:type sns-endpoint
// +overmind:descriptiveType SNS Endpoint
// +overmind:get Get an SNS endpoint by its ARN
// +overmind:search Search SNS endpoints by associated Platform Application ARN
// +overmind:group AWS

func NewEndpointAdapter(client endpointClient, accountID string, region string) *adapters.AlwaysGetAdapter[*sns.ListEndpointsByPlatformApplicationInput, *sns.ListEndpointsByPlatformApplicationOutput, *sns.GetEndpointAttributesInput, *sns.GetEndpointAttributesOutput, endpointClient, *sns.Options] {
	return &adapters.AlwaysGetAdapter[*sns.ListEndpointsByPlatformApplicationInput, *sns.ListEndpointsByPlatformApplicationOutput, *sns.GetEndpointAttributesInput, *sns.GetEndpointAttributesOutput, endpointClient, *sns.Options]{
		ItemType:        "sns-endpoint",
		Client:          client,
		AccountID:       accountID,
		Region:          region,
		DisableList:     true, // This source only supports listing by platform application ARN
		AdapterMetadata: EndpointMetadata(),
		SearchInputMapper: func(scope, query string) (*sns.ListEndpointsByPlatformApplicationInput, error) {
			return &sns.ListEndpointsByPlatformApplicationInput{
				PlatformApplicationArn: &query,
			}, nil
		},
		GetInputMapper: func(scope, query string) *sns.GetEndpointAttributesInput {
			return &sns.GetEndpointAttributesInput{
				EndpointArn: &query,
			}
		},
		ListFuncPaginatorBuilder: func(client endpointClient, input *sns.ListEndpointsByPlatformApplicationInput) adapters.Paginator[*sns.ListEndpointsByPlatformApplicationOutput, *sns.Options] {
			return sns.NewListEndpointsByPlatformApplicationPaginator(client, input)
		},
		ListFuncOutputMapper: func(output *sns.ListEndpointsByPlatformApplicationOutput, input *sns.ListEndpointsByPlatformApplicationInput) ([]*sns.GetEndpointAttributesInput, error) {
			var inputs []*sns.GetEndpointAttributesInput
			for _, endpoint := range output.Endpoints {
				inputs = append(inputs, &sns.GetEndpointAttributesInput{
					EndpointArn: endpoint.EndpointArn,
				})
			}
			return inputs, nil
		},
		GetFunc: getEndpointFunc,
	}
}

func EndpointMetadata() sdp.AdapterMetadata {
	return sdp.AdapterMetadata{
		Type:            "sns-endpoint",
		DescriptiveName: "SNS Endpoint",
		SupportedQueryMethods: &sdp.AdapterSupportedQueryMethods{
			Get:               true,
			Search:            true,
			GetDescription:    "Get an SNS endpoint by its ARN",
			SearchDescription: "Search SNS endpoints by associated Platform Application ARN",
		},
		Category: sdp.AdapterCategory_ADAPTER_CATEGORY_CONFIGURATION,
	}
}
