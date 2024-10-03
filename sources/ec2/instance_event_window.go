package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func instanceEventWindowInputMapperGet(scope, query string) (*ec2.DescribeInstanceEventWindowsInput, error) {
	return &ec2.DescribeInstanceEventWindowsInput{
		InstanceEventWindowIds: []string{
			query,
		},
	}, nil
}

func instanceEventWindowInputMapperList(scope string) (*ec2.DescribeInstanceEventWindowsInput, error) {
	return &ec2.DescribeInstanceEventWindowsInput{}, nil
}

func instanceEventWindowOutputMapper(_ context.Context, _ *ec2.Client, scope string, _ *ec2.DescribeInstanceEventWindowsInput, output *ec2.DescribeInstanceEventWindowsOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, ew := range output.InstanceEventWindows {
		attrs, err := sources.ToAttributesWithExclude(ew, "tags")

		if err != nil {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_OTHER,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		item := sdp.Item{
			Type:            "ec2-instance-event-window",
			UniqueAttribute: "InstanceEventWindowId",
			Scope:           scope,
			Attributes:      attrs,
			Tags:            tagsToMap(ew.Tags),
		}

		if at := ew.AssociationTarget; at != nil {
			for _, id := range at.DedicatedHostIds {
				// +overmind:link ec2-host
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ec2-host",
						Method: sdp.QueryMethod_GET,
						Query:  id,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Changing the host won't affect the window
						In: false,
						// Changing the windows will affect the host
						Out: true,
					},
				})
			}

			for _, id := range at.InstanceIds {
				// +overmind:link ec2-instance
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ec2-instance",
						Method: sdp.QueryMethod_GET,
						Query:  id,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Changing the host won't affect the window
						In: false,
						// Changing the windows will affect the instance
						Out: true,
					},
				})
			}
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type ec2-instance-event-window
// +overmind:descriptiveType EC2 Instance Event Window
// +overmind:get Get an event window by ID
// +overmind:list List all event windows
// +overmind:search Search for event windows by ARN
// +overmind:group AWS

func NewInstanceEventWindowSource(client *ec2.Client, accountID string, region string) *sources.DescribeOnlySource[*ec2.DescribeInstanceEventWindowsInput, *ec2.DescribeInstanceEventWindowsOutput, *ec2.Client, *ec2.Options] {
	return &sources.DescribeOnlySource[*ec2.DescribeInstanceEventWindowsInput, *ec2.DescribeInstanceEventWindowsOutput, *ec2.Client, *ec2.Options]{
		Region:          region,
		Client:          client,
		AccountID:       accountID,
		ItemType:        "ec2-instance-event-window",
		AdapterMetadata: InstanceEventWindowMetadata(),
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeInstanceEventWindowsInput) (*ec2.DescribeInstanceEventWindowsOutput, error) {
			return client.DescribeInstanceEventWindows(ctx, input)
		},
		InputMapperGet:  instanceEventWindowInputMapperGet,
		InputMapperList: instanceEventWindowInputMapperList,
		PaginatorBuilder: func(client *ec2.Client, params *ec2.DescribeInstanceEventWindowsInput) sources.Paginator[*ec2.DescribeInstanceEventWindowsOutput, *ec2.Options] {
			return ec2.NewDescribeInstanceEventWindowsPaginator(client, params)
		},
		OutputMapper: instanceEventWindowOutputMapper,
	}
}

func InstanceEventWindowMetadata() sdp.AdapterMetadata {
	return sdp.AdapterMetadata{
		Type:            "ec2-instance-event-window",
		DescriptiveName: "EC2 Instance Event Window",
		SupportedQueryMethods: &sdp.AdapterSupportedQueryMethods{
			Get:               true,
			List:              true,
			Search:            true,
			GetDescription:    "Get an event window by ID",
			ListDescription:   "List all event windows",
			SearchDescription: "Search for event windows by ARN",
		},
		PotentialLinks: []string{"ec2-host", "ec2-instance"},
		Category:       sdp.AdapterCategory_ADAPTER_CATEGORY_CONFIGURATION,
	}
}
