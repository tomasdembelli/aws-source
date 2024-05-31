package integration

import "testing"

func TestIntegrationEFS(t *testing.T) {
	shouldRunIntegrationTests(t, "efs")

	t.Run("Setup", func(t *testing.T) {
		t.Logf("Setting up EFS integration tests")
	})

	t.Run("TestSomeSource", func(t *testing.T) {
		t.Logf("Running EFS integration test TestSomeSource")
	})

	t.Run("Teardown", func(t *testing.T) {
		t.Logf("Tearing down EFS integration tests")
	})
}
