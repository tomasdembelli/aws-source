package dynamodb

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"
)

func tableGetFunc(ctx context.Context, client Client, scope string, input *dynamodb.DescribeTableInput) (*sdp.Item, error) {
	out, err := client.DescribeTable(ctx, input)

	if err != nil {
		return nil, err
	}

	if out.Table == nil {
		return nil, errors.New("returned table is nil")
	}

	table := out.Table

	var nextToken *string
	tagsMap := make(map[string]string)

	// Get the tags for this table, keep looping until we run out of pages
	for {
		tagsOut, err := client.ListTagsOfResource(ctx, &dynamodb.ListTagsOfResourceInput{
			ResourceArn: table.TableArn,
			NextToken:   nextToken,
		})

		if err != nil {
			tagsMap = adapters.HandleTagsError(ctx, err)
			break
		}

		// Add tags to map
		for _, tag := range tagsOut.Tags {
			if tag.Key != nil && tag.Value != nil {
				tagsMap[*tag.Key] = *tag.Value
			}
		}

		nextToken = tagsOut.NextToken

		if nextToken == nil {
			break
		}
	}

	attributes, err := adapters.ToAttributesWithExclude(table)

	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "dynamodb-table",
		UniqueAttribute: "TableName",
		Scope:           scope,
		Attributes:      attributes,
		Tags:            tagsMap,
	}

	var a *adapters.ARN

	streamsOut, err := client.DescribeKinesisStreamingDestination(ctx, &dynamodb.DescribeKinesisStreamingDestinationInput{
		TableName: table.TableName,
	})

	if err == nil {
		for _, dest := range streamsOut.KinesisDataStreamDestinations {
			if dest.StreamArn != nil {
				if a, err = adapters.ParseARN(*dest.StreamArn); err == nil {
					// +overmind:link kinesis-stream
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "kinesis-stream",
							Method: sdp.QueryMethod_SEARCH,
							Query:  *dest.StreamArn,
							Scope:  adapters.FormatScope(a.AccountID, a.Region),
						},
						BlastPropagation: &sdp.BlastPropagation{
							// If you change the stream, it could mean the table
							// is no longer replicated
							In: true,
							// Changing this table will affect the stream and
							// whatever is listening to it
							Out: true,
						},
					})
				}
			}
		}
	}

	if table.RestoreSummary != nil {
		if table.RestoreSummary.SourceBackupArn != nil {
			if a, err = adapters.ParseARN(*table.RestoreSummary.SourceBackupArn); err == nil {
				// +overmind:link backup-recovery-point
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "backup-recovery-point",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *table.RestoreSummary.SourceBackupArn,
						Scope:  adapters.FormatScope(a.AccountID, a.Region),
					},
					BlastPropagation: &sdp.BlastPropagation{
						// The backup is just the source from which the table
						// was created, so I guess you'd say that the recovery
						// point affects the table
						In: true,
						// Changing the table won't affect the recovery point
						Out: false,
					},
				})
			}
		}

		if table.RestoreSummary.SourceTableArn != nil {
			if a, err = adapters.ParseARN(*table.RestoreSummary.SourceTableArn); err == nil {
				// +overmind:link dynamodb-table
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "dynamodb-table",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *table.RestoreSummary.SourceTableArn,
						Scope:  adapters.FormatScope(a.AccountID, a.Region),
					},
					BlastPropagation: &sdp.BlastPropagation{
						// If the table was restored from another table, and
						// this is normal, then changing the source table could
						// affect this one
						In: true,
						// Changing this table won't affect the source table
						Out: false,
					},
				})
			}
		}
	}

	if table.SSEDescription != nil {
		if table.SSEDescription.KMSMasterKeyArn != nil {
			if a, err = adapters.ParseARN(*table.SSEDescription.KMSMasterKeyArn); err == nil {
				// +overmind:link kms-key
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "kms-key",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *table.SSEDescription.KMSMasterKeyArn,
						Scope:  adapters.FormatScope(a.AccountID, a.Region),
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Changing the key could affect the table
						In: true,
						// Changing the table won't affect the key
						Out: false,
					},
				})
			}
		}
	}

	return &item, nil
}

//go:generate docgen ../../docs-data
// +overmind:type dynamodb-table
// +overmind:descriptiveType DynamoDB Table
// +overmind:get Get a DynamoDB table by name
// +overmind:list List all DynamoDB tables
// +overmind:search Search for DynamoDB tables by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_dynamodb_table.arn
// +overmind:terraform:method SEARCH

func NewTableAdapter(client Client, accountID string, region string) *adapters.AlwaysGetAdapter[*dynamodb.ListTablesInput, *dynamodb.ListTablesOutput, *dynamodb.DescribeTableInput, *dynamodb.DescribeTableOutput, Client, *dynamodb.Options] {
	return &adapters.AlwaysGetAdapter[*dynamodb.ListTablesInput, *dynamodb.ListTablesOutput, *dynamodb.DescribeTableInput, *dynamodb.DescribeTableOutput, Client, *dynamodb.Options]{
		ItemType:        "dynamodb-table",
		Client:          client,
		AccountID:       accountID,
		Region:          region,
		GetFunc:         tableGetFunc,
		ListInput:       &dynamodb.ListTablesInput{},
		AdapterMetadata: TableMetadata(),
		GetInputMapper: func(scope, query string) *dynamodb.DescribeTableInput {
			return &dynamodb.DescribeTableInput{
				TableName: &query,
			}
		},
		ListFuncPaginatorBuilder: func(client Client, input *dynamodb.ListTablesInput) adapters.Paginator[*dynamodb.ListTablesOutput, *dynamodb.Options] {
			return dynamodb.NewListTablesPaginator(client, input)
		},
		ListFuncOutputMapper: func(output *dynamodb.ListTablesOutput, input *dynamodb.ListTablesInput) ([]*dynamodb.DescribeTableInput, error) {
			if output == nil {
				return nil, errors.New("cannot map nil output")
			}

			inputs := make([]*dynamodb.DescribeTableInput, 0, len(output.TableNames))

			for i := range output.TableNames {
				inputs = append(inputs, &dynamodb.DescribeTableInput{
					TableName: &output.TableNames[i],
				})
			}

			return inputs, nil
		},
	}
}

func TableMetadata() sdp.AdapterMetadata {
	return sdp.AdapterMetadata{
		Type:            "dynamodb-table",
		DescriptiveName: "DynamoDB Table",
		SupportedQueryMethods: &sdp.AdapterSupportedQueryMethods{
			Get:               true,
			List:              true,
			Search:            true,
			GetDescription:    "Get a DynamoDB table by name",
			ListDescription:   "List all DynamoDB tables",
			SearchDescription: "Search for DynamoDB tables by ARN",
		},
		Category:       sdp.AdapterCategory_ADAPTER_CATEGORY_DATABASE,
		PotentialLinks: []string{"kinesis-stream", "backup-recovery-point", "dynamodb-table", "kms-key"},
		TerraformMappings: []*sdp.TerraformMapping{
			{TerraformMethod: sdp.QueryMethod_SEARCH, TerraformQueryMap: "aws_dynamodb_table.arn"},
		},
	}
}
