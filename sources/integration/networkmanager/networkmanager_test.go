package networkmanager

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/aws-source/sources/integration"
	"github.com/overmindtech/aws-source/sources/networkmanager"
	"github.com/overmindtech/sdp-go"
)

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
	if globalNetworkSource.Validate() != nil {
		t.Fatalf("failed to validate NetworkManager global network source: %v", err)
	}

	siteSource := networkmanager.NewSiteSource(networkManagerCli, awsCfg.AccountID)
	if siteSource.Validate() != nil {
		t.Fatalf("failed to validate NetworkManager site source: %v", err)
	}

	linkSource := networkmanager.NewLinkSource(networkManagerCli, awsCfg.AccountID)
	if linkSource.Validate() != nil {
		t.Fatalf("failed to validate NetworkManager link source: %v", err)
	}

	linkAssociationSource := networkmanager.NewLinkAssociationSource(networkManagerCli, awsCfg.AccountID)
	if linkAssociationSource.Validate() != nil {
		t.Fatalf("failed to validate NetworkManager link association source: %v", err)
	}

	globalScope := sources.FormatScope(awsCfg.AccountID, "")

	t.Run("Global Network", func(t *testing.T) {
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

		// Search global network by ARN
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

		globalNetworkIDFromSearch, err := integration.GetUniqueAttributeValue(
			globalNetworkUniqueAttribute,
			sdpSearchGlobalNetworks,
			integration.ResourceTags(integration.NetworkManager, globalNetworkSrc),
		)
		if err != nil {
			t.Fatalf("failed to get global network ID from search: %v", err)
		}

		if globalNetworkID != globalNetworkIDFromSearch {
			t.Fatalf("expected global network ID %s, got %s", globalNetworkID, globalNetworkIDFromSearch)
		}

		t.Run("Site", func(t *testing.T) {
			// Search sites by the global network ID that they are created on
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

			siteID := strings.Split(compositeSiteID, "|")[1]

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

			t.Run("Link", func(t *testing.T) {
				// Search links by the global network ID that they are created on
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

				linkID := strings.Split(compositeLinkID, "|")[1]

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

				t.Run("Device", func(t *testing.T) {
					// Search devices by the global network ID and site ID
					deviceSource := networkmanager.NewDeviceSource(networkManagerCli, awsCfg.AccountID)

					// Search device: query format = globalNetworkID|siteID
					query := fmt.Sprintf("%s|%s", globalNetworkID, siteID)
					sdpSearchDevices, err := deviceSource.Search(ctx, globalScope, query, true)
					if err != nil {
						t.Fatalf("failed to search for device: %v", err)
					}

					// TODO: Add search for other query: just globalNetworkID

					if len(sdpSearchDevices) == 0 {
						t.Fatalf("no devices found")
					}

					deviceUniqueAttribute := sdpSearchDevices[0].GetUniqueAttribute()

					// composite device id is in the format of: {globalNetworkID}|{deviceID}
					compositeDeviceOneID, err := integration.GetUniqueAttributeValue(
						deviceUniqueAttribute,
						sdpSearchDevices,
						integration.ResourceTags(integration.NetworkManager, deviceSrc, deviceOneName),
					)
					if err != nil {
						t.Fatalf("failed to get device ID from search: %v", err)
					}

					// Get device: query format = globalNetworkID|deviceID
					sdpGetDevice, err := deviceSource.Get(ctx, globalScope, compositeDeviceOneID, true)
					if err != nil {
						t.Fatalf("failed to get device: %v", err)
					}

					deviceOneIDFromGet, err := integration.GetUniqueAttributeValue(
						deviceUniqueAttribute,
						[]*sdp.Item{sdpGetDevice},
						integration.ResourceTags(integration.NetworkManager, deviceSrc, deviceOneName),
					)
					if err != nil {
						t.Fatalf("failed to get device ID from get: %v", err)
					}

					if compositeDeviceOneID != deviceOneIDFromGet {
						t.Fatalf("expected device ID %s, got %s", compositeDeviceOneID, deviceOneIDFromGet)
					}

					t.Run("Link Association", func(t *testing.T) {
						// Search link associations by the global network ID, link ID
						query := fmt.Sprintf("%s|link|%s", globalNetworkID, linkID)
						sdpSearchLinkAssociations, err := linkAssociationSource.Search(ctx, globalScope, query, true)
						if err != nil {
							t.Fatalf("failed to search for link association: %v", err)
						}

						// TODO: Add search for other 2 formats (just globalNetworkID, globalNetworkID|deviceID)

						if len(sdpSearchLinkAssociations) == 0 {
							t.Fatalf("no link associations found")
						}

						linkAssociationUniqueAttribute := sdpSearchLinkAssociations[0].GetUniqueAttribute()

						// composite link association id is in the format of: {globalNetworkID}|{linkID}|{deviceID}
						compositeLinkAssociationID, err := integration.GetUniqueAttributeValue(
							linkAssociationUniqueAttribute,
							sdpSearchLinkAssociations,
							nil, // we didn't use tags on associations
						)
						if err != nil {
							t.Fatalf("failed to get link association ID from search: %v", err)
						}

						// Get link association: query format = globalNetworkID|linkID|deviceID
						sdpGetLinkAssociation, err := linkAssociationSource.Get(ctx, globalScope, compositeLinkAssociationID, true)
						if err != nil {
							t.Fatalf("failed to get link association: %v", err)
						}

						linkAssociationIDFromGet, err := integration.GetUniqueAttributeValue(
							linkAssociationUniqueAttribute,
							[]*sdp.Item{sdpGetLinkAssociation},
							nil, // we didn't use tags on associations
						)

						if compositeLinkAssociationID != linkAssociationIDFromGet {
							t.Fatalf("expected link association ID %s, got %s", compositeLinkAssociationID, linkAssociationIDFromGet)
						}
					})
				})
			})
		})
	})
}
