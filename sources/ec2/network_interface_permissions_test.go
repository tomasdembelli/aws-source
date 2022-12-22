package ec2

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func TestNetworkInterfacePermissionInputMapperGet(t *testing.T) {
	input, err := NetworkInterfacePermissionInputMapperGet("foo", "bar")

	if err != nil {
		t.Error(err)
	}

	if len(input.NetworkInterfacePermissionIds) != 1 {
		t.Fatalf("expected 1 NetworkInterfacePermission ID, got %v", len(input.NetworkInterfacePermissionIds))
	}

	if input.NetworkInterfacePermissionIds[0] != "bar" {
		t.Errorf("expected NetworkInterfacePermission ID to be bar, got %v", input.NetworkInterfacePermissionIds[0])
	}
}

func TestNetworkInterfacePermissionInputMapperList(t *testing.T) {
	input, err := NetworkInterfacePermissionInputMapperList("foo")

	if err != nil {
		t.Error(err)
	}

	if len(input.Filters) != 0 || len(input.NetworkInterfacePermissionIds) != 0 {
		t.Errorf("non-empty input: %v", input)
	}
}

func TestNetworkInterfacePermissionOutputMapper(t *testing.T) {
	output := &ec2.DescribeNetworkInterfacePermissionsOutput{
		NetworkInterfacePermissions: []types.NetworkInterfacePermission{
			{
				NetworkInterfacePermissionId: sources.PtrString("eni-perm-0b6211455242c105e"),
				NetworkInterfaceId:           sources.PtrString("eni-07f8f3d404036c833"),
				AwsService:                   sources.PtrString("routing.hyperplane.eu-west-2.amazonaws.com"),
				Permission:                   types.InterfacePermissionTypeInstanceAttach,
				PermissionState: &types.NetworkInterfacePermissionState{
					State: types.NetworkInterfacePermissionStateCodeGranted,
				},
			},
		},
	}

	items, err := NetworkInterfacePermissionOutputMapper("foo", output)

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
			ExpectedType:   "ec2-network-interface",
			ExpectedMethod: sdp.RequestMethod_GET,
			ExpectedQuery:  "eni-07f8f3d404036c833",
			ExpectedScope:  "foo",
		},
	}

	tests.Execute(t, item)

}