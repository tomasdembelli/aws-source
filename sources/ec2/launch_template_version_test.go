package ec2

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func TestLaunchTemplateVersionInputMapperGet(t *testing.T) {
	input, err := launchTemplateVersionInputMapperGet("foo", "bar.10")

	if err != nil {
		t.Error(err)
	}

	if len(input.Versions) != 1 {
		t.Fatalf("expected 1 version, got %v", len(input.Versions))
	}

	if input.Versions[0] != "10" {
		t.Fatalf("expected version to be 10, got %v", input.Versions[0])
	}

	if *input.LaunchTemplateId != "bar" {
		t.Errorf("expected LaunchTemplateId to be bar, got %v", *input.LaunchTemplateId)
	}
}

func TestLaunchTemplateVersionInputMapperList(t *testing.T) {
	input, err := launchTemplateVersionInputMapperList("foo")

	if err != nil {
		t.Error(err)
	}

	if len(input.Versions) != 2 {
		t.Errorf("expected 2 inputs, got %v: %v", len(input.Versions), input)
	}
}

func TestLaunchTemplateVersionOutputMapper(t *testing.T) {
	output := &ec2.DescribeLaunchTemplateVersionsOutput{
		LaunchTemplateVersions: []types.LaunchTemplateVersion{
			{
				LaunchTemplateId:   sources.PtrString("lt-015547202038ae102"),
				LaunchTemplateName: sources.PtrString("test"),
				VersionNumber:      sources.PtrInt64(1),
				CreateTime:         sources.PtrTime(time.Now()),
				CreatedBy:          sources.PtrString("arn:aws:sts::052392120703:assumed-role/AWSReservedSSO_AWSAdministratorAccess_c1c3c9c54821c68a/dylan@overmind.tech"),
				DefaultVersion:     sources.PtrBool(true),
				LaunchTemplateData: &types.ResponseLaunchTemplateData{
					NetworkInterfaces: []types.LaunchTemplateInstanceNetworkInterfaceSpecification{
						{
							Ipv6Addresses: []types.InstanceIpv6Address{
								{
									Ipv6Address: sources.PtrString("ipv6"),
								},
							},
							NetworkInterfaceId: sources.PtrString("networkInterface"),
							PrivateIpAddresses: []types.PrivateIpAddressSpecification{
								{
									Primary:          sources.PtrBool(true),
									PrivateIpAddress: sources.PtrString("ip"),
								},
							},
							SubnetId:    sources.PtrString("subnet"),
							DeviceIndex: sources.PtrInt32(0),
							Groups: []string{
								"sg-094e151c9fc5da181",
							},
						},
					},
					ImageId:      sources.PtrString("ami-084e8c05825742534"),
					InstanceType: types.InstanceTypeT1Micro,
					KeyName:      sources.PtrString("dylan.ratcliffe"),
					BlockDeviceMappings: []types.LaunchTemplateBlockDeviceMapping{
						{
							Ebs: &types.LaunchTemplateEbsBlockDevice{
								SnapshotId: sources.PtrString("snap"),
							},
						},
					},
					CapacityReservationSpecification: &types.LaunchTemplateCapacityReservationSpecificationResponse{
						CapacityReservationPreference: types.CapacityReservationPreferenceNone,
						CapacityReservationTarget: &types.CapacityReservationTargetResponse{
							CapacityReservationId: sources.PtrString("cap"),
						},
					},
					CpuOptions:                   &types.LaunchTemplateCpuOptions{},
					CreditSpecification:          &types.CreditSpecification{},
					ElasticGpuSpecifications:     []types.ElasticGpuSpecificationResponse{},
					EnclaveOptions:               &types.LaunchTemplateEnclaveOptions{},
					ElasticInferenceAccelerators: []types.LaunchTemplateElasticInferenceAcceleratorResponse{},
					Placement: &types.LaunchTemplatePlacement{
						AvailabilityZone: sources.PtrString("foo"),
						GroupId:          sources.PtrString("placement"),
						HostId:           sources.PtrString("host"),
					},
					SecurityGroupIds: []string{
						"secGroup",
					},
				},
			},
		},
	}

	items, err := launchTemplateVersionOutputMapper(context.Background(), nil, "foo", nil, output)

	if err != nil {
		t.Fatal(err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %v", len(items))
	}

	item := items[0]

	// It doesn't really make sense to test anything other than the linked items
	// since the attributes are converted automatically
	tests := sources.QueryTests{
		{
			ExpectedType:   "ip",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "ipv6",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "ec2-network-interface",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "networkInterface",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ip",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "ip",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "ec2-subnet",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "subnet",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-security-group",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "sg-094e151c9fc5da181",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-image",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "ami-084e8c05825742534",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-key-pair",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "dylan.ratcliffe",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-snapshot",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "snap",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-capacity-reservation",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "cap",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-placement-group",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "placement",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-host",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "host",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "ec2-security-group",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "secGroup",
			ExpectedScope:  "foo",
		},
	}

	tests.Execute(t, item)

}

func TestNewLaunchTemplateVersionSource(t *testing.T) {
	client, account, region := GetAutoConfig(t)

	source := NewLaunchTemplateVersionSource(client, account, region)

	test := sources.E2ETest{
		Adapter:           source,
		Timeout:           10 * time.Second,
		SkipNotFoundCheck: true,
	}

	test.Run(t)
}
