package networkmanager

import (
	"context"
	"errors"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/overmindtech/aws-source/sources/integration"
)

func teardown(ctx context.Context, logger *slog.Logger, client *networkmanager.Client) error {
	globalNetworkID, err := findGlobalNetworkIDByTags(ctx, client, resourceTags(globalNetworkSrc, integration.TestID()))
	if err != nil {
		nf := integration.NewNotFoundError(globalNetworkSrc)
		if errors.As(err, &nf) {
			logger.WarnContext(ctx, "Global network not found")
			return nil
		} else {
			return err
		}
	}

	siteID, err := findSiteIDByTags(ctx, client, globalNetworkID, resourceTags(siteSrc, integration.TestID()))
	if err != nil {
		nf := integration.NewNotFoundError(siteSrc)
		if errors.As(err, &nf) {
			logger.WarnContext(ctx, "Site not found")
			return nil
		} else {
			return err
		}
	}

	linkID, err := findLinkIDByTags(ctx, client, globalNetworkID, siteID, resourceTags(linkSrc, integration.TestID()))
	if err != nil {
		nf := integration.NewNotFoundError(linkSrc)
		if errors.As(err, &nf) {
			logger.WarnContext(ctx, "Link not found")
			return nil
		} else {
			return err
		}
	}

	deviceID, err := findDeviceIDByTags(ctx, client, globalNetworkID, siteID, resourceTags(deviceSrc, integration.TestID()))
	if err != nil {
		nf := integration.NewNotFoundError(deviceSrc)
		if errors.As(err, &nf) {
			logger.WarnContext(ctx, "Device not found")
			return nil
		} else {
			return err
		}
	}

	err = deleteDevice(ctx, client, globalNetworkID, deviceID)
	if err != nil {
		return err
	}

	err = deleteLink(ctx, client, globalNetworkID, linkID)
	if err != nil {
		return err
	}

	err = deleteSite(ctx, client, globalNetworkID, siteID)
	if err != nil {
		return err
	}

	err = deleteGlobalNetwork(ctx, client, *globalNetworkID)
	if err != nil {
		return err
	}

	return nil
}
