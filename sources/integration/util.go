package integration

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi/types"
	"os"
	"strconv"
	"testing"
)

const (
	TagTestKey     = "test"
	TagTestValue   = "true"
	TagTestIDKey   = "test-id"
	TagTestTypeKey = "test-type"
)

type resourceGroup int

const (
	NetworkManager resourceGroup = iota
	EC2
)

func (rg resourceGroup) String() string {
	switch rg {
	case NetworkManager:
		return "network-manager"
	case EC2:
		return "EC2"
	default:
		return "unknown"
	}
}

func ShouldRunIntegrationTests(t *testing.T) {
	run, found := os.LookupEnv("RUN_INTEGRATION_TESTS")
	if found {
		shouldRun, err := strconv.ParseBool(run)
		if err != nil {
			t.Skipf("failed to parse RUN_INTEGRATION_TESTS")
			return
		}

		if !shouldRun {
			t.Skipf("skipping integration tests.. set RUN_INTEGRATION_TESTS=true to run them")
		}
	}
}

func TestID() string {
	tagTestID, found := os.LookupEnv("INTEGRATION_TEST_ID")
	if !found {
		var err error
		tagTestID, err = os.Hostname()
		if err != nil {
			panic("failed to get hostname")
		}
	}

	return tagTestID
}

func TestName(resourceGroup resourceGroup) string {
	return fmt.Sprintf("%s-integration-tests", resourceGroup.String())
}

func TagFilter(resourceGroup resourceGroup) []types.TagFilter {
	return []types.TagFilter{
		{
			Key:    aws.String(TagTestTypeKey),
			Values: []string{TestName(resourceGroup)},
		},
		{
			Key:    aws.String(TagTestIDKey),
			Values: []string{TestID()},
		},
		{
			Key:    aws.String(TagTestKey),
			Values: []string{TagTestValue},
		},
	}
}

type AWSCfg struct {
	AccountID string
	Region    string
}

func AWSSettings(ctx context.Context) (*AWSCfg, error) {
	accountID, found := os.LookupEnv("AWS_ACCOUNT_ID")
	if !found {
		return nil, fmt.Errorf("AWS_ACCOUNT_ID not found")
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	return &AWSCfg{
		AccountID: accountID,
		Region:    cfg.Region,
	}, nil
}
