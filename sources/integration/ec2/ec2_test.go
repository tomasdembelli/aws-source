package ec2

import (
	"context"
	"github.com/overmindtech/aws-source/sources/integration"
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
		t.Logf("Running EC2 integration tests")
		TestInstanceSource(t)
	})

	t.Run("Teardown", func(t *testing.T) {
		if err := integration.Teardown(ctx, integration.TagFilter(integration.EC2)); err != nil {
			t.Fatalf("Failed to teardown EC2 integration tests: %v", err)
		}
	})
}
