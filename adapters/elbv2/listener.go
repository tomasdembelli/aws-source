package elbv2

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"

	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"
)

func listenerOutputMapper(ctx context.Context, client elbClient, scope string, _ *elbv2.DescribeListenersInput, output *elbv2.DescribeListenersOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	// Get the ARNs so that we can get the tags
	arns := make([]string, 0)

	for _, listener := range output.Listeners {
		if listener.ListenerArn != nil {
			arns = append(arns, *listener.ListenerArn)
		}
	}

	tagsMap := getTagsMap(ctx, client, arns)

	for _, listener := range output.Listeners {
		// Redact the client secret and replace with the first 12 characters of
		// the SHA256 hash so that we can at least tell if it has changed
		for _, action := range listener.DefaultActions {
			if action.AuthenticateOidcConfig != nil {
				if action.AuthenticateOidcConfig.ClientSecret != nil {
					h := sha256.New()
					h.Write([]byte(*action.AuthenticateOidcConfig.ClientSecret))
					sha := base64.URLEncoding.EncodeToString(h.Sum(nil))

					if len(sha) > 12 {
						action.AuthenticateOidcConfig.ClientSecret = adapters.PtrString(fmt.Sprintf("REDACTED (Version: %v)", sha[:11]))
					} else {
						action.AuthenticateOidcConfig.ClientSecret = adapters.PtrString("[REDACTED]")
					}
				}
			}
		}

		attrs, err := adapters.ToAttributesWithExclude(listener)

		if err != nil {
			return nil, err
		}

		var tags map[string]string

		if listener.ListenerArn != nil {
			tags = tagsMap[*listener.ListenerArn]
		}

		item := sdp.Item{
			Type:            "elbv2-listener",
			UniqueAttribute: "ListenerArn",
			Attributes:      attrs,
			Scope:           scope,
			Tags:            tags,
		}

		if listener.LoadBalancerArn != nil {
			if a, err := adapters.ParseARN(*listener.LoadBalancerArn); err == nil {
				// +overmind:link elbv2-load-balancer
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "elbv2-load-balancer",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *listener.LoadBalancerArn,
						Scope:  adapters.FormatScope(a.AccountID, a.Region),
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Load balancers and their listeners are tightly coupled
						In:  true,
						Out: true,
					},
				})

				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "elbv2-rule",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *listener.ListenerArn,
						Scope:  adapters.FormatScope(a.AccountID, a.Region),
					},
					BlastPropagation: &sdp.BlastPropagation{
						// Tightly coupled
						In:  true,
						Out: true,
					},
				})
			}
		}

		for _, cert := range listener.Certificates {
			if cert.CertificateArn != nil {
				if a, err := adapters.ParseARN(*cert.CertificateArn); err == nil {
					// +overmind:link acm-certificate
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
						Query: &sdp.Query{
							Type:   "acm-certificate",
							Method: sdp.QueryMethod_SEARCH,
							Query:  *cert.CertificateArn,
							Scope:  adapters.FormatScope(a.AccountID, a.Region),
						},
						BlastPropagation: &sdp.BlastPropagation{
							// Changing the cert will affect the LB
							In: true,
							// The LB won't affect the cert
							Out: false,
						},
					})
				}
			}
		}

		var requests []*sdp.LinkedItemQuery

		for _, action := range listener.DefaultActions {
			// These types can be returned by `ActionToRequests()`
			// +overmind:link cognito-idp-user-pool
			// +overmind:link http
			// +overmind:link elbv2-target-group

			requests = ActionToRequests(action)
			item.LinkedItemQueries = append(item.LinkedItemQueries, requests...)
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type elbv2-listener
// +overmind:descriptiveType ELB Listener
// +overmind:get Get a listener by ARN
// +overmind:search Search for listeners by load balancer ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_alb_listener.arn
// +overmind:terraform:queryMap aws_lb_listener.arn
// +overmind:terraform:method SEARCH

func NewListenerAdapter(client elbClient, accountID string, region string) *adapters.DescribeOnlyAdapter[*elbv2.DescribeListenersInput, *elbv2.DescribeListenersOutput, elbClient, *elbv2.Options] {
	return &adapters.DescribeOnlyAdapter[*elbv2.DescribeListenersInput, *elbv2.DescribeListenersOutput, elbClient, *elbv2.Options]{
		Region:          region,
		Client:          client,
		AccountID:       accountID,
		ItemType:        "elbv2-listener",
		AdapterMetadata: ListenerMetadata(),
		DescribeFunc: func(ctx context.Context, client elbClient, input *elbv2.DescribeListenersInput) (*elbv2.DescribeListenersOutput, error) {
			return client.DescribeListeners(ctx, input)
		},
		InputMapperGet: func(scope, query string) (*elbv2.DescribeListenersInput, error) {
			return &elbv2.DescribeListenersInput{
				ListenerArns: []string{query},
			}, nil
		},
		InputMapperList: func(scope string) (*elbv2.DescribeListenersInput, error) {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_NOTFOUND,
				ErrorString: "list not supported for elbv2-listener, use search",
			}
		},
		InputMapperSearch: func(ctx context.Context, client elbClient, scope, query string) (*elbv2.DescribeListenersInput, error) {
			// Search by LB ARN
			return &elbv2.DescribeListenersInput{
				LoadBalancerArn: &query,
			}, nil
		},
		PaginatorBuilder: func(client elbClient, params *elbv2.DescribeListenersInput) adapters.Paginator[*elbv2.DescribeListenersOutput, *elbv2.Options] {
			return elbv2.NewDescribeListenersPaginator(client, params)
		},
		OutputMapper: listenerOutputMapper,
	}
}

func ListenerMetadata() sdp.AdapterMetadata {
	return sdp.AdapterMetadata{
		Type:            "elbv2-listener",
		DescriptiveName: "ELB Listener",
		SupportedQueryMethods: &sdp.AdapterSupportedQueryMethods{
			Get:    true,
			Search: true,
		},
		TerraformMappings: []*sdp.TerraformMapping{
			{
				TerraformMethod:   sdp.QueryMethod_SEARCH,
				TerraformQueryMap: "aws_alb_listener.arn",
			},
			{
				TerraformMethod:   sdp.QueryMethod_SEARCH,
				TerraformQueryMap: "aws_lb_listener.arn",
			},
		},
		PotentialLinks: []string{"elbv2-load-balancer", "acm-certificate", "elbv2-rule", "cognito-idp-user-pool", "http", "elbv2-target-group"},
		Category:       sdp.AdapterCategory_ADAPTER_CATEGORY_NETWORK,
	}
}
