package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"
)

func launchTemplateInputMapperGet(scope string, query string) (*ec2.DescribeLaunchTemplatesInput, error) {
	return &ec2.DescribeLaunchTemplatesInput{
		LaunchTemplateIds: []string{
			query,
		},
	}, nil
}

func launchTemplateInputMapperList(scope string) (*ec2.DescribeLaunchTemplatesInput, error) {
	return &ec2.DescribeLaunchTemplatesInput{}, nil
}

func launchTemplateOutputMapper(_ context.Context, _ *ec2.Client, scope string, _ *ec2.DescribeLaunchTemplatesInput, output *ec2.DescribeLaunchTemplatesOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, LaunchTemplate := range output.LaunchTemplates {
		var err error
		var attrs *sdp.ItemAttributes
		attrs, err = adapters.ToAttributesWithExclude(LaunchTemplate, "tags")

		if err != nil {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_OTHER,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		item := sdp.Item{
			Type:            "ec2-launch-template",
			UniqueAttribute: "LaunchTemplateId",
			Scope:           scope,
			Attributes:      attrs,
			Tags:            tagsToMap(LaunchTemplate.Tags),
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type ec2-launch-template
// +overmind:descriptiveType Launch Template
// +overmind:get Get a launch template by ID
// +overmind:list List all launch templates
// +overmind:search Search for launch templates by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_launch_template.id

func NewLaunchTemplateAdapter(client *ec2.Client, accountID string, region string) *adapters.DescribeOnlyAdapter[*ec2.DescribeLaunchTemplatesInput, *ec2.DescribeLaunchTemplatesOutput, *ec2.Client, *ec2.Options] {
	return &adapters.DescribeOnlyAdapter[*ec2.DescribeLaunchTemplatesInput, *ec2.DescribeLaunchTemplatesOutput, *ec2.Client, *ec2.Options]{
		Region:          region,
		Client:          client,
		AccountID:       accountID,
		ItemType:        "ec2-launch-template",
		AdapterMetadata: LaunchTemplateMetadata(),
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeLaunchTemplatesInput) (*ec2.DescribeLaunchTemplatesOutput, error) {
			return client.DescribeLaunchTemplates(ctx, input)
		},
		InputMapperGet:  launchTemplateInputMapperGet,
		InputMapperList: launchTemplateInputMapperList,
		PaginatorBuilder: func(client *ec2.Client, params *ec2.DescribeLaunchTemplatesInput) adapters.Paginator[*ec2.DescribeLaunchTemplatesOutput, *ec2.Options] {
			return ec2.NewDescribeLaunchTemplatesPaginator(client, params)
		},
		OutputMapper: launchTemplateOutputMapper,
	}
}

func LaunchTemplateMetadata() sdp.AdapterMetadata {
	return sdp.AdapterMetadata{
		Type:            "ec2-launch-template",
		DescriptiveName: "Launch Template",
		SupportedQueryMethods: &sdp.AdapterSupportedQueryMethods{
			Get:               true,
			List:              true,
			Search:            true,
			GetDescription:    "Get a launch template by ID",
			ListDescription:   "List all launch templates",
			SearchDescription: "Search for launch templates by ARN",
		},
		TerraformMappings: []*sdp.TerraformMapping{
			{TerraformQueryMap: "aws_launch_template.id"},
		},
		Category: sdp.AdapterCategory_ADAPTER_CATEGORY_COMPUTE_APPLICATION,
	}
}
