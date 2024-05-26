package integration

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi/types"
	"log/slog"
)

func Teardown(logger *slog.Logger) error {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return err
	}

	taggingClient := resourcegroupstaggingapi.NewFromConfig(cfg)
	ec2Client := ec2.NewFromConfig(cfg)

	// Get resources with the specified tag
	tagInput := &resourcegroupstaggingapi.GetResourcesInput{
		TagFilters: []types.TagFilter{
			{
				Key:    aws.String(tagTestTypeKey),
				Values: []string{tagTestTypeValue},
			},
			{
				Key:    aws.String(tagTestIDKey),
				Values: []string{testID()},
			},
		},
	}

	tagOutput, err := taggingClient.GetResources(context.Background(), tagInput)
	if err != nil {
		return err
	}

	if len(tagOutput.ResourceTagMappingList) == 0 {
		return fmt.Errorf(
			"no resources found with the specified tags: %s:%s, %s:%s",
			tagTestTypeKey,
			tagTestTypeValue,
			tagTestIDKey,
			testID(),
		)
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
		_, err = ec2Client.TerminateInstances(context.Background(), &ec2.TerminateInstancesInput{
			InstanceIds: []string{arn},
		})
		if err != nil {
			logger.Error(
				"failed to delete EC2 instance",
				slog.String("arn", arn),
				slog.String("err", err.Error()),
			)
			continue
		}
		numOfDeletedEC2Instances++
		logger.Info("deleted EC2 instance", slog.String("arn", arn))
	}

	if numOfDeletedEC2Instances == 0 {
		logger.Warn(
			"no EC2 instances deleted with the specified tags",
			slog.String(tagTestTypeKey, tagTestTypeValue),
			slog.String(tagTestIDKey, testID()),
		)
		return nil
	}

	logger.Info("deleted EC2 instances with the specified tags",
		slog.String(tagTestTypeKey, tagTestTypeValue),
		slog.String(tagTestIDKey, testID()),
		slog.Int("num_of_deleted_ec2_instances", numOfDeletedEC2Instances),
	)

	return nil
}
