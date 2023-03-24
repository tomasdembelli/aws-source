package elbv2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func targetGroupOutputMapper(scope string, _ *elbv2.DescribeTargetGroupsInput, output *elbv2.DescribeTargetGroupsOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, tg := range output.TargetGroups {
		attrs, err := sources.ToAttributesCase(tg)

		if err != nil {
			return nil, err
		}

		item := sdp.Item{
			Type:            "elbv2-target-group",
			UniqueAttribute: "targetGroupName",
			Attributes:      attrs,
			Scope:           scope,
		}

		if tg.TargetGroupArn != nil {
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
				Type:   "elbv2-target-health",
				Method: sdp.QueryMethod_SEARCH,
				Query:  *tg.TargetGroupArn,
				Scope:  scope,
			})
		}

		if tg.VpcId != nil {
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
				Type:   "ec2-vpc",
				Method: sdp.QueryMethod_GET,
				Query:  *tg.VpcId,
				Scope:  scope,
			})
		}

		for _, lbArn := range tg.LoadBalancerArns {
			if a, err := sources.ParseARN(lbArn); err == nil {
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "elbv2-load-balancer",
					Method: sdp.QueryMethod_SEARCH,
					Query:  lbArn,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				})
			}
		}

		items = append(items, &item)
	}

	return items, nil
}

func NewTargetGroupSource(config aws.Config, accountID string) *sources.DescribeOnlySource[*elbv2.DescribeTargetGroupsInput, *elbv2.DescribeTargetGroupsOutput, *elbv2.Client, *elbv2.Options] {
	return &sources.DescribeOnlySource[*elbv2.DescribeTargetGroupsInput, *elbv2.DescribeTargetGroupsOutput, *elbv2.Client, *elbv2.Options]{
		Config:    config,
		Client:    elbv2.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "elbv2-target-group",
		DescribeFunc: func(ctx context.Context, client *elbv2.Client, input *elbv2.DescribeTargetGroupsInput) (*elbv2.DescribeTargetGroupsOutput, error) {
			return client.DescribeTargetGroups(ctx, input)
		},
		InputMapperGet: func(scope, query string) (*elbv2.DescribeTargetGroupsInput, error) {
			return &elbv2.DescribeTargetGroupsInput{
				Names: []string{query},
			}, nil
		},
		InputMapperList: func(scope string) (*elbv2.DescribeTargetGroupsInput, error) {
			return &elbv2.DescribeTargetGroupsInput{}, nil
		},
		InputMapperSearch: func(ctx context.Context, client *elbv2.Client, scope, query string) (*elbv2.DescribeTargetGroupsInput, error) {
			// Search by load balancer
			return &elbv2.DescribeTargetGroupsInput{
				LoadBalancerArn: &query,
			}, nil
		},
		PaginatorBuilder: func(client *elbv2.Client, params *elbv2.DescribeTargetGroupsInput) sources.Paginator[*elbv2.DescribeTargetGroupsOutput, *elbv2.Options] {
			return elbv2.NewDescribeTargetGroupsPaginator(client, params)
		},
		OutputMapper: targetGroupOutputMapper,
	}
}