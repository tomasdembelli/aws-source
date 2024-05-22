package integration

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/overmindtech/aws-source/sources"
	ec2overmind "github.com/overmindtech/aws-source/sources/ec2"
	"github.com/overmindtech/sdp-go"
)

func TestInstanceSource(t *testing.T) {
	ec2Cli, err := createEC2Client()
	if err != nil {
		t.Fatalf("failed to create EC2 client: %v", err)
	}

	accountID, found := os.LookupEnv("AWS_ACCOUNT_ID")
	if !found {
		t.Fatalf("AWS_ACCOUNT_ID not found")
	}

	region, found := os.LookupEnv("AWS_REGION")
	if !found {
		t.Fatalf("AWS_REGION not found")
	}

	instanceSource := ec2overmind.NewInstanceSource(ec2Cli, accountID, region)

	err = instanceSource.Validate()
	if err != nil {
		t.Fatalf("failed to validate EC2 instance source: %v", err)
	}

	scope := sources.FormatScope(accountID, region)

	// List instances
	sdpListInstances, err := instanceSource.List(context.Background(), scope, true)
	if err != nil {
		t.Fatalf("failed to list EC2 instances: %v", err)
	}

	instanceID, err := getInstanceID(sdpListInstances)
	if err != nil {
		t.Fatalf("failed to get instance ID: %v", err)
	}

	// Get instance
	sdpInstance, err := instanceSource.Get(context.Background(), scope, instanceID, true)
	if err != nil {
		t.Fatalf("failed to get EC2 instance: %v", err)
	}

	// assertions
	if sdpInstance.GetHealth() != sdp.Health_HEALTH_OK {
		t.Fatalf("expected instance to be healthy, got %v", sdpInstance.GetHealth())
	}

	val, ok := sdpInstance.GetTags()[tagKey]
	if !ok {
		t.Fatalf("expected tag key %v not found", tagKey)
	}
	if val != tagVal {
		t.Fatalf("expected tag value %v, got %v", tagVal, val)
	}

	// TODO: we can add more assertions for other attributes

	instanceARN := fmt.Sprintf("arn:aws:ec2:%s:%s:instance/%s", region, accountID, instanceID)
	sdpSearchInstances, err := instanceSource.Search(context.Background(), scope, instanceARN, true)
	if err != nil {
		t.Fatalf("failed to search EC2 instances: %v", err)
	}

	instanceIDFromSearch, err := getInstanceID(sdpSearchInstances)
	if err != nil {
		t.Fatalf("failed to get instance ID from search: %v", err)
	}

	if instanceIDFromSearch != instanceID {
		t.Fatalf("expected instance ID %v, got %v", instanceID, instanceIDFromSearch)
	}
}

func getInstanceID(sdpInstances []*sdp.Item) (string, error) {
	if len(sdpInstances) != 1 {
		return "", fmt.Errorf("expected 1 instance, got %v", len(sdpInstances))
	}

	instanceIDAttrVal, err := sdpInstances[0].GetAttributes().Get("instanceId")
	if err != nil {
		return "", fmt.Errorf("failed to get instanceId: %v", err)
	}

	instanceID := instanceIDAttrVal.(string)
	if instanceID == "" {
		return "", fmt.Errorf("instanceId is empty")
	}

	return instanceID, nil
}
