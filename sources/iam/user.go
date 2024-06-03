package iam

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"

	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

type UserDetails struct {
	User       *types.User
	UserGroups []types.Group
}

func userGetFunc(ctx context.Context, client IAMClient, _, query string) (*UserDetails, error) {
	out, err := client.GetUser(ctx, &iam.GetUserInput{
		UserName: &query,
	})

	if err != nil {
		return nil, err
	}

	details := UserDetails{
		User: out.User,
	}

	if out.User != nil {
		enrichUser(ctx, client, &details)
	}

	return &details, nil
}

// enrichUser Enriches the user with group and tag info
func enrichUser(ctx context.Context, client IAMClient, userDetails *UserDetails) error {
	var err error

	userDetails.UserGroups, err = getUserGroups(ctx, client, userDetails.User.UserName)

	if err != nil {
		return err
	}

	return nil
}

// Gets all of the groups that a user is in
func getUserGroups(ctx context.Context, client IAMClient, userName *string) ([]types.Group, error) {
	var out *iam.ListGroupsForUserOutput
	var err error
	groups := make([]types.Group, 0)

	paginator := iam.NewListGroupsForUserPaginator(client, &iam.ListGroupsForUserInput{
		UserName: userName,
	})

	for paginator.HasMorePages() {
		out, err = paginator.NextPage(ctx)

		if err != nil {
			return nil, err

		}

		groups = append(groups, out.Groups...)
	}

	return groups, nil
}

func userListFunc(ctx context.Context, client IAMClient, _ string) ([]*UserDetails, error) {
	var out *iam.ListUsersOutput
	var err error
	users := make([]types.User, 0)

	paginator := iam.NewListUsersPaginator(client, &iam.ListUsersInput{})

	for paginator.HasMorePages() {
		out, err = paginator.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		users = append(users, out.Users...)
	}

	userDetails := make([]*UserDetails, len(users))

	for i := range users {
		details := UserDetails{
			User: &users[i],
		}

		enrichUser(ctx, client, &details)

		userDetails[i] = &details
	}

	return userDetails, nil
}

func userItemMapper(scope string, awsItem *UserDetails) (*sdp.Item, error) {
	attributes, err := sources.ToAttributesCase(awsItem.User)

	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "iam-user",
		UniqueAttribute: "userName",
		Attributes:      attributes,
		Scope:           scope,
	}

	for _, group := range awsItem.UserGroups {
		// +overmind:link iam-group
		item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
			Query: &sdp.Query{
				Type:   "iam-group",
				Method: sdp.QueryMethod_GET,
				Query:  *group.GroupName,
				Scope:  scope,
			},
			BlastPropagation: &sdp.BlastPropagation{
				// Changing the group can affect the user
				In: true,
				// Changing the user won't affect the group
				Out: false,
			},
		})
	}

	return &item, nil
}

func userListTagsFunc(ctx context.Context, u *UserDetails, client IAMClient) (map[string]string, error) {
	tags := make(map[string]string)

	paginator := iam.NewListUserTagsPaginator(client, &iam.ListUserTagsInput{
		UserName: u.User.UserName,
	})

	for paginator.HasMorePages() {
		out, err := paginator.NextPage(ctx)

		if err != nil {
			return sources.HandleTagsError(ctx, err), nil
		}

		for _, tag := range out.Tags {
			if tag.Key != nil && tag.Value != nil {
				tags[*tag.Key] = *tag.Value
			}
		}
	}

	return tags, nil
}

//go:generate docgen ../../docs-data
// +overmind:type iam-user
// +overmind:descriptiveType IAM User
// +overmind:get Get a user by name
// +overmind:list List all users
// +overmind:search Search for users by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_iam_user.arn
// +overmind:terraform:method SEARCH

func NewUserSource(client *iam.Client, accountID string, region string) *sources.GetListSource[*UserDetails, IAMClient, *iam.Options] {
	return &sources.GetListSource[*UserDetails, IAMClient, *iam.Options]{
		ItemType:      "iam-user",
		Client:        client,
		AccountID:     accountID,
		CacheDuration: 3 * time.Hour, // IAM has very low rate limits, we need to cache for a long time
		Region:        region,
		GetFunc: func(ctx context.Context, client IAMClient, scope, query string) (*UserDetails, error) {
			return userGetFunc(ctx, client, scope, query)
		},
		ListFunc: func(ctx context.Context, client IAMClient, scope string) ([]*UserDetails, error) {
			return userListFunc(ctx, client, scope)
		},
		ListTagsFunc: userListTagsFunc,
		ItemMapper:   userItemMapper,
	}
}
