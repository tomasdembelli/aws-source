package ec2

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/overmindtech/aws-source/sources/integration"
	"time"
)

func setup() error {
	// Create EC2 client
	ec2Client, err := createEC2Client()
	if err != nil {
		return err
	}

	// Create EC2 instance
	return createEC2Instance(ec2Client, integration.TestID())
}

func createEC2Client() (*ec2.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}

	client := ec2.NewFromConfig(cfg)
	return client, nil
}

func createEC2Instance(client *ec2.Client, testID string) error {
	input := &ec2.RunInstancesInput{
		ImageId:      aws.String("ami-0abcdef1234567890"), // replace with a free tier AMI
		InstanceType: types.InstanceTypeT2Nano,
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceTypeInstance,
				// TODO: Create a convenience function to add shared tags to the resources
				Tags: []types.Tag{
					{
						Key:   aws.String(integration.TagTestKey),
						Value: aws.String(integration.TagTestValue),
					},
					{
						Key:   aws.String(integration.TagTestTypeKey),
						Value: aws.String(integration.TestName(integration.EC2)),
					},
					{
						Key:   aws.String(integration.TagTestIDKey),
						Value: aws.String(testID),
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
