package networkmanager

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/overmindtech/aws-source/sources/integration"
)

func setup(ctx context.Context) error {
	networkmanagerClient, err := createNetworkManagerClient(ctx)
	if err != nil {
		return err
	}

	// Create a global network
	return createGlobalNetwork(networkmanagerClient, integration.TestID())
}

func createNetworkManagerClient(ctx context.Context) (*networkmanager.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	client := networkmanager.NewFromConfig(cfg)

	return client, nil
}

func createGlobalNetwork(client *networkmanager.Client, testID string) error {
	input := &networkmanager.CreateGlobalNetworkInput{
		Description: aws.String("Integration test global network"),
		Tags: []types.Tag{
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
		},
	}

	_, err := client.CreateGlobalNetwork(context.Background(), input)
	return err
}
