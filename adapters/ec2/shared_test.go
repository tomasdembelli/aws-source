package ec2

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/adapters"
)

func GetAutoConfig(t *testing.T) (*ec2.Client, string, string) {
	config, account, region := adapters.GetAutoConfig(t)
	client := ec2.NewFromConfig(config)

	return client, account, region
}
