package ec2

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"
)

func TestSecurityGroupInputMapperGet(t *testing.T) {
	input, err := securityGroupInputMapperGet("foo", "bar")

	if err != nil {
		t.Error(err)
	}

	if len(input.GroupIds) != 1 {
		t.Fatalf("expected 1 SecurityGroup ID, got %v", len(input.GroupIds))
	}

	if input.GroupIds[0] != "bar" {
		t.Errorf("expected SecurityGroup ID to be bar, got %v", input.GroupIds[0])
	}
}

func TestSecurityGroupInputMapperList(t *testing.T) {
	input, err := securityGroupInputMapperList("foo")

	if err != nil {
		t.Error(err)
	}

	if len(input.Filters) != 0 || len(input.GroupIds) != 0 {
		t.Errorf("non-empty input: %v", input)
	}
}

func TestSecurityGroupOutputMapper(t *testing.T) {
	output := &ec2.DescribeSecurityGroupsOutput{
		SecurityGroups: []types.SecurityGroup{
			{
				Description: adapters.PtrString("default VPC security group"),
				GroupName:   adapters.PtrString("default"),
				IpPermissions: []types.IpPermission{
					{
						IpProtocol:    adapters.PtrString("-1"),
						IpRanges:      []types.IpRange{},
						Ipv6Ranges:    []types.Ipv6Range{},
						PrefixListIds: []types.PrefixListId{},
						UserIdGroupPairs: []types.UserIdGroupPair{
							{
								GroupId: adapters.PtrString("sg-094e151c9fc5da181"),
								UserId:  adapters.PtrString("052392120704"),
							},
						},
					},
				},
				OwnerId: adapters.PtrString("052392120703"),
				GroupId: adapters.PtrString("sg-094e151c9fc5da181"),
				IpPermissionsEgress: []types.IpPermission{
					{
						IpProtocol: adapters.PtrString("-1"),
						IpRanges: []types.IpRange{
							{
								CidrIp: adapters.PtrString("0.0.0.0/0"),
							},
						},
						Ipv6Ranges:       []types.Ipv6Range{},
						PrefixListIds:    []types.PrefixListId{},
						UserIdGroupPairs: []types.UserIdGroupPair{},
					},
				},
				VpcId: adapters.PtrString("vpc-0d7892e00e573e701"),
			},
		},
	}

	items, err := securityGroupOutputMapper(context.Background(), nil, "052392120703.eu-west-2", nil, output)

	if err != nil {
		t.Fatal(err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %v", len(items))
	}

	item := items[0]

	// It doesn't really make sense to test anything other than the linked items
	// since the attributes are converted automatically
	tests := adapters.QueryTests{
		{
			ExpectedType:   "ec2-vpc",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "vpc-0d7892e00e573e701",
			ExpectedScope:  item.GetScope(),
		},
		{
			ExpectedType:   "ec2-security-group",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "sg-094e151c9fc5da181",
			ExpectedScope:  "052392120704.eu-west-2",
		},
	}

	tests.Execute(t, item)

}

func TestNewSecurityGroupAdapter(t *testing.T) {
	client, account, region := GetAutoConfig(t)

	adapter := NewSecurityGroupAdapter(client, account, region)

	test := adapters.E2ETest{
		Adapter: adapter,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
