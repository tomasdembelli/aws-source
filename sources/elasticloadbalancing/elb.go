package elasticloadbalancing

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	elb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	"github.com/overmindtech/sdp-go"
)

type ELBSource struct {
	// Config AWS Config including region and credentials
	Config aws.Config

	// AccountID The id of the account that is being used. This is used by
	// sources as the first element in the context
	AccountID string

	// client The AWS client to use when making requests
	client        *elb.Client
	clientCreated bool
	clientMutex   sync.Mutex
}

func (s *ELBSource) Client() *elb.Client {
	s.clientMutex.Lock()
	defer s.clientMutex.Unlock()

	// If the client already exists then return it
	if s.clientCreated {
		return s.client
	}

	// Otherwise create a new client from the config
	s.client = elb.NewFromConfig(s.Config)
	s.clientCreated = true

	return s.client
}

// Type The type of items that this source is capable of finding
func (s *ELBSource) Type() string {
	return "elasticloadbalancer"
}

// Descriptive name for the source, used in logging and metadata
func (s *ELBSource) Name() string {
	return "elasticloadbalancing-aws-source"
}

// List of contexts that this source is capable of find items for. This will be
// in the format {accountID}.{region}
func (s *ELBSource) Contexts() []string {
	return []string{
		fmt.Sprintf("%v.%v", s.AccountID, s.Config.Region),
	}
}

// Get Get a single item with a given context and query. The item returned
// should have a UniqueAttributeValue that matches the `query` parameter. The
// ctx parameter contains a golang context object which should be used to allow
// this source to timeout or be cancelled when executing potentially
// long-running actions
func (s *ELBSource) Get(ctx context.Context, itemContext string, query string) (*sdp.Item, error) {
	lbs, err := s.Client().DescribeLoadBalancers(
		ctx,
		&elb.DescribeLoadBalancersInput{
			LoadBalancerNames: []string{
				query,
			},
		},
	)

	if err != nil {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: err.Error(),
			Context:     itemContext,
		}
	}

	switch len(lbs.LoadBalancerDescriptions) {
	case 0:
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOTFOUND,
			ErrorString: "elasticloadbalancer not found",
			Context:     itemContext,
		}
	case 1:
		return mapToItem(lbs.LoadBalancerDescriptions[0], itemContext)
	default:
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: fmt.Sprintf("more than 1 elasticloadbalancer found, found: %v", len(lbs.LoadBalancerDescriptions)),
			Context:     itemContext,
		}
	}
}

// Find Finds all items in a given context
func (s *ELBSource) Find(ctx context.Context, itemContext string) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	lbs, err := s.Client().DescribeLoadBalancers(
		ctx,
		&elb.DescribeLoadBalancersInput{},
	)

	if err != nil {
		return items, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: err.Error(),
			Context:     itemContext,
		}
	}

	for _, lb := range lbs.LoadBalancerDescriptions {
		item, err := mapToItem(lb, itemContext)

		if err == nil {
			items = append(items, item)
		}
	}

	return items, nil
}

// Weight Returns the priority weighting of items returned by this source.
// This is used to resolve conflicts where two sources of the same type
// return an item for a GET request. In this instance only one item can be
// sen on, so the one with the higher weight value will win.
func (s *ELBSource) Weight() int {
	return 100
}

// mapToItem Maps a load balancer to an item
func mapToItem(lb types.LoadBalancerDescription, itemContext string) (*sdp.Item, error) {
	attrMap := make(map[string]interface{})

	if lb.LoadBalancerName == nil || *lb.LoadBalancerName == "" {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: "elasticloadbalancer was returned with an empty name",
			Context:     itemContext,
		}
	}

	item := sdp.Item{
		Type:            "elasticloadbalancer",
		UniqueAttribute: "name",
		Context:         itemContext,
	}

	attrMap["name"] = lb.LoadBalancerName
	attrMap["availabilityZones"] = lb.AvailabilityZones
	attrMap["backendServerDescriptions"] = lb.BackendServerDescriptions
	attrMap["instances"] = lb.Instances
	attrMap["listenerDescriptions"] = lb.ListenerDescriptions
	attrMap["subnets"] = lb.Subnets
	attrMap["securityGroups"] = lb.SecurityGroups

	if lb.CanonicalHostedZoneName != nil {
		attrMap["canonicalHostedZoneName"] = lb.CanonicalHostedZoneName
	}
	if lb.CanonicalHostedZoneNameID != nil {
		attrMap["canonicalHostedZoneNameID"] = lb.CanonicalHostedZoneNameID

		item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
			Type:    "hostedzone",
			Method:  sdp.RequestMethod_GET,
			Query:   *lb.CanonicalHostedZoneNameID,
			Context: itemContext,
		})
	}
	if lb.CreatedTime != nil {
		attrMap["createdTime"] = lb.CreatedTime.String()
	}
	if lb.DNSName != nil {
		attrMap["DNSName"] = lb.DNSName

		item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
			Type:    "dns",
			Method:  sdp.RequestMethod_GET,
			Query:   *lb.DNSName,
			Context: "global",
		})
	}
	if lb.HealthCheck != nil {
		attrMap["healthCheck"] = lb.HealthCheck
	}
	if lb.Policies != nil {
		attrMap["policies"] = lb.Policies
	}
	if lb.Scheme != nil {
		attrMap["scheme"] = lb.Scheme
	}
	if lb.SourceSecurityGroup != nil {
		attrMap["sourceSecurityGroup"] = lb.SourceSecurityGroup
	}
	if lb.VPCId != nil {
		attrMap["VPCId"] = lb.VPCId
	}

	attributes, err := sdp.ToAttributes(attrMap)

	if err != nil {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: fmt.Sprintf("error creating attributes: %v", err),
			Context:     itemContext,
		}
	}

	item.Attributes = attributes

	return &item, nil
}
