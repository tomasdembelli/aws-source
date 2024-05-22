package integration

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi/types"
)

func TestTeardown(t *testing.T) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		t.Fatalf("failed to load configuration, %v", err)
	}

	taggingClient := resourcegroupstaggingapi.NewFromConfig(cfg)
	ec2Client := ec2.NewFromConfig(cfg)

	// Get resources with the specified tag
	tagInput := &resourcegroupstaggingapi.GetResourcesInput{
		TagFilters: []types.TagFilter{
			{
				Key:    aws.String("your-tag-key"),
				Values: []string{"your-tag-value"},
			},
		},
	}

	tagOutput, err := taggingClient.GetResources(context.Background(), tagInput)
	if err != nil {
		t.Fatalf("failed to get resources, %v", err)
	}

	if len(tagOutput.ResourceTagMappingList) == 0 {
		t.Logf("no resources found with the specified tag, %s", "your-tag-key=your-tag-value")
		return
	}

	// Delete ec2 instances
	numOfDeletedEC2Instances := 0
	for _, resourceTagMapping := range tagOutput.ResourceTagMappingList {
		arn := aws.ToString(resourceTagMapping.ResourceARN)
		if arn == "" {
			continue
		}

		// TODO: Implement the logic to determine the service from the ARN and
		// delete the resource from the relevant service
		// for now we are assuming the ARN is for an EC2 instance

		// Call the EC2 deletion API
		_, err := ec2Client.TerminateInstances(context.Background(), &ec2.TerminateInstancesInput{
			InstanceIds: []string{arn},
		})
		if err != nil {
			t.Logf("failed to delete EC2 instance %s, %v", arn, err)
			continue
		}
		numOfDeletedEC2Instances++
		t.Logf("deleted EC2 instance %s", arn)
	}

	if numOfDeletedEC2Instances == 0 {
		t.Logf("no EC2 instances deleted with the specified tag, %s", "your-tag-key=your-tag-value")
		return
	}

	t.Logf("deleted %d EC2 instances with the specified tag, %s", numOfDeletedEC2Instances, "your-tag-key=your-tag-value")
}
