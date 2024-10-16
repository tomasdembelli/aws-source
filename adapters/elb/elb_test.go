package elb

import (
	"context"
	"testing"
	"time"

	elb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"
)

type mockElbClient struct{}

func (m mockElbClient) DescribeTags(ctx context.Context, params *elb.DescribeTagsInput, optFns ...func(*elb.Options)) (*elb.DescribeTagsOutput, error) {
	return &elb.DescribeTagsOutput{
		TagDescriptions: []types.TagDescription{
			{
				LoadBalancerName: adapters.PtrString("a8c3c8851f0df43fda89797c8e941a91"),
				Tags: []types.Tag{
					{
						Key:   adapters.PtrString("foo"),
						Value: adapters.PtrString("bar"),
					},
				},
			},
		},
	}, nil
}

func (m mockElbClient) DescribeLoadBalancers(ctx context.Context, params *elb.DescribeLoadBalancersInput, optFns ...func(*elb.Options)) (*elb.DescribeLoadBalancersOutput, error) {
	return nil, nil
}

func TestLoadBalancerOutputMapper(t *testing.T) {
	output := &elb.DescribeLoadBalancersOutput{
		LoadBalancerDescriptions: []types.LoadBalancerDescription{
			{
				LoadBalancerName:          adapters.PtrString("a8c3c8851f0df43fda89797c8e941a91"),
				DNSName:                   adapters.PtrString("a8c3c8851f0df43fda89797c8e941a91-182843316.eu-west-2.elb.amazonaws.com"), // link
				CanonicalHostedZoneName:   adapters.PtrString("a8c3c8851f0df43fda89797c8e941a91-182843316.eu-west-2.elb.amazonaws.com"), // link
				CanonicalHostedZoneNameID: adapters.PtrString("ZHURV8PSTC4K8"),                                                          // link
				ListenerDescriptions: []types.ListenerDescription{
					{
						Listener: &types.Listener{
							Protocol:         adapters.PtrString("TCP"),
							LoadBalancerPort: 7687,
							InstanceProtocol: adapters.PtrString("TCP"),
							InstancePort:     adapters.PtrInt32(30133),
						},
						PolicyNames: []string{},
					},
					{
						Listener: &types.Listener{
							Protocol:         adapters.PtrString("TCP"),
							LoadBalancerPort: 7473,
							InstanceProtocol: adapters.PtrString("TCP"),
							InstancePort:     adapters.PtrInt32(31459),
						},
						PolicyNames: []string{},
					},
					{
						Listener: &types.Listener{
							Protocol:         adapters.PtrString("TCP"),
							LoadBalancerPort: 7474,
							InstanceProtocol: adapters.PtrString("TCP"),
							InstancePort:     adapters.PtrInt32(30761),
						},
						PolicyNames: []string{},
					},
				},
				Policies: &types.Policies{
					AppCookieStickinessPolicies: []types.AppCookieStickinessPolicy{
						{
							CookieName: adapters.PtrString("foo"),
							PolicyName: adapters.PtrString("policy"),
						},
					},
					LBCookieStickinessPolicies: []types.LBCookieStickinessPolicy{
						{
							CookieExpirationPeriod: adapters.PtrInt64(10),
							PolicyName:             adapters.PtrString("name"),
						},
					},
					OtherPolicies: []string{},
				},
				BackendServerDescriptions: []types.BackendServerDescription{
					{
						InstancePort: adapters.PtrInt32(443),
						PolicyNames:  []string{},
					},
				},
				AvailabilityZones: []string{ // link
					"euwest-2b",
					"euwest-2a",
					"euwest-2c",
				},
				Subnets: []string{ // link
					"subnet0960234bbc4edca03",
					"subnet09d5f6fa75b0b4569",
					"subnet0e234bef35fc4a9e1",
				},
				VPCId: adapters.PtrString("vpc-0c72199250cd479ea"), // link
				Instances: []types.Instance{
					{
						InstanceId: adapters.PtrString("i-0337802d908b4a81e"), // link *2 to ec2-instance and health
					},
				},
				HealthCheck: &types.HealthCheck{
					Target:             adapters.PtrString("HTTP:31151/healthz"),
					Interval:           adapters.PtrInt32(10),
					Timeout:            adapters.PtrInt32(5),
					UnhealthyThreshold: adapters.PtrInt32(6),
					HealthyThreshold:   adapters.PtrInt32(2),
				},
				SourceSecurityGroup: &types.SourceSecurityGroup{
					OwnerAlias: adapters.PtrString("944651592624"),
					GroupName:  adapters.PtrString("k8s-elb-a8c3c8851f0df43fda89797c8e941a91"), // link
				},
				SecurityGroups: []string{
					"sg097e3cfdfc6d53b77", // link
				},
				CreatedTime: adapters.PtrTime(time.Now()),
				Scheme:      adapters.PtrString("internet-facing"),
			},
		},
	}

	items, err := loadBalancerOutputMapper(context.Background(), mockElbClient{}, "foo", nil, output)

	if err != nil {
		t.Error(err)
	}

	for _, item := range items {
		if err := item.Validate(); err != nil {
			t.Error(err)
		}
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %v", len(items))
	}

	item := items[0]

	if item.GetTags()["foo"] != "bar" {
		t.Errorf("expected tag foo to be bar, got %v", item.GetTags()["foo"])
	}

	// It doesn't really make sense to test anything other than the linked items
	// since the attributes are converted automatically
	tests := adapters.QueryTests{
		{
			ExpectedType:   "dns",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "a8c3c8851f0df43fda89797c8e941a91-182843316.eu-west-2.elb.amazonaws.com",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "dns",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "a8c3c8851f0df43fda89797c8e941a91-182843316.eu-west-2.elb.amazonaws.com",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "route53-hosted-zone",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "ZHURV8PSTC4K8",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-subnet",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "subnet0960234bbc4edca03",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-subnet",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "subnet09d5f6fa75b0b4569",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-subnet",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "subnet0e234bef35fc4a9e1",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-vpc",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "vpc-0c72199250cd479ea",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-instance",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "i-0337802d908b4a81e",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "elb-instance-health",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "a8c3c8851f0df43fda89797c8e941a91/i-0337802d908b4a81e",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-security-group",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "k8s-elb-a8c3c8851f0df43fda89797c8e941a91",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-security-group",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "sg097e3cfdfc6d53b77",
			ExpectedScope:  "foo",
		},
	}

	tests.Execute(t, item)
}
