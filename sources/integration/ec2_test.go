package integration

import "testing"

func TestIntegrationEC2(t *testing.T) {
	ShouldRunIntegrationTests(t)

	t.Run("Setup", func(t *testing.T) {
		t.Logf("Setting up EC2 integration tests")
	})

	t.Run("TestSomeSource", func(t *testing.T) {
		t.Logf("Running EC2 integration test TestSomeSource")
	})

	t.Run("Teardown", func(t *testing.T) {
		t.Logf("Tearing down EC2 integration tests")
	})
}
