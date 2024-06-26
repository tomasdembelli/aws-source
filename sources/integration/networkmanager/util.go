package networkmanager

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/overmindtech/aws-source/sources/integration"
)

func resourceTags(resourceName, testID string, additionalAttr ...string) []types.Tag {
	return []types.Tag{
		{
			Key:   aws.String(integration.TagTestKey),
			Value: aws.String(integration.TagTestValue),
		},
		{
			Key:   aws.String(integration.TagTestTypeKey),
			Value: aws.String(integration.TestName(integration.NetworkManager)),
		},
		{
			Key:   aws.String(integration.TagTestIDKey),
			Value: aws.String(testID),
		},
		{
			Key:   aws.String(integration.TagResourceIDKey),
			Value: aws.String(integration.ResourceName(integration.NetworkManager, resourceName, additionalAttr...)),
		},
	}
}

func findGlobalNetworkIDByTags(ctx context.Context, client *networkmanager.Client, requiredTags []types.Tag) (*string, error) {
	result, err := client.DescribeGlobalNetworks(ctx, &networkmanager.DescribeGlobalNetworksInput{})
	if err != nil {
		return nil, err
	}

	for _, globalNetwork := range result.GlobalNetworks {
		if hasTags(globalNetwork.Tags, requiredTags) {
			return globalNetwork.GlobalNetworkId, nil
		}
	}

	return nil, integration.NewNotFoundError(integration.ResourceName(integration.NetworkManager, globalNetworkSrc))
}

func deleteGlobalNetwork(ctx context.Context, client *networkmanager.Client, globalNetworkID string) error {
	input := &networkmanager.DeleteGlobalNetworkInput{
		GlobalNetworkId: aws.String(globalNetworkID),
	}

	_, err := client.DeleteGlobalNetwork(ctx, input)
	return err
}

func hasTags(tags []types.Tag, requiredTags []types.Tag) bool {
	rT := make(map[string]string)
	for _, t := range requiredTags {
		rT[*t.Key] = *t.Value
	}

	oT := make(map[string]string)
	for _, t := range tags {
		oT[*t.Key] = *t.Value
	}

	for k, v := range rT {
		if oT[k] != v {
			return false
		}
	}

	return true
}

func findSiteIDByTags(ctx context.Context, client *networkmanager.Client, globalNetworkID *string, requiredTags []types.Tag) (*string, error) {
	result, err := client.GetSites(ctx, &networkmanager.GetSitesInput{
		GlobalNetworkId: globalNetworkID,
	})
	if err != nil {
		return nil, err
	}

	for _, site := range result.Sites {
		if hasTags(site.Tags, requiredTags) {
			return site.SiteId, nil
		}
	}

	return nil, integration.NewNotFoundError(integration.ResourceName(integration.NetworkManager, siteSrc))
}

func deleteSite(ctx context.Context, client *networkmanager.Client, globalNetworkID, siteID *string) error {
	input := &networkmanager.DeleteSiteInput{
		GlobalNetworkId: globalNetworkID,
		SiteId:          siteID,
	}

	_, err := client.DeleteSite(ctx, input)
	return err
}
