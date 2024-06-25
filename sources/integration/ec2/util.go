package ec2

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/overmindtech/aws-source/sources/integration"
)

func resourceTags(resourceName, testID string, nameAdditionalAttr ...string) []types.Tag {
	return []types.Tag{
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
		{
			Key:   aws.String(integration.TagResourceIDKey),
			Value: aws.String(integration.ResourceName(integration.EC2, resourceName, nameAdditionalAttr...)),
		},
	}
}

func hasTags(tags []types.Tag, requiredTags []types.Tag) bool {
	rT := make(map[string]string)
	for _, t := range requiredTags {
		rT[*t.Key] = *t.Value
	}

	oT := make(map[string]string)
	for _, t := range tags {
		oT[*t.Key] = *t.Value
	}

	for k, v := range rT {
		if oT[k] != v {
			return false
		}
	}

	return true
}

// findActiveInstanceIDByTags finds an instance by tags
// additionalAttr is a variadic parameter that allows to specify additional attributes to search for
// it ignores terminated instances
func findActiveInstanceIDByTags(client *ec2.Client, additionalAttr ...string) (*string, error) {
	result, err := client.DescribeInstances(context.Background(), &ec2.DescribeInstancesInput{})
	if err != nil {
		return nil, err
	}

	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			// ignore terminated or shutting down instances
			if instance.State.Name == types.InstanceStateNameTerminated ||
				instance.State.Name == types.InstanceStateNameShuttingDown {
				// ignore terminated instances
				continue
			}

			if hasTags(instance.Tags, resourceTags(instanceSrc, integration.TestID(), additionalAttr...)) {
				return instance.InstanceId, nil
			}
		}
	}

	return nil, integration.NewNotFoundError(integration.ResourceName(integration.EC2, instanceSrc, additionalAttr...))
}

func deleteInstance(ctx context.Context, client *ec2.Client, instanceID string) error {
	input := &ec2.TerminateInstancesInput{
		InstanceIds: []string{instanceID},
	}

	_, err := client.TerminateInstances(ctx, input)
	return err
}
