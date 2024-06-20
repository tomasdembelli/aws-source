package networkmanager

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/overmindtech/aws-source/sources/integration"
	"log/slog"
)

const globalNetworkSource = "global-network"

func setup(ctx context.Context, logger *slog.Logger, networkmanagerClient *networkmanager.Client) error {

	// Create a global network
	return createGlobalNetwork(ctx, logger, networkmanagerClient, integration.TestID())
}

func createNetworkManagerClient(ctx context.Context) (*networkmanager.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	client := networkmanager.NewFromConfig(cfg)

	return client, nil
}

func createGlobalNetwork(ctx context.Context, logger *slog.Logger, client *networkmanager.Client, testID string) error {
	tags := resourceTags(globalNetworkSource, testID)

	input := &networkmanager.CreateGlobalNetworkInput{
		Description: aws.String("Integration test global network"),
		Tags:        tags,
	}

	id, err := findGlobalNetworkIDByTags(client)
	if err != nil {
		if errors.As(err, new(integration.NotFoundError)) {
			logger.InfoContext(ctx, "Creating global network")
		} else {
			return err
		}
	}

	if id != nil {
		logger.InfoContext(ctx, "Global network already exists")
		return nil
	}

	_, err = client.CreateGlobalNetwork(context.Background(), input)
	return err
}
