package integration

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

const (
	tagKey = "test"
	tagVal = "ec2-integration"
)

func TestSetup(t *testing.T) {
	// Create EC2 client
	ec2Client, err := createEC2Client()
	if err != nil {
		t.Fatalf("failed to create EC2 client: %v", err)
	}

	// Create EC2 instance
	err = createEC2Instance(ec2Client)
	if err != nil {
		t.Fatalf("failed to create EC2 instance: %v", err)
	}
}

func createEC2Client() (*ec2.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}

	client := ec2.NewFromConfig(cfg)
	return client, nil
}

func createEC2Instance(client *ec2.Client) error {
	input := &ec2.RunInstancesInput{
		ImageId:      aws.String("ami-0abcdef1234567890"), // replace with a free tier AMI
		InstanceType: types.InstanceTypeT2Nano,
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceTypeInstance,
				Tags: []types.Tag{
					{
						Key:   aws.String(tagKey),
						Value: aws.String(tagVal),
					},
				},
			},
		},
	}

	result, err := client.RunInstances(context.Background(), input)
	if err != nil {
		return err
	}

	waiter := ec2.NewInstanceRunningWaiter(client)
	err = waiter.Wait(context.Background(), &ec2.DescribeInstancesInput{
		InstanceIds: []string{*result.Instances[0].InstanceId},
	},
		5*time.Minute)
	if err != nil {
		return err
	}

	return nil
}
