package integration

import (
	"context"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

const (
	tagTestTypeKey   = "test"
	tagTestTypeValue = "ec2-integration"
	tagTestIDKey     = "test-id"
)

func testID() string {
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

//func TestMain(m *testing.M) {
//	if !main.shouldRunIntegrationTests() {
//		log.Println("skipping integration tests.. set RUN_INTEGRATION_TESTS=true to run them")
//		os.Exit(0)
//	}
//	log.Println("running integration tests..")
//	os.Exit(m.Run())
//}

func Setup() error {
	// Create EC2 client
	ec2Client, err := createEC2Client()
	if err != nil {
		return err
	}

	// Create EC2 instance
	return createEC2Instance(ec2Client, testID())
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
				Tags: []types.Tag{
					{
						Key:   aws.String(tagTestTypeKey),
						Value: aws.String(tagTestTypeValue),
					},
					{
						Key:   aws.String(tagTestIDKey),
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
