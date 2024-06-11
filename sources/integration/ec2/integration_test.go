package ec2

import (
	"context"
	"fmt"
	"github.com/overmindtech/aws-source/sources/integration"
	"log/slog"
	"testing"

	"github.com/overmindtech/aws-source/sources"
	ec2overmind "github.com/overmindtech/aws-source/sources/ec2"
)

func TestInstanceSource(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	logger.Info("Running EC2 integration test TestInstanceSource")

	ec2Cli, err := createEC2Client()
	if err != nil {
		t.Fatalf("failed to create EC2 client: %v", err)
	}

	awsCfg, err := integration.AWSSettings(ctx)
	if err != nil {
		t.Fatalf("failed to get AWS settings: %v", err)
	}

	instanceSource := ec2overmind.NewInstanceSource(ec2Cli, awsCfg.AccountID, awsCfg.Region)

	err = instanceSource.Validate()
	if err != nil {
		t.Fatalf("failed to validate EC2 instance source: %v", err)
	}

	scope := sources.FormatScope(awsCfg.AccountID, awsCfg.Region)

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

	if sdpInstance.GetUniqueAttribute() != instanceID {
		t.Fatalf("expected instance ID %v, got %v", instanceID, sdpInstance.GetUniqueAttribute())
	}

	instanceARN := fmt.Sprintf("arn:aws:ec2:%s:%s:instance/%s", awsCfg.AccountID, awsCfg.Region, instanceID)
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
