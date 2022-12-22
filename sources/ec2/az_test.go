package ec2

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/overmindtech/aws-source/sources"
)

func TestAvailabilityZoneInputMapperGet(t *testing.T) {
	input, err := AvailabilityZoneInputMapperGet("foo", "az-name")

	if err != nil {
		t.Error(err)
	}

	if len(input.ZoneNames) != 1 {
		t.Fatalf("expected 1 zone names, got %v", len(input.ZoneNames))
	}

	if input.ZoneNames[0] != "az-name" {
		t.Errorf("expected zone name to be to be az-name, got %v", input.ZoneNames[0])
	}
}

func TestAvailabilityZoneInputMapperList(t *testing.T) {

	input, err := AvailabilityZoneInputMapperList("foo")

	if err != nil {
		t.Error(err)
	}

	if len(input.ZoneNames) != 0 {
		t.Fatalf("expected 0 zone names, got %v", len(input.ZoneNames))
	}
}

func TestAvailabilityZoneOutputMapper(t *testing.T) {
	output := ec2.DescribeAvailabilityZonesOutput{
		AvailabilityZones: []types.AvailabilityZone{
			{
				State:       types.AvailabilityZoneStateAvailable,
				OptInStatus: types.AvailabilityZoneOptInStatusOptInNotRequired,
				Messages: []types.AvailabilityZoneMessage{
					{
						Message: sources.PtrString("everything is fine..."),
					},
				},
				RegionName:         sources.PtrString("eu-west-2"),
				ZoneName:           sources.PtrString("eu-west-2a"),
				ZoneId:             sources.PtrString("euw2-az2"),
				GroupName:          sources.PtrString("eu-west-2"),
				NetworkBorderGroup: sources.PtrString("eu-west-2"),
				ZoneType:           sources.PtrString("availability-zone"),
			},
			{
				State:              types.AvailabilityZoneStateAvailable,
				OptInStatus:        types.AvailabilityZoneOptInStatusOptInNotRequired,
				Messages:           []types.AvailabilityZoneMessage{},
				RegionName:         sources.PtrString("eu-west-2"),
				ZoneName:           sources.PtrString("eu-west-2b"),
				ZoneId:             sources.PtrString("euw2-az3"),
				GroupName:          sources.PtrString("eu-west-2"),
				NetworkBorderGroup: sources.PtrString("eu-west-2"),
				ZoneType:           sources.PtrString("availability-zone"),
			},
		},
	}

	items, err := AvailabilityZoneOutputMapper("foo", &output)

	if err != nil {
		t.Error(err)
	}

	for _, item := range items {
		if err := item.Validate(); err != nil {
			t.Error(err)
		}
	}

	if len(items) != 2 {
		t.Errorf("expected 2 items, got %v", len(items))
	}

	firstItem := items[0]
	secondItem := items[1]

	if firstItem.UniqueAttributeValue() != *output.AvailabilityZones[0].ZoneName {
		t.Errorf("expected first item name to be %v, got %v", *output.AvailabilityZones[0].ZoneName, firstItem.UniqueAttributeValue())
	}

	if secondItem.UniqueAttributeValue() != *output.AvailabilityZones[1].ZoneName {
		t.Errorf("expected second item name to be %v, got %v", *output.AvailabilityZones[1].ZoneName, secondItem.UniqueAttributeValue())
	}
}