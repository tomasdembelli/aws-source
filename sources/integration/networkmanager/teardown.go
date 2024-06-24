package networkmanager

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/overmindtech/aws-source/sources/integration"
	"log/slog"
)

func teardown(ctx context.Context, logger *slog.Logger, client *networkmanager.Client) error {
	globalNetworkID, err := findGlobalNetworkIDByTags(client)
	if err != nil {
		nf := integration.NewNotFoundError(globalNetworkSource)
		if errors.As(err, &nf) {
			logger.WarnContext(ctx, "Global network not found")
			return nil
		} else {
			return err
		}
	}

	return deleteGlobalNetwork(ctx, client, *globalNetworkID)
}
