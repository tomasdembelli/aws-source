package integration

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi/types"
	"github.com/overmindtech/sdp-go"
)

const (
	TagTestKey       = "test"
	TagTestValue     = "true"
	TagTestIDKey     = "test-id"
	TagTestTypeKey   = "test-type"
	TagResourceIDKey = "resource-id"
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
		return "ec2"
	default:
		return "unknown"
	}
}

func ShouldRunIntegrationTests() bool {
	run, found := os.LookupEnv("RUN_INTEGRATION_TESTS")

	if !found {
		return false
	}

	shouldRun, err := strconv.ParseBool(run)
	if err != nil {
		return false
	}

	return shouldRun
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

func removeUnhealthy(sdpInstances []*sdp.Item) []*sdp.Item {
	var filteredInstances []*sdp.Item
	for _, instance := range sdpInstances {
		if instance.GetHealth() != sdp.Health_HEALTH_OK {
			continue
		}
		filteredInstances = append(filteredInstances, instance)
	}
	return filteredInstances
}

func GetUniqueAttributeValue(uniqueAttrKey string, items []*sdp.Item, filterTags map[string]string) (string, error) {
	var filteredItems []*sdp.Item
	for _, item := range removeUnhealthy(items) {
		if hasTags(item.GetTags(), filterTags) {
			filteredItems = append(filteredItems, item)
		}
	}

	if len(filteredItems) != 1 {
		return "", fmt.Errorf("expected 1 item, got %v", len(filteredItems))
	}

	uniqueAttrValue, err := filteredItems[0].GetAttributes().Get(uniqueAttrKey)
	if err != nil {
		return "", fmt.Errorf("failed to get %s: %v", uniqueAttrKey, err)
	}

	uniqueAttrValueStr := uniqueAttrValue.(string)
	if uniqueAttrValueStr == "" {
		return "", fmt.Errorf("%s is empty", uniqueAttrKey)
	}

	return uniqueAttrValueStr, nil
}

// ResourceName returns a unique resource name for integration tests
// I.e., integration-test-networkmanager-global-network-1
func ResourceName(resourceGroup resourceGroup, resourceName string, additionalAttr ...string) string {
	name := []string{"integration-test", resourceGroup.String(), resourceName}

	name = append(name, additionalAttr...)

	return strings.Join(name, "-")
}

func ResourceTags(resourceGroup resourceGroup, resourceName string, additionalAttr ...string) map[string]string {
	return map[string]string{
		TagTestKey:       TagTestValue,
		TagTestTypeKey:   TestName(resourceGroup),
		TagTestIDKey:     TestID(),
		TagResourceIDKey: ResourceName(resourceGroup, resourceName, additionalAttr...),
	}
}

func hasTags(tags map[string]string, requiredTags map[string]string) bool {
	for k, v := range requiredTags {
		if tags[k] != v {
			return false
		}
	}

	return true
}
