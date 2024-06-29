package networkmanager

import (
	"context"
	"errors"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/service/networkmanager/types"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/overmindtech/aws-source/sources/integration"
)

const (
	globalNetworkSrc = "global-network"
	siteSrc          = "site"
	linkSrc          = "link"
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

func createGlobalNetwork(ctx context.Context, logger *slog.Logger, client *networkmanager.Client, testID string) (*string, error) {
	tags := resourceTags(globalNetworkSrc, testID)

	input := &networkmanager.CreateGlobalNetworkInput{
		Description: aws.String("Integration test global network"),
		Tags:        tags,
	}

	id, err := findGlobalNetworkIDByTags(ctx, client, tags)
	if err != nil {
		if errors.As(err, new(integration.NotFoundError)) {
			logger.InfoContext(ctx, "Creating global network")
		} else {
			return nil, err
		}
	}

	if id != nil {
		logger.InfoContext(ctx, "Global network already exists")
		return id, nil
	}

	response, err := client.CreateGlobalNetwork(context.Background(), input)
	if err != nil {
		return nil, err
	}

	return response.GlobalNetwork.GlobalNetworkId, nil
}

func createSite(ctx context.Context, logger *slog.Logger, client *networkmanager.Client, testID string, globalNetworkID *string) (*string, error) {
	tags := resourceTags(siteSrc, testID)

	input := &networkmanager.CreateSiteInput{
		GlobalNetworkId: globalNetworkID,
		Description:     aws.String("Integration test site"),
		Tags:            tags,
	}

	id, err := findSiteIDByTags(ctx, client, globalNetworkID, tags)
	if err != nil {
		if errors.As(err, new(integration.NotFoundError)) {
			logger.InfoContext(ctx, "Creating site")
		} else {
			return nil, err
		}
	}

	if id != nil {
		logger.InfoContext(ctx, "Site already exists")
		return id, nil
	}

	response, err := client.CreateSite(ctx, input)
	if err != nil {
		return nil, err
	}

	return response.Site.SiteId, nil
}

func createLink(ctx context.Context, logger *slog.Logger, client *networkmanager.Client, testID string, globalNetworkID, siteID *string) (*string, error) {
	tags := resourceTags(linkSrc, testID)

	input := &networkmanager.CreateLinkInput{
		GlobalNetworkId: globalNetworkID,
		SiteId:          siteID,
		Description:     aws.String("Integration test link"),
		Bandwidth: &types.Bandwidth{
			UploadSpeed:   aws.Int32(50),
			DownloadSpeed: aws.Int32(50),
		},
		Tags: tags,
	}

	id, err := findLinkIDByTags(ctx, client, globalNetworkID, siteID, tags)
	if err != nil {
		if errors.As(err, new(integration.NotFoundError)) {
			logger.InfoContext(ctx, "Creating link")
		} else {
			return nil, err
		}
	}

	if id != nil {
		logger.InfoContext(ctx, "Link already exists")
		return id, nil
	}

	response, err := client.CreateLink(ctx, input)
	if err != nil {
		return nil, err
	}

	return response.Link.LinkId, nil
}
