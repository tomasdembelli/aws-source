package sns

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"
)

type topicClient interface {
	GetTopicAttributes(ctx context.Context, params *sns.GetTopicAttributesInput, optFns ...func(*sns.Options)) (*sns.GetTopicAttributesOutput, error)
	ListTopics(context.Context, *sns.ListTopicsInput, ...func(*sns.Options)) (*sns.ListTopicsOutput, error)
	ListTagsForResource(context.Context, *sns.ListTagsForResourceInput, ...func(*sns.Options)) (*sns.ListTagsForResourceOutput, error)
}

func getTopicFunc(ctx context.Context, client topicClient, scope string, input *sns.GetTopicAttributesInput) (*sdp.Item, error) {
	output, err := client.GetTopicAttributes(ctx, input)
	if err != nil {
		return nil, err
	}

	if output.Attributes == nil {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOTFOUND,
			ErrorString: "get topic attributes response was nil",
		}
	}

	attributes, err := adapters.ToAttributesWithExclude(output.Attributes)
	if err != nil {
		return nil, err
	}

	item := &sdp.Item{
		Type:            "sns-topic",
		UniqueAttribute: "TopicArn",
		Attributes:      attributes,
		Scope:           scope,
	}

	if resourceTags, err := tagsByResourceARN(ctx, client, *input.TopicArn); err == nil {
		item.Tags = tagsToMap(resourceTags)
	}

	if kmsMasterKeyID, err := attributes.Get("kmsMasterKeyId"); err == nil {
		// +overmind:link kms-key
		item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Type:   "kms-key",
				Method: sdp.QueryMethod_GET,
				Query:  fmt.Sprint(kmsMasterKeyID),
				Scope:  scope,
			},
			BlastPropagation: &sdp.BlastPropagation{
				// Changing the key will affect the topic
				In: true,
				// Changing the topic won't affect the key
				Out: false,
			},
		})
	}

	return item, nil
}

//go:generate docgen ../../docs-data
// +overmind:type sns-topic
// +overmind:descriptiveType SNS Topic
// +overmind:get Get an SNS topic by its ARN
// +overmind:list List all SNS topics
// +overmind:search Search SNS topic by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_sns_topic.id

func NewTopicAdapter(client topicClient, accountID string, region string) *adapters.AlwaysGetAdapter[*sns.ListTopicsInput, *sns.ListTopicsOutput, *sns.GetTopicAttributesInput, *sns.GetTopicAttributesOutput, topicClient, *sns.Options] {
	return &adapters.AlwaysGetAdapter[*sns.ListTopicsInput, *sns.ListTopicsOutput, *sns.GetTopicAttributesInput, *sns.GetTopicAttributesOutput, topicClient, *sns.Options]{
		ItemType:        "sns-topic",
		Client:          client,
		AccountID:       accountID,
		Region:          region,
		ListInput:       &sns.ListTopicsInput{},
		AdapterMetadata: TopicMetadata(),
		GetInputMapper: func(scope, query string) *sns.GetTopicAttributesInput {
			return &sns.GetTopicAttributesInput{
				TopicArn: &query,
			}
		},
		ListFuncPaginatorBuilder: func(client topicClient, input *sns.ListTopicsInput) adapters.Paginator[*sns.ListTopicsOutput, *sns.Options] {
			return sns.NewListTopicsPaginator(client, input)
		},
		ListFuncOutputMapper: func(output *sns.ListTopicsOutput, input *sns.ListTopicsInput) ([]*sns.GetTopicAttributesInput, error) {
			var inputs []*sns.GetTopicAttributesInput
			for _, topic := range output.Topics {
				inputs = append(inputs, &sns.GetTopicAttributesInput{
					TopicArn: topic.TopicArn,
				})
			}
			return inputs, nil
		},
		GetFunc: getTopicFunc,
	}
}

func TopicMetadata() sdp.AdapterMetadata {
	return sdp.AdapterMetadata{
		Type:            "sns-topic",
		DescriptiveName: "SNS Topic",
		SupportedQueryMethods: &sdp.AdapterSupportedQueryMethods{
			Get:               true,
			List:              true,
			Search:            true,
			GetDescription:    "Get an SNS topic by its ARN",
			SearchDescription: "Search SNS topic by ARN",
			ListDescription:   "List all SNS topics",
		},
		TerraformMappings: []*sdp.TerraformMapping{
			{TerraformQueryMap: "aws_sns_topic.id"},
		},
		PotentialLinks: []string{"kms-key"},
		Category:       sdp.AdapterCategory_ADAPTER_CATEGORY_CONFIGURATION,
	}
}
