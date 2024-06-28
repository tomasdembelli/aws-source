package networkmanager

import (
	"context"
	"testing"

	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/aws-source/sources/integration"
	"github.com/overmindtech/aws-source/sources/networkmanager"
	"github.com/overmindtech/sdp-go"
)

func TestGlobalNetworkSource(t *testing.T) {
	ctx := context.Background()

	t.Logf("Running NetworkManager integration tests")

	networkManagerCli, err := createNetworkManagerClient(ctx)
	if err != nil {
		t.Fatalf("failed to create NetworkManager client: %v", err)
	}

	awsCfg, err := integration.AWSSettings(ctx)
	if err != nil {
		t.Fatalf("failed to get AWS settings: %v", err)
	}

	globalNetworkSource := networkmanager.NewGlobalNetworkSource(networkManagerCli, awsCfg.AccountID)

	err = globalNetworkSource.Validate()
	if err != nil {
		t.Fatalf("failed to validate NetworkManager global network source: %v", err)
	}

	scope := sources.FormatScope(awsCfg.AccountID, "")

	// List global networks
	sdpListGlobalNetworks, err := globalNetworkSource.List(context.Background(), scope, true)
	if err != nil {
		t.Fatalf("failed to list NetworkManager global networks: %v", err)
	}

	if len(sdpListGlobalNetworks) == 0 {
		t.Fatalf("no global networks found")
	}

	uniqueAttribute := sdpListGlobalNetworks[0].GetUniqueAttribute()

	globalNetworkID, err := integration.GetUniqueAttributeValue(
		uniqueAttribute,
		sdpListGlobalNetworks,
		integration.ResourceTags(integration.NetworkManager, globalNetworkSrc),
	)
	if err != nil {
		t.Fatalf("failed to get global network ID: %v", err)
	}

	// Get global network
	sdpGlobalNetwork, err := globalNetworkSource.Get(context.Background(), scope, globalNetworkID, true)
	if err != nil {
		t.Fatalf("failed to get NetworkManager global network: %v", err)
	}

	globalNetworkIDFromGet, err := integration.GetUniqueAttributeValue(
		uniqueAttribute,
		[]*sdp.Item{sdpGlobalNetwork},
		integration.ResourceTags(integration.NetworkManager, globalNetworkSrc),
	)
	if err != nil {
		t.Fatalf("failed to get global network ID from get: %v", err)
	}

	if globalNetworkID != globalNetworkIDFromGet {
		t.Fatalf("expected global network ID %s, got %s", globalNetworkID, globalNetworkIDFromGet)
	}

	// Search global network
	globalNetworkARN, err := sdpGlobalNetwork.GetAttributes().Get("globalNetworkArn")
	if err != nil {
		t.Fatalf("failed to get global network ARN: %v", err)
	}

	scope = sdpGlobalNetwork.GetScope()

	sdpSearchGlobalNetworks, err := globalNetworkSource.Search(context.Background(), scope, globalNetworkARN.(string), true)
	if err != nil {
		t.Fatalf("failed to search NetworkManager global networks: %v", err)
	}

	instanceIDFromSearch, err := integration.GetUniqueAttributeValue(
		uniqueAttribute,
		sdpSearchGlobalNetworks,
		integration.ResourceTags(integration.NetworkManager, globalNetworkSrc),
	)
	if err != nil {
		t.Fatalf("failed to get global network ID from search: %v", err)
	}

	if globalNetworkID != instanceIDFromSearch {
		t.Fatalf("expected global network ID %s, got %s", globalNetworkID, instanceIDFromSearch)
	}
}
