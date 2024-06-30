package networkmanager

import (
	"context"
	"log/slog"
	"testing"

	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/aws-source/sources/integration"
	"github.com/overmindtech/aws-source/sources/networkmanager"
	"github.com/overmindtech/sdp-go"
)

func TestIntegrationNetworkManager(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	networkmanagerClient, err := createNetworkManagerClient(ctx)
	if err != nil {
		t.Fatalf("Failed to create NetworkManager client: %v", err)
	}

	t.Run("Setup", func(t *testing.T) {
		if err := setup(ctx, logger, networkmanagerClient); err != nil {
			t.Fatalf("Failed to setup NetworkManager integration tests: %v", err)
		}
	})

	t.Run("Test Global Network", func(t *testing.T) {
		t.Logf("Running NetworkManager integration tests")
		TestNetworkManager(t)
	})

	t.Run("Teardown", func(t *testing.T) {
		if err := teardown(ctx, logger, networkmanagerClient); err != nil {
			t.Fatalf("Failed to teardown NetworkManager integration tests: %v", err)
		}
	})
}

func TestNetworkManager(t *testing.T) {
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

	globalScope := sources.FormatScope(awsCfg.AccountID, "")

	// List global networks
	sdpListGlobalNetworks, err := globalNetworkSource.List(context.Background(), globalScope, true)
	if err != nil {
		t.Fatalf("failed to list NetworkManager global networks: %v", err)
	}

	if len(sdpListGlobalNetworks) == 0 {
		t.Fatalf("no global networks found")
	}

	globalNetworkUniqueAttribute := sdpListGlobalNetworks[0].GetUniqueAttribute()

	globalNetworkID, err := integration.GetUniqueAttributeValue(
		globalNetworkUniqueAttribute,
		sdpListGlobalNetworks,
		integration.ResourceTags(integration.NetworkManager, globalNetworkSrc),
	)
	if err != nil {
		t.Fatalf("failed to get global network ID: %v", err)
	}

	// Get global network
	sdpGlobalNetwork, err := globalNetworkSource.Get(context.Background(), globalScope, globalNetworkID, true)
	if err != nil {
		t.Fatalf("failed to get NetworkManager global network: %v", err)
	}

	globalNetworkIDFromGet, err := integration.GetUniqueAttributeValue(
		globalNetworkUniqueAttribute,
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

	globalScope = sdpGlobalNetwork.GetScope()

	sdpSearchGlobalNetworks, err := globalNetworkSource.Search(context.Background(), globalScope, globalNetworkARN.(string), true)
	if err != nil {
		t.Fatalf("failed to search NetworkManager global networks: %v", err)
	}

	if len(sdpSearchGlobalNetworks) == 0 {
		t.Fatalf("no global networks found")
	}

	instanceIDFromSearch, err := integration.GetUniqueAttributeValue(
		globalNetworkUniqueAttribute,
		sdpSearchGlobalNetworks,
		integration.ResourceTags(integration.NetworkManager, globalNetworkSrc),
	)
	if err != nil {
		t.Fatalf("failed to get global network ID from search: %v", err)
	}

	if globalNetworkID != instanceIDFromSearch {
		t.Fatalf("expected global network ID %s, got %s", globalNetworkID, instanceIDFromSearch)
	}

	// Search sites by the global network ID that they are created on
	siteSource := networkmanager.NewSiteSource(networkManagerCli, awsCfg.AccountID)

	sdpSearchSites, err := siteSource.Search(ctx, globalScope, globalNetworkID, true)
	if err != nil {
		t.Fatalf("failed to search for site: %v", err)
	}

	if len(sdpSearchSites) == 0 {
		t.Fatalf("no sites found")
	}

	siteUniqueAttribute := sdpSearchSites[0].GetUniqueAttribute()

	// composite site id is in the format of {globalNetworkID}|{siteID}
	compositeSiteID, err := integration.GetUniqueAttributeValue(
		siteUniqueAttribute,
		sdpSearchSites,
		integration.ResourceTags(integration.NetworkManager, siteSrc),
	)
	if err != nil {
		t.Fatalf("failed to get site ID from search: %v", err)
	}

	// Get site: query format = globalNetworkID|siteID
	sdpGetSite, err := siteSource.Get(ctx, globalScope, compositeSiteID, true)
	if err != nil {
		t.Fatalf("failed to get site: %v", err)
	}

	siteIDFromGet, err := integration.GetUniqueAttributeValue(
		siteUniqueAttribute,
		[]*sdp.Item{sdpGetSite},
		integration.ResourceTags(integration.NetworkManager, siteSrc),
	)
	if err != nil {
		t.Fatalf("failed to get site ID from get: %v", err)
	}

	if compositeSiteID != siteIDFromGet {
		t.Fatalf("expected site ID %s, got %s", compositeSiteID, siteIDFromGet)
	}

	// Search links by the global network ID that they are created on
	linkSource := networkmanager.NewLinkSource(networkManagerCli, awsCfg.AccountID)

	sdpSearchLinks, err := linkSource.Search(ctx, globalScope, globalNetworkID, true)
	if err != nil {
		t.Fatalf("failed to search for link: %v", err)
	}

	if len(sdpSearchLinks) == 0 {
		t.Fatalf("no links found")
	}

	linkUniqueAttribute := sdpSearchLinks[0].GetUniqueAttribute()

	compositeLinkID, err := integration.GetUniqueAttributeValue(
		linkUniqueAttribute,
		sdpSearchLinks,
		integration.ResourceTags(integration.NetworkManager, linkSrc),
	)
	if err != nil {
		t.Fatalf("failed to get link ID from search: %v", err)
	}

	// Get link: query format = globalNetworkID|linkID
	sdpGetLink, err := linkSource.Get(ctx, globalScope, compositeLinkID, true)
	if err != nil {
		t.Fatalf("failed to get link: %v", err)
	}

	linkIDFromGet, err := integration.GetUniqueAttributeValue(
		linkUniqueAttribute,
		[]*sdp.Item{sdpGetLink},
		integration.ResourceTags(integration.NetworkManager, linkSrc),
	)

	if compositeLinkID != linkIDFromGet {
		t.Fatalf("expected link ID %s, got %s", compositeLinkID, linkIDFromGet)
	}

	// Search devices by the global network ID and site ID
	deviceSource := networkmanager.NewDeviceSource(networkManagerCli, awsCfg.AccountID)

	sdpSearchDevices, err := deviceSource.Search(ctx, globalScope, compositeSiteID, true)
	if err != nil {
		t.Fatalf("failed to search for device: %v", err)
	}

	if len(sdpSearchDevices) == 0 {
		t.Fatalf("no devices found")
	}

	deviceUniqueAttribute := sdpSearchDevices[0].GetUniqueAttribute()

	// composite device id is in the format of: {globalNetworkID}|{deviceID}
	compositeDeviceID, err := integration.GetUniqueAttributeValue(
		deviceUniqueAttribute,
		sdpSearchDevices,
		integration.ResourceTags(integration.NetworkManager, deviceSrc),
	)
	if err != nil {
		t.Fatalf("failed to get device ID from search: %v", err)
	}

	// Get device: query format = globalNetworkID|deviceID
	sdpGetDevice, err := deviceSource.Get(ctx, globalScope, compositeDeviceID, true)
	if err != nil {
		t.Fatalf("failed to get device: %v", err)
	}

	deviceIDFromGet, err := integration.GetUniqueAttributeValue(
		deviceUniqueAttribute,
		[]*sdp.Item{sdpGetDevice},
		integration.ResourceTags(integration.NetworkManager, deviceSrc),
	)

	if compositeDeviceID != deviceIDFromGet {
		t.Fatalf("expected device ID %s, got %s", compositeDeviceID, deviceIDFromGet)
	}

	// Search link associations by the global network ID, link ID, and device ID
	linkAssociationSource := networkmanager.NewLinkAssociationSource(networkManagerCli, awsCfg.AccountID)
}
