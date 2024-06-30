package networkmanager

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
)

func deleteGlobalNetwork(ctx context.Context, client *networkmanager.Client, globalNetworkID string) error {
	input := &networkmanager.DeleteGlobalNetworkInput{
		GlobalNetworkId: aws.String(globalNetworkID),
	}

	_, err := client.DeleteGlobalNetwork(ctx, input)
	return err
}

func deleteSite(ctx context.Context, client *networkmanager.Client, globalNetworkID, siteID *string) error {
	input := &networkmanager.DeleteSiteInput{
		GlobalNetworkId: globalNetworkID,
		SiteId:          siteID,
	}

	_, err := client.DeleteSite(ctx, input)
	return err
}

func deleteLink(ctx context.Context, client *networkmanager.Client, globalNetworkID, linkID *string) error {
	input := &networkmanager.DeleteLinkInput{
		GlobalNetworkId: globalNetworkID,
		LinkId:          linkID,
	}

	_, err := client.DeleteLink(ctx, input)
	return err
}

func deleteDevice(ctx context.Context, client *networkmanager.Client, globalNetworkID, deviceID *string) error {
	input := &networkmanager.DeleteDeviceInput{
		GlobalNetworkId: globalNetworkID,
		DeviceId:        deviceID,
	}

	_, err := client.DeleteDevice(ctx, input)
	return err
}
