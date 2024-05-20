package ec2

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/overmindtech/aws-source/sources"
)

func createEC2Client() (*ec2.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}

	client := ec2.NewFromConfig(cfg)
	return client, nil
}

func createEC2Instance(client *ec2.Client) (*types.Instance, error) {
	input := &ec2.RunInstancesInput{
		ImageId:      aws.String("ami-0abcdef1234567890"), // replace with a free tier AMI
		InstanceType: types.InstanceTypeT2Nano,
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
	}

	result, err := client.RunInstances(context.Background(), input)
	if err != nil {
		return nil, err
	}

	// Use waiters here to wait for the instance to be in running state
	// https://aws.github.io/aws-sdk-go-v2/docs/making-requests/#using-waiters

	outputs, err := client.DescribeInstances(context.Background(), &ec2.DescribeInstancesInput{
		InstanceIds: []string{*result.Instances[0].InstanceId},
	})
	if err != nil {
		return nil, err
	}

	if len(outputs.Reservations) == 0 || len(outputs.Reservations[0].Instances) == 0 {
		return nil, fmt.Errorf("instance not found")
	}

	return &outputs.Reservations[0].Instances[0], nil
}

func deleteEC2Instance(client *ec2.Client, instanceID string) error {
	input := &ec2.TerminateInstancesInput{
		InstanceIds: []string{instanceID},
	}

	_, err := client.TerminateInstances(context.Background(), input)
	return err
}

func TestInstanceSource(t *testing.T) {
	ec2Cli, err := createEC2Client()
	if err != nil {
		t.Fatalf("failed to create EC2 client: %v", err)
	}

	ec2Instance, err := createEC2Instance(ec2Cli)
	if err != nil {
		t.Fatalf("failed to create EC2 instance: %v", err)
	}

	t.Cleanup(
		func() {
			if err := deleteEC2Instance(ec2Cli, *ec2Instance.InstanceId); err != nil {
				t.Fatalf("failed to delete EC2 instance: %v", err)
			}
		})

	accountID := "123456789012" // read from env or config
	region := "us-west-2"       // read from env or config

	instanceSource := NewInstanceSource(ec2Cli, accountID, region)

	if instanceSource.Validate() != nil {
		t.Fatalf("failed to validate EC2 instance source: %v", err)
	}

	scope := sources.FormatScope(accountID, region)
	sdpGetInstance, err := instanceSource.Get(context.Background(), scope, *ec2Instance.InstanceId, true)
	if err != nil {
		t.Fatalf("failed to get EC2 instance: %v", err)
	}

	fmt.Println(sdpGetInstance) // assert attributes, tags, health, etc.

	sdpListInstances, err := instanceSource.List(context.Background(), scope, true)
	if err != nil {
		t.Fatalf("failed to list EC2 instances: %v", err)
	}

	fmt.Println(sdpListInstances) // assert that there is single instance and it matches the instanceID

	sdpSearchInstances, err := instanceSource.Search(context.Background(), scope, "instance", true)
	if err != nil {
		t.Fatalf("failed to search EC2 instances: %v", err)
	}

	fmt.Println(sdpSearchInstances) // assert that there is single instance and it matches the instanceID
}
