package ec2

import (
	"context"
	"fmt"
	"github.com/overmindtech/aws-source/sources/integration"
	"github.com/overmindtech/sdp-go"
	"testing"
)

func TestIntegrationEC2(t *testing.T) {
	integration.ShouldRunIntegrationTests(t)

	ctx := context.Background()

	t.Run("Setup", func(t *testing.T) {
		if err := setup(); err != nil {
			t.Fatalf("Failed to setup EC2 integration tests: %v", err)
		}
	})

	t.Run("TestSomeSource", func(t *testing.T) {
		t.Logf("Running EC2 integration test TestSomeSource")
	})

	t.Run("Teardown", func(t *testing.T) {
		if err := integration.Teardown(ctx, integration.TagFilter(integration.EC2)); err != nil {
			t.Fatalf("Failed to teardown EC2 integration tests: %v", err)
		}
	})
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
