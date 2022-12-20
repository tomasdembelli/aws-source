package ec2

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func TestEgressOnlyInternetGatewayInputMapperGet(t *testing.T) {
	input, err := EgressOnlyInternetGatewayInputMapperGet("foo", "bar")

	if err != nil {
		t.Error(err)
	}

	if len(input.EgressOnlyInternetGatewayIds) != 1 {
		t.Fatalf("expected 1 EgressOnlyInternetGateway ID, got %v", len(input.EgressOnlyInternetGatewayIds))
	}

	if input.EgressOnlyInternetGatewayIds[0] != "bar" {
		t.Errorf("expected EgressOnlyInternetGateway ID to be bar, got %v", input.EgressOnlyInternetGatewayIds[0])
	}
}

func TestEgressOnlyInternetGatewayInputMapperList(t *testing.T) {
	input, err := EgressOnlyInternetGatewayInputMapperList("foo")

	if err != nil {
		t.Error(err)
	}

	if len(input.Filters) != 0 || len(input.EgressOnlyInternetGatewayIds) != 0 {
		t.Errorf("non-empty input: %v", input)
	}
}

func TestEgressOnlyInternetGatewayOutputMapper(t *testing.T) {
	output := &ec2.DescribeEgressOnlyInternetGatewaysOutput{
		EgressOnlyInternetGateways: []types.EgressOnlyInternetGateway{
			{
				Attachments: []types.InternetGatewayAttachment{
					{
						State: types.AttachmentStatusAttached,
						VpcId: sources.PtrString("vpc-0d7892e00e573e701"),
					},
				},
				EgressOnlyInternetGatewayId: sources.PtrString("eigw-0ff50f360e066777a"),
			},
		},
	}

	items, err := EgressOnlyInternetGatewayOutputMapper("foo", output)

	if err != nil {
		t.Fatal(err)
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

	// It doesn't really make sense to test anything other than the linked items
	// since the attributes are converted automatically
	tests := sources.ItemRequestTests{
		{
			ExpectedType:   "ec2-vpc",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "vpc-0d7892e00e573e701",
			ExpectedScope:  "foo",
		},
	}

	tests.Execute(t, item)

}
