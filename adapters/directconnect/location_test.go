package directconnect

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	"github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/overmindtech/aws-source/adapters"
)

func TestLocationOutputMapper(t *testing.T) {
	output := &directconnect.DescribeLocationsOutput{
		Locations: []types.Location{
			{
				AvailableMacSecPortSpeeds: []string{"1 Gbps", "10 Gbps"},
				AvailablePortSpeeds:       []string{"50 Mbps", "100 Mbps", "1 Gbps", "10 Gbps"},
				AvailableProviders:        []string{"ProviderA", "ProviderB", "ProviderC"},
				LocationName:              adapters.PtrString("NAP do Brasil, Barueri, Sao Paulo"),
				LocationCode:              adapters.PtrString("TNDB"),
				Region:                    adapters.PtrString("us-east-1"),
			},
		},
	}

	items, err := locationOutputMapper(context.Background(), nil, "foo", nil, output)
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
}

func TestNewLocationAdapter(t *testing.T) {
	client, account, region := GetAutoConfig(t)

	adapter := NewLocationAdapter(client, account, region)

	test := adapters.E2ETest{
		Adapter: adapter,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
