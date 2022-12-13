package elasticloadbalancing

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

type ELBv2Source struct {
	// Config AWS Config including region and credentials
	Config aws.Config

	// AccountID The id of the account that is being used. This is used by
	// sources as the first element in the scope
	AccountID string

	// client The AWS client to use when making requests
	client        *elbv2.Client
	clientCreated bool
	clientMutex   sync.Mutex
}

func (s *ELBv2Source) Client() *elbv2.Client {
	s.clientMutex.Lock()
	defer s.clientMutex.Unlock()

	// If the client already exists then return it
	if s.clientCreated {
		return s.client
	}

	// Otherwise create a new client from the config
	s.client = elbv2.NewFromConfig(s.Config)
	s.clientCreated = true

	return s.client
}

// Type The type of items that this source is capable of finding
func (s *ELBv2Source) Type() string {
	return "elasticloadbalancing-loadbalancer-v2"
}

// Descriptive name for the source, used in logging and metadata
func (s *ELBv2Source) Name() string {
	return "elasticloadbalancing-v2-aws-source"
}

// List of scopes that this source is capable of find items for. This will be
// in the format {accountID}.{region}
func (s *ELBv2Source) Scopes() []string {
	return []string{
		fmt.Sprintf("%v.%v", s.AccountID, s.Config.Region),
	}
}

// ELBv2Client Collects all functions this code uses from the AWS SDK, for test replacement.
type ELBv2Client interface {
	DescribeLoadBalancers(ctx context.Context, params *elbv2.DescribeLoadBalancersInput, optFns ...func(*elbv2.Options)) (*elbv2.DescribeLoadBalancersOutput, error)
	DescribeListeners(ctx context.Context, params *elbv2.DescribeListenersInput, optFns ...func(*elbv2.Options)) (*elbv2.DescribeListenersOutput, error)
	DescribeTargetGroups(ctx context.Context, params *elbv2.DescribeTargetGroupsInput, optFns ...func(*elbv2.Options)) (*elbv2.DescribeTargetGroupsOutput, error)
	DescribeTargetHealth(ctx context.Context, params *elbv2.DescribeTargetHealthInput, optFns ...func(*elbv2.Options)) (*elbv2.DescribeTargetHealthOutput, error)
}

// Get Get a single item with a given scope and query. The item returned
// should have a UniqueAttributeValue that matches the `query` parameter. The
// ctx parameter contains a golang context object which should be used to allow
// this source to timeout or be cancelled when executing potentially
// long-running actions
func (s *ELBv2Source) Get(ctx context.Context, scope string, query string) (*sdp.Item, error) {
	if scope != s.Scopes()[0] {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOSCOPE,
			ErrorString: fmt.Sprintf("requested scope %v does not match source scope %v", scope, s.Scopes()[0]),
			Scope:       scope,
		}
	}

	return getv2Impl(ctx, s.Client(), scope, query)
}

func getv2Impl(ctx context.Context, client ELBv2Client, scope string, query string) (*sdp.Item, error) {
	lbs, err := client.DescribeLoadBalancers(
		ctx,
		&elbv2.DescribeLoadBalancersInput{
			Names: []string{
				query,
			},
		},
	)

	if err != nil {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: err.Error(),
			Scope:       scope,
		}
	}

	switch len(lbs.LoadBalancers) {
	case 0:
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOTFOUND,
			ErrorString: "elasticloadbalancing-loadbalancer-v2 not found",
			Scope:       scope,
		}
	case 1:
		expanded, err := ExpandLBv2(ctx, client, lbs.LoadBalancers[0])

		if err != nil {
			return nil, &sdp.ItemRequestError{
				ErrorType:   sdp.ItemRequestError_OTHER,
				ErrorString: fmt.Sprintf("error during details expansion: %v", err.Error()),
				Scope:       scope,
			}
		}

		return mapExpandedELBv2ToItem(expanded, scope)
	default:
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: fmt.Sprintf("more than 1 elasticloadbalancing-loadbalancer-v2 found, found: %v", len(lbs.LoadBalancers)),
			Scope:       scope,
		}
	}
}

// List Lists all items in a given scope
func (s *ELBv2Source) List(ctx context.Context, scope string) ([]*sdp.Item, error) {
	if scope != s.Scopes()[0] {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOSCOPE,
			ErrorString: fmt.Sprintf("requested scope %v does not match source scope %v", scope, s.Scopes()[0]),
			Scope:       scope,
		}
	}

	client := s.Client()
	return findV2Impl(ctx, client, scope)
}

func findV2Impl(ctx context.Context, client ELBv2Client, scope string) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)
	var maxResults int32 = 100
	var nextToken *string

	for morePages := true; morePages; {

		lbs, err := client.DescribeLoadBalancers(
			ctx,
			&elbv2.DescribeLoadBalancersInput{
				Marker:   nextToken,
				PageSize: &maxResults,
			},
		)

		if err != nil {
			return items, &sdp.ItemRequestError{
				ErrorType:   sdp.ItemRequestError_OTHER,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		for _, lb := range lbs.LoadBalancers {
			expanded, err := ExpandLBv2(ctx, client, lb)

			if err != nil {
				continue
			}

			var item *sdp.Item

			item, err = mapExpandedELBv2ToItem(expanded, scope)

			if err != nil {
				continue
			}

			items = append(items, item)
		}

		// If there is more data we should store the token so that we can use
		// that. We also need to set morePages to true so that the loop runs
		// again
		nextToken = lbs.NextMarker
		morePages = (nextToken != nil)
	}
	return items, nil
}

// Weight Returns the priority weighting of items returned by this source.
// This is used to resolve conflicts where two sources of the same type
// return an item for a GET request. In this instance only one item can be
// seen on, so the one with the higher weight value will win.
func (s *ELBv2Source) Weight() int {
	return 100
}

type ExpandedTargetGroup struct {
	types.TargetGroup

	TargetHealthDescriptions []types.TargetHealthDescription
}

type ExpandedELBv2 struct {
	types.LoadBalancer

	Listeners    []types.Listener
	TargetGroups []ExpandedTargetGroup
}

func ExpandLBv2(ctx context.Context, client ELBv2Client, lb types.LoadBalancer) (*ExpandedELBv2, error) {
	var listenersOutput *elbv2.DescribeListenersOutput
	var targetGroupsOutput *elbv2.DescribeTargetGroupsOutput
	var targetHealthOutput *elbv2.DescribeTargetHealthOutput
	var err error

	// Copy all fields from LB
	expandedELB := ExpandedELBv2{
		LoadBalancer: lb,
	}

	// Get listeners
	var nextMarker *string
	for morePages := true; morePages; {
		listenersOutput, err = client.DescribeListeners(
			ctx,
			&elbv2.DescribeListenersInput{
				LoadBalancerArn: lb.LoadBalancerArn,
				Marker:          nextMarker,
			},
		)

		if err != nil {
			return nil, err
		}

		if expandedELB.Listeners == nil {
			expandedELB.Listeners = listenersOutput.Listeners
		} else {
			expandedELB.Listeners = append(expandedELB.Listeners, listenersOutput.Listeners...)
		}
		// If there is more data we should store the marker so that we can use
		// that. We also need to set morePages to true so that the loop runs
		// again
		nextMarker = listenersOutput.NextMarker
		morePages = (nextMarker != nil)
	}

	// Get target groups
	targetGroupsOutput, err = client.DescribeTargetGroups(
		ctx,
		&elbv2.DescribeTargetGroupsInput{
			LoadBalancerArn: lb.LoadBalancerArn,
		},
	)

	if err != nil {
		return nil, err
	}

	expandedELB.TargetGroups = make([]ExpandedTargetGroup, 0)

	// For each target group get targets and their health
	for _, tg := range targetGroupsOutput.TargetGroups {
		etg := ExpandedTargetGroup{
			TargetGroup: tg,
		}

		targetHealthOutput, err = client.DescribeTargetHealth(
			ctx,
			&elbv2.DescribeTargetHealthInput{
				TargetGroupArn: tg.TargetGroupArn,
			},
		)

		if err != nil {
			return nil, err
		}

		etg.TargetHealthDescriptions = targetHealthOutput.TargetHealthDescriptions

		expandedELB.TargetGroups = append(expandedELB.TargetGroups, etg)
	}

	return &expandedELB, nil
}

// mapExpandedELBv2ToItem Maps a load balancer to an item
func mapExpandedELBv2ToItem(lb *ExpandedELBv2, scope string) (*sdp.Item, error) {
	attrMap := make(map[string]interface{})

	if lb.LoadBalancerName == nil || *lb.LoadBalancerName == "" {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: "elasticloadbalancing-loadbalancer-v2 was returned with an empty name",
			Scope:       scope,
		}
	}

	item := sdp.Item{
		Type:            "elasticloadbalancing-loadbalancer-v2",
		UniqueAttribute: "name",
		Scope:           scope,
	}

	attrMap["name"] = lb.LoadBalancerName
	attrMap["availabilityZones"] = lb.AvailabilityZones
	attrMap["ipAddressType"] = lb.IpAddressType
	attrMap["scheme"] = lb.Scheme
	attrMap["securityGroups"] = lb.SecurityGroups
	attrMap["type"] = lb.Type
	attrMap["listeners"] = lb.Listeners
	attrMap["targetGroups"] = lb.TargetGroups
	attrMap["canonicalHostedZoneId"] = lb.CanonicalHostedZoneId
	attrMap["loadBalancerArn"] = lb.LoadBalancerArn
	attrMap["customerOwnedIpv4Pool"] = lb.CustomerOwnedIpv4Pool
	attrMap["state"] = lb.State

	if lb.CreatedTime != nil {
		attrMap["createdTime"] = lb.CreatedTime.String()
	}

	if lb.DNSName != nil {
		attrMap["dNSName"] = lb.DNSName

		item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
			Type:   "dns",
			Method: sdp.RequestMethod_GET,
			Query:  *lb.DNSName,
			Scope:  "global",
		})
	}

	if lb.VpcId != nil {
		attrMap["vpcId"] = lb.VpcId

		item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
			Type:   "ec2-vpc",
			Method: sdp.RequestMethod_GET,
			Query:  *lb.VpcId,
			Scope:  scope,
		})
	}

	for _, tg := range lb.TargetGroups {
		for _, healthDescription := range tg.TargetHealthDescriptions {
			if target := healthDescription.Target; target != nil {
				if id := target.Id; id != nil {
					// The ID of the target. If the target type of the target group is instance,
					// specify an instance ID. If the target type is ip, specify an IP address. If the
					// target type is lambda, specify the ARN of the Lambda function. If the target
					// type is alb, specify the ARN of the Application Load Balancer target.
					if net.ParseIP(*id) != nil {
						item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
							Type:   "ip",
							Method: sdp.RequestMethod_GET,
							Query:  *id,
							Scope:  "global",
						})
					}

					if strings.HasPrefix(*id, "i-") {
						item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
							Type:   "ec2-instance",
							Method: sdp.RequestMethod_GET,
							Query:  *id,
							Scope:  scope,
						})
					}

					if strings.HasPrefix(*id, "arn:aws:lambda") {
						item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
							Type:   "lambda-function",
							Method: sdp.RequestMethod_GET,
							Query:  *id,
							Scope:  scope,
						})
					}

					if strings.HasPrefix(*id, "arn:aws:elasticloadbalancing") {
						item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
							Type:   "elasticloadbalancing-loadbalancer-v2",
							Method: sdp.RequestMethod_GET,
							Query:  *id,
							Scope:  scope,
						})
					}
				}
			}
		}
	}

	// Security groups
	for _, group := range lb.SecurityGroups {
		item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
			Type:   "ec2-securitygroup",
			Method: sdp.RequestMethod_GET,
			Query:  group,
			Scope:  scope,
		})
	}
	attributes, err := sources.ToAttributesCase(attrMap)

	if err != nil {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: fmt.Sprintf("error creating attributes: %v", err),
			Scope:       scope,
		}
	}

	item.Attributes = attributes

	return &item, nil
}
