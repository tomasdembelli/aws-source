package ec2

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func routeTableInputMapperGet(scope string, query string) (*ec2.DescribeRouteTablesInput, error) {
	return &ec2.DescribeRouteTablesInput{
		RouteTableIds: []string{
			query,
		},
	}, nil
}

func routeTableInputMapperList(scope string) (*ec2.DescribeRouteTablesInput, error) {
	return &ec2.DescribeRouteTablesInput{}, nil
}

func routeTableOutputMapper(scope string, _ *ec2.DescribeRouteTablesInput, output *ec2.DescribeRouteTablesOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, rt := range output.RouteTables {
		var err error
		var attrs *sdp.ItemAttributes
		attrs, err = sources.ToAttributesCase(rt)

		if err != nil {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_OTHER,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		item := sdp.Item{
			Type:            "ec2-route-table",
			UniqueAttribute: "routeTableId",
			Scope:           scope,
			Attributes:      attrs,
		}

		for _, assoc := range rt.Associations {
			if assoc.SubnetId != nil {
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "ec2-subnet",
					Method: sdp.QueryMethod_GET,
					Query:  *assoc.SubnetId,
					Scope:  scope,
				})
			}

			if assoc.GatewayId != nil {
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "ec2-internet-gateway",
					Method: sdp.QueryMethod_GET,
					Query:  *assoc.GatewayId,
					Scope:  scope,
				})
			}
		}

		for _, route := range rt.Routes {
			if route.GatewayId != nil {
				if strings.HasPrefix(*route.GatewayId, "igw") {
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
						Type:   "ec2-internet-gateway",
						Method: sdp.QueryMethod_GET,
						Query:  *route.GatewayId,
						Scope:  scope,
					})
				}
				if strings.HasPrefix(*route.GatewayId, "vpce") {
					item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
						Type:   "ec2-vpc-endpoint",
						Method: sdp.QueryMethod_GET,
						Query:  *route.GatewayId,
						Scope:  scope,
					})
				}
			}
			if route.CarrierGatewayId != nil {
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "ec2-carrier-gateway",
					Method: sdp.QueryMethod_GET,
					Query:  *route.CarrierGatewayId,
					Scope:  scope,
				})
			}
			if route.EgressOnlyInternetGatewayId != nil {
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "ec2-egress-only-internet-gateway",
					Method: sdp.QueryMethod_GET,
					Query:  *route.EgressOnlyInternetGatewayId,
					Scope:  scope,
				})
			}
			if route.InstanceId != nil {
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "ec2-instance",
					Method: sdp.QueryMethod_GET,
					Query:  *route.InstanceId,
					Scope:  scope,
				})
			}
			if route.LocalGatewayId != nil {
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "ec2-local-gateway",
					Method: sdp.QueryMethod_GET,
					Query:  *route.LocalGatewayId,
					Scope:  scope,
				})
			}
			if route.NatGatewayId != nil {
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "ec2-nat-gateway",
					Method: sdp.QueryMethod_GET,
					Query:  *route.NatGatewayId,
					Scope:  scope,
				})
			}
			if route.NetworkInterfaceId != nil {
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "ec2-network-interface",
					Method: sdp.QueryMethod_GET,
					Query:  *route.NetworkInterfaceId,
					Scope:  scope,
				})
			}
			if route.TransitGatewayId != nil {
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "ec2-transit-gateway",
					Method: sdp.QueryMethod_GET,
					Query:  *route.TransitGatewayId,
					Scope:  scope,
				})
			}
			if route.VpcPeeringConnectionId != nil {
				item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
					Type:   "ec2-vpc-peering-connection",
					Method: sdp.QueryMethod_GET,
					Query:  *route.VpcPeeringConnectionId,
					Scope:  scope,
				})
			}
		}

		if rt.VpcId != nil {
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.Query{
				Type:   "ec2-vpc",
				Method: sdp.QueryMethod_GET,
				Query:  *rt.VpcId,
				Scope:  scope,
			})
		}

		items = append(items, &item)
	}

	return items, nil
}

func NewRouteTableSource(config aws.Config, accountID string, limit *LimitBucket) *sources.DescribeOnlySource[*ec2.DescribeRouteTablesInput, *ec2.DescribeRouteTablesOutput, *ec2.Client, *ec2.Options] {
	return &sources.DescribeOnlySource[*ec2.DescribeRouteTablesInput, *ec2.DescribeRouteTablesOutput, *ec2.Client, *ec2.Options]{
		Config:    config,
		Client:    ec2.NewFromConfig(config),
		AccountID: accountID,
		ItemType:  "ec2-route-table",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeRouteTablesInput) (*ec2.DescribeRouteTablesOutput, error) {
			<-limit.C // Wait for late limiting
			return client.DescribeRouteTables(ctx, input)
		},
		InputMapperGet:  routeTableInputMapperGet,
		InputMapperList: routeTableInputMapperList,
		PaginatorBuilder: func(client *ec2.Client, params *ec2.DescribeRouteTablesInput) sources.Paginator[*ec2.DescribeRouteTablesOutput, *ec2.Options] {
			return ec2.NewDescribeRouteTablesPaginator(client, params)
		},
		OutputMapper: routeTableOutputMapper,
	}
}
