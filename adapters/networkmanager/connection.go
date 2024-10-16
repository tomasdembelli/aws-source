package networkmanager

import (
	"context"
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"
)

func connectionOutputMapper(_ context.Context, _ *networkmanager.Client, scope string, _ *networkmanager.GetConnectionsInput, output *networkmanager.GetConnectionsOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, s := range output.Connections {
		var err error
		var attrs *sdp.ItemAttributes
		attrs, err = adapters.ToAttributesWithExclude(s, "tags")

		if err != nil {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_OTHER,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		if s.GlobalNetworkId == nil || s.ConnectionId == nil {
			return nil, sdp.NewQueryError(errors.New("globalNetworkId or connectionId is nil for connection"))
		}

		attrs.Set("GlobalNetworkIdConnectionId", idWithGlobalNetwork(*s.GlobalNetworkId, *s.ConnectionId))

		item := sdp.Item{
			Type:            "networkmanager-connection",
			UniqueAttribute: "GlobalNetworkIdConnectionId",
			Scope:           scope,
			Attributes:      attrs,
			Tags:            tagsToMap(s.Tags),
			LinkedItemQueries: []*sdp.LinkedItemQuery{
				{
					Query: &sdp.Query{
						// +overmind:link networkmanager-global-network
						Type:   "networkmanager-global-network",
						Method: sdp.QueryMethod_GET,
						Query:  *s.GlobalNetworkId,
						Scope:  scope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						In:  true,
						Out: false,
					},
				},
			},
		}

		if s.LinkId != nil {
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					// +overmind:link networkmanager-link
					Type:   "networkmanager-link",
					Method: sdp.QueryMethod_GET,
					Query:  idWithGlobalNetwork(*s.GlobalNetworkId, *s.LinkId),
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: true,
				},
			})
		}

		if s.ConnectedLinkId != nil {
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					// +overmind:link networkmanager-link
					Type:   "networkmanager-link",
					Method: sdp.QueryMethod_GET,
					Query:  idWithGlobalNetwork(*s.GlobalNetworkId, *s.ConnectedLinkId),
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: true,
				},
			})
		}

		if s.DeviceId != nil {
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					// +overmind:link networkmanager-device
					Type:   "networkmanager-device",
					Method: sdp.QueryMethod_GET,
					Query:  idWithGlobalNetwork(*s.GlobalNetworkId, *s.DeviceId),
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: true,
				},
			})
		}

		if s.ConnectedDeviceId != nil {
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					// +overmind:link networkmanager-device
					Type:   "networkmanager-device",
					Method: sdp.QueryMethod_GET,
					Query:  idWithGlobalNetwork(*s.GlobalNetworkId, *s.ConnectedDeviceId),
					Scope:  scope,
				},
				BlastPropagation: &sdp.BlastPropagation{
					In:  true,
					Out: true,
				},
			})
		}

		switch s.State {
		case types.ConnectionStatePending:
			item.Health = sdp.Health_HEALTH_PENDING.Enum()
		case types.ConnectionStateAvailable:
			item.Health = sdp.Health_HEALTH_OK.Enum()
		case types.ConnectionStateDeleting:
			item.Health = sdp.Health_HEALTH_PENDING.Enum()
		case types.ConnectionStateUpdating:
			item.Health = sdp.Health_HEALTH_PENDING.Enum()
		}

		items = append(items, &item)
	}

	return items, nil
}

//go:generate docgen ../../docs-data
// +overmind:type networkmanager-connection
// +overmind:descriptiveType Networkmanager Connection
// +overmind:get Get a Networkmanager Connection
// +overmind:search Search for Networkmanager Connections by GlobalNetworkId
// +overmind:group AWS
// +overmind:terraform:queryMap aws_networkmanager_connection.arn
// +overmind:terraform:method SEARCH

func NewConnectionAdapter(client *networkmanager.Client, accountID string) *adapters.DescribeOnlyAdapter[*networkmanager.GetConnectionsInput, *networkmanager.GetConnectionsOutput, *networkmanager.Client, *networkmanager.Options] {
	return &adapters.DescribeOnlyAdapter[*networkmanager.GetConnectionsInput, *networkmanager.GetConnectionsOutput, *networkmanager.Client, *networkmanager.Options]{
		Client:    client,
		AccountID: accountID,
		ItemType:  "networkmanager-connection",
		DescribeFunc: func(ctx context.Context, client *networkmanager.Client, input *networkmanager.GetConnectionsInput) (*networkmanager.GetConnectionsOutput, error) {
			return client.GetConnections(ctx, input)
		},
		AdapterMetadata: ConnectionMetadata(),
		InputMapperGet: func(scope, query string) (*networkmanager.GetConnectionsInput, error) {
			// We are using a custom id of {globalNetworkId}|{connectionId}
			sections := strings.Split(query, "|")

			if len(sections) != 2 {
				return nil, &sdp.QueryError{
					ErrorType:   sdp.QueryError_NOTFOUND,
					ErrorString: "invalid query for networkmanager-connection get function",
				}
			}
			return &networkmanager.GetConnectionsInput{
				GlobalNetworkId: &sections[0],
				ConnectionIds: []string{
					sections[1],
				},
			}, nil
		},
		InputMapperList: func(scope string) (*networkmanager.GetConnectionsInput, error) {
			return nil, &sdp.QueryError{
				ErrorType:   sdp.QueryError_NOTFOUND,
				ErrorString: "list not supported for networkmanager-connection, use search",
			}
		},
		PaginatorBuilder: func(client *networkmanager.Client, params *networkmanager.GetConnectionsInput) adapters.Paginator[*networkmanager.GetConnectionsOutput, *networkmanager.Options] {
			return networkmanager.NewGetConnectionsPaginator(client, params)
		},
		OutputMapper: connectionOutputMapper,
		InputMapperSearch: func(ctx context.Context, client *networkmanager.Client, scope, query string) (*networkmanager.GetConnectionsInput, error) {
			// We may search by only globalNetworkId or by using a custom id of {globalNetworkId}|{deviceId}
			sections := strings.Split(query, "|")
			switch len(sections) {
			case 1:
				// globalNetworkId
				return &networkmanager.GetConnectionsInput{
					GlobalNetworkId: &sections[0],
				}, nil
			case 2:
				// {globalNetworkId}|{deviceId}
				return &networkmanager.GetConnectionsInput{
					GlobalNetworkId: &sections[0],
					DeviceId:        &sections[1],
				}, nil
			default:
				return nil, &sdp.QueryError{
					ErrorType:   sdp.QueryError_NOTFOUND,
					ErrorString: "invalid query for networkmanager-connection get function",
				}
			}
		},
	}
}

func ConnectionMetadata() sdp.AdapterMetadata {
	return sdp.AdapterMetadata{
		Type:            "networkmanager-connection",
		DescriptiveName: "Networkmanager Connection",
		SupportedQueryMethods: &sdp.AdapterSupportedQueryMethods{
			Get:               true,
			Search:            true,
			GetDescription:    "Get a Networkmanager Connection",
			SearchDescription: "Search for Networkmanager Connections by GlobalNetworkId",
		},
		TerraformMappings: []*sdp.TerraformMapping{
			{
				TerraformQueryMap: "aws_networkmanager_connection.arn",
				TerraformMethod:   sdp.QueryMethod_SEARCH,
			},
		},
		PotentialLinks: []string{"networkmanager-global-network", "networkmanager-link", "networkmanager-device"},
		Category:       sdp.AdapterCategory_ADAPTER_CATEGORY_NETWORK,
	}
}
