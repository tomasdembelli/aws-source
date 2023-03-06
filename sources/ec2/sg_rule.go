package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func SecurityGroupRuleInputMapperGet(scope string, query string) (*ec2.DescribeSecurityGroupRulesInput, error) {
	return &ec2.DescribeSecurityGroupRulesInput{
		SecurityGroupRuleIds: []string{
			query,
		},
	}, nil
}

func SecurityGroupRuleInputMapperList(scope string) (*ec2.DescribeSecurityGroupRulesInput, error) {
	return &ec2.DescribeSecurityGroupRulesInput{}, nil
}

func SecurityGroupRuleOutputMapper(scope string, _ *ec2.DescribeSecurityGroupRulesInput, output *ec2.DescribeSecurityGroupRulesOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, securityGroupRule := range output.SecurityGroupRules {
		var err error
		var attrs *sdp.ItemAttributes
		attrs, err = sources.ToAttributesCase(securityGroupRule)

		if err != nil {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_OTHER,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		item := sdp.Item{
			Type:            "ec2-security-group-rule",
			UniqueAttribute: "securityGroupRuleId",
			Scope:           scope,
			Attributes:      attrs,
		}

		if securityGroupRule.GroupId != nil {
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
				Type:   "ec2-security-group",
				Method: sdp.RequestMethod_GET,
				Query:  *securityGroupRule.GroupId,
				Scope:  scope,
			})
		}

		if rg := securityGroupRule.ReferencedGroupInfo; rg != nil {
			if rg.GroupId != nil {
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "ec2-security-group",
					Method: sdp.RequestMethod_GET,
					Query:  *rg.GroupId,
					Scope:  scope,
				})
			}
		}

		items = append(items, &item)
	}

	return items, nil
}

func NewSecurityGroupRuleSource(config aws.Config, accountID string, limit *LimitBucket) *sources.DescribeOnlySource[*ec2.DescribeSecurityGroupRulesInput, *ec2.DescribeSecurityGroupRulesOutput, *ec2.Client, *ec2.Options] {
	return &sources.DescribeOnlySource[*ec2.DescribeSecurityGroupRulesInput, *ec2.DescribeSecurityGroupRulesOutput, *ec2.Client, *ec2.Options]{
		Config:    config,
		Client:    ec2.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "ec2-security-group-rule",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeSecurityGroupRulesInput) (*ec2.DescribeSecurityGroupRulesOutput, error) {
			<-limit.C // Wait for late limiting
			return client.DescribeSecurityGroupRules(ctx, input)
		},
		InputMapperGet:  SecurityGroupRuleInputMapperGet,
		InputMapperList: SecurityGroupRuleInputMapperList,
		PaginatorBuilder: func(client *ec2.Client, params *ec2.DescribeSecurityGroupRulesInput) sources.Paginator[*ec2.DescribeSecurityGroupRulesOutput, *ec2.Options] {
			return ec2.NewDescribeSecurityGroupRulesPaginator(client, params)
		},
		OutputMapper: SecurityGroupRuleOutputMapper,
	}
}
