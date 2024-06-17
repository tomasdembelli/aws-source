package integration

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi/types"
	"log/slog"
	"strings"
)

func Teardown(ctx context.Context, tagFilter []types.TagFilter) error {
	logger := slog.Default()

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return err
	}

	taggingClient := resourcegroupstaggingapi.NewFromConfig(cfg)
	ec2Client := ec2.NewFromConfig(cfg)
	networkManagerClient := networkmanager.NewFromConfig(cfg)

	// Get resources with the specified tag
	tagInput := &resourcegroupstaggingapi.GetResourcesInput{
		TagFilters: tagFilter,
	}

	tagOutput, err := taggingClient.GetResources(context.Background(), tagInput)
	if err != nil {
		return err
	}

	// make this idempotent
	if len(tagOutput.ResourceTagMappingList) == 0 {

		logger.WarnContext(ctx, "no resources found with the specified tags", "tag_filter", tagFilter)

		return nil
	}

	// Delete ec2 instances
	numOfDeletedEC2Instances := 0
	numOfDeletedNetworkManagerConnectAttachments := 0

	// For each resource, delete it
	for _, resourceTagMapping := range tagOutput.ResourceTagMappingList {
		arn := aws.ToString(resourceTagMapping.ResourceARN)

		// Determine the service from the ARN
		service := determineServiceFromARN(arn)

		switch service {
		case "ec2":
			// Call the EC2 deletion API
			_, err := ec2Client.TerminateInstances(context.Background(), &ec2.TerminateInstancesInput{
				InstanceIds: []string{strings.Split(arn, "/")[1]},
			})
			if err != nil {
				logger.ErrorContext(ctx,
					"failed to delete EC2 instance",
					slog.String("arn", arn),
					slog.String("err", err.Error()),
				)
				continue
			} else {
				numOfDeletedEC2Instances++
				logger.InfoContext(ctx, "deleted EC2 instance", slog.String("arn", arn))
			}
		case "networkmanager":
			// Determine the specific type of networkmanager resource from the ARN
			resourceType := determineResourceTypeFromARN(arn)

			switch resourceType {
			case "connect-attachment":
				// Call the Network Manager deletion API for ConnectAttachment
				_, err := networkManagerClient.DeleteAttachment(
					context.Background(),
					&networkmanager.DeleteAttachmentInput{
						AttachmentId: aws.String(arn),
					},
				)
				if err != nil {
					logger.ErrorContext(
						ctx,
						"failed to delete Network Manager ConnectAttachment",
						slog.String("arn", arn),
						slog.String("err", err.Error()),
					)
					continue
				} else {
					numOfDeletedNetworkManagerConnectAttachments++
					logger.InfoContext(
						ctx,
						"deleted Network Manager ConnectAttachment",
						slog.String("arn", arn),
					)
				}
			default:
				logger.Warn("Unsupported networkmanager resource type: ", resourceType)
			}
		default:
			fmt.Println("Unsupported service: ", service)
		}
	}

	logger.Info("deleted resources for the given tags",
		slog.Int("num_of_deleted_ec2_instances", numOfDeletedEC2Instances),
		slog.Int("num_of_deleted_network_manager_connect_attachments", numOfDeletedNetworkManagerConnectAttachments),
	)

	return nil
}

func determineServiceFromARN(arn string) string {
	// Implement this function to determine the service from the ARN
	// This is a simplified example and may not work for all ARNs
	parts := strings.Split(arn, ":")
	if len(parts) > 2 {
		return parts[2]
	}
	return ""
}

func determineResourceTypeFromARN(arn string) string {
	parts := strings.Split(arn, ":")
	if len(parts) < 6 {
		return ""
	}

	resourceParts := strings.Split(parts[5], "/")
	if len(resourceParts) < 1 {
		return ""
	}

	return resourceParts[0]
}
