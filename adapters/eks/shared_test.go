package eks

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/overmindtech/aws-source/adapters"
)

func GetAutoConfig(t *testing.T) (*eks.Client, string, string) {
	config, account, region := adapters.GetAutoConfig(t)
	client := eks.NewFromConfig(config)

	return client, account, region
}
