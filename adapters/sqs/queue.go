package sqs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"
)

type sqsClient interface {
	GetQueueAttributes(ctx context.Context, params *sqs.GetQueueAttributesInput, optFns ...func(*sqs.Options)) (*sqs.GetQueueAttributesOutput, error)
	ListQueueTags(ctx context.Context, params *sqs.ListQueueTagsInput, optFns ...func(*sqs.Options)) (*sqs.ListQueueTagsOutput, error)
	ListQueues(context.Context, *sqs.ListQueuesInput, ...func(*sqs.Options)) (*sqs.ListQueuesOutput, error)
}

func getFunc(ctx context.Context, client sqsClient, scope string, input *sqs.GetQueueAttributesInput) (*sdp.Item, error) {
	output, err := client.GetQueueAttributes(ctx, input)
	if err != nil {
		return nil, err
	}

	if output.Attributes == nil {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOTFOUND,
			ErrorString: "get queue attributes response was nil",
		}
	}

	attributes, err := adapters.ToAttributesWithExclude(output.Attributes)
	if err != nil {
		return nil, err
	}

	err = attributes.Set("QueueURL", input.QueueUrl)
	if err != nil {
		return nil, err
	}

	resourceTags, err := tags(ctx, client, *input.QueueUrl)
	if err != nil {
		return nil, &sdp.QueryError{
			ErrorType:   sdp.QueryError_NOTFOUND,
			ErrorString: err.Error(),
		}
	}

	return &sdp.Item{
		Type:            "sqs-queue",
		UniqueAttribute: "QueueURL",
		Attributes:      attributes,
		Scope:           scope,
		Tags:            resourceTags,
	}, nil
}

//go:generate docgen ../../docs-data
// +overmind:type sqs-queue
// +overmind:descriptiveType SQS Queue
// +overmind:get Get an SQS queue attributes by its URL
// +overmind:list List all SQS queue URLs
// +overmind:search Search SQS queue by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_sqs_queue.id

func NewQueueAdapter(client sqsClient, accountID string, region string) *adapters.AlwaysGetAdapter[*sqs.ListQueuesInput, *sqs.ListQueuesOutput, *sqs.GetQueueAttributesInput, *sqs.GetQueueAttributesOutput, sqsClient, *sqs.Options] {
	return &adapters.AlwaysGetAdapter[*sqs.ListQueuesInput, *sqs.ListQueuesOutput, *sqs.GetQueueAttributesInput, *sqs.GetQueueAttributesOutput, sqsClient, *sqs.Options]{
		ItemType:        "sqs-queue",
		Client:          client,
		AccountID:       accountID,
		Region:          region,
		ListInput:       &sqs.ListQueuesInput{},
		AdapterMetadata: QueueMetadata(),
		GetInputMapper: func(scope, query string) *sqs.GetQueueAttributesInput {
			return &sqs.GetQueueAttributesInput{
				QueueUrl: &query,
				// Providing All will return all attributes.
				AttributeNames: []types.QueueAttributeName{"All"},
			}
		},
		ListFuncPaginatorBuilder: func(client sqsClient, input *sqs.ListQueuesInput) adapters.Paginator[*sqs.ListQueuesOutput, *sqs.Options] {
			return sqs.NewListQueuesPaginator(client, input)
		},
		ListFuncOutputMapper: func(output *sqs.ListQueuesOutput, _ *sqs.ListQueuesInput) ([]*sqs.GetQueueAttributesInput, error) {
			var inputs []*sqs.GetQueueAttributesInput
			for _, url := range output.QueueUrls {
				inputs = append(inputs, &sqs.GetQueueAttributesInput{
					QueueUrl: &url,
				})
			}
			return inputs, nil
		},
		GetFunc: getFunc,
	}
}

func QueueMetadata() sdp.AdapterMetadata {
	return sdp.AdapterMetadata{
		Type:            "sqs-queue",
		DescriptiveName: "SQS Queue",
		SupportedQueryMethods: &sdp.AdapterSupportedQueryMethods{
			Get:               true,
			List:              true,
			Search:            true,
			GetDescription:    "Get an SQS queue attributes by its URL",
			ListDescription:   "List all SQS queue URLs",
			SearchDescription: "Search SQS queue by ARN",
		},
		TerraformMappings: []*sdp.TerraformMapping{
			{TerraformQueryMap: "aws_sqs_queue.id"},
		},
		Category: sdp.AdapterCategory_ADAPTER_CATEGORY_COMPUTE_APPLICATION,
	}
}
