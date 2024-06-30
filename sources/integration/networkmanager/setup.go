package networkmanager

import (
	"context"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/overmindtech/aws-source/sources/integration"
)

const (
	globalNetworkSrc   = "global-network"
	siteSrc            = "site"
	linkSrc            = "link"
	deviceSrc          = "device"
	linkAssociationSrc = "link-association"
)

func setup(ctx context.Context, logger *slog.Logger, networkmanagerClient *networkmanager.Client) error {
	testID := integration.TestID()
	// Create a global network
	globalNetworkID, err := createGlobalNetwork(ctx, logger, networkmanagerClient, testID)
	if err != nil {
		return err
	}

	// Create a site in the global network
	siteID, err := createSite(ctx, logger, networkmanagerClient, testID, globalNetworkID)
	if err != nil {
		return err
	}

	// Create a link in the global network for the site
	_, err = createLink(ctx, logger, networkmanagerClient, testID, globalNetworkID, siteID)
	if err != nil {
		return err
	}

	// Create a device in the global network for the site
	_, err = createDevice(ctx, logger, networkmanagerClient, testID, globalNetworkID, siteID)
	if err != nil {
		return err
	}

	return nil
}

func createNetworkManagerClient(ctx context.Context) (*networkmanager.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	client := networkmanager.NewFromConfig(cfg)

	return client, nil
}
