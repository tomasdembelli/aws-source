package iam

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"

	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func groupGetFunc(ctx context.Context, client *iam.Client, _, query string) (*types.Group, error) {
	out, err := client.GetGroup(ctx, &iam.GetGroupInput{
		GroupName: &query,
	})

	if err != nil {
		return nil, err
	}

	return out.Group, nil
}

func groupListFunc(ctx context.Context, client *iam.Client, _ string) ([]*types.Group, error) {
	out, err := client.ListGroups(ctx, &iam.ListGroupsInput{})

	if err != nil {
		return nil, err
	}

	zones := make([]*types.Group, 0, len(out.Groups))

	for i := range out.Groups {
		zones = append(zones, &out.Groups[i])
	}

	return zones, nil
}

func groupItemMapper(_, scope string, awsItem *types.Group) (*sdp.Item, error) {
	attributes, err := sources.ToAttributesCase(awsItem)

	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "iam-group",
		UniqueAttribute: "groupName",
		Attributes:      attributes,
		Scope:           scope,
	}

	return &item, nil
}

//go:generate docgen ../../docs-data
// +overmind:type iam-group
// +overmind:descriptiveType IAM Group
// +overmind:get Get a group by name
// +overmind:list List all IAM groups
// +overmind:search Search for a group by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_iam_group.arn
// +overmind:terraform:method SEARCH

func NewGroupSource(client *iam.Client, accountID string, region string) *sources.GetListSource[*types.Group, *iam.Client, *iam.Options] {
	return &sources.GetListSource[*types.Group, *iam.Client, *iam.Options]{
		ItemType:      "iam-group",
		Client:        client,
		CacheDuration: 3 * time.Hour, // IAM has very low rate limits, we need to cache for a long time
		AccountID:     accountID,
		GetFunc: func(ctx context.Context, client *iam.Client, scope, query string) (*types.Group, error) {
			return groupGetFunc(ctx, client, scope, query)
		},
		ListFunc: func(ctx context.Context, client *iam.Client, scope string) ([]*types.Group, error) {
			return groupListFunc(ctx, client, scope)
		},
		ItemMapper: groupItemMapper,
	}
}
