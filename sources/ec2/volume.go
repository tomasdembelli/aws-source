package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

func VolumeInputMapperGet(scope string, query string) (*ec2.DescribeVolumesInput, error) {
	return &ec2.DescribeVolumesInput{
		VolumeIds: []string{
			query,
		},
	}, nil
}

func VolumeInputMapperList(scope string) (*ec2.DescribeVolumesInput, error) {
	return &ec2.DescribeVolumesInput{}, nil
}

func VolumeOutputMapper(scope string, output *ec2.DescribeVolumesOutput) ([]*sdp.Item, error) {
	items := make([]*sdp.Item, 0)

	for _, volume := range output.Volumes {
		var err error
		var attrs *sdp.ItemAttributes
		attrs, err = sources.ToAttributesCase(volume)

		if err != nil {
			return nil, &sdp.ItemRequestError{
				ErrorType:   sdp.ItemRequestError_OTHER,
				ErrorString: err.Error(),
				Scope:       scope,
			}
		}

		item := sdp.Item{
			Type:            "ec2-volume",
			UniqueAttribute: "volumeId",
			Scope:           scope,
			Attributes:      attrs,
		}

		for _, attachment := range volume.Attachments {
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Type:   "ec2-instance",
				Method: sdp.RequestMethod_GET,
				Query:  *attachment.InstanceId,
				Scope:  scope,
			})
		}

		items = append(items, &item)
	}

	return items, nil
}

func NewVolumeSource(config aws.Config, accountID string) *EC2Source[*ec2.DescribeVolumesInput, *ec2.DescribeVolumesOutput] {
	return &EC2Source[*ec2.DescribeVolumesInput, *ec2.DescribeVolumesOutput]{
		Config:    config,
		AccountID: accountID,
		ItemType:  "ec2-volume",
		DescribeFunc: func(ctx context.Context, client *ec2.Client, input *ec2.DescribeVolumesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVolumesOutput, error) {
			return client.DescribeVolumes(ctx, input)
		},
		InputMapperGet:  VolumeInputMapperGet,
		InputMapperList: VolumeInputMapperList,
		PaginatorBuilder: func(client *ec2.Client, params *ec2.DescribeVolumesInput) Paginator[*ec2.DescribeVolumesOutput] {
			return ec2.NewDescribeVolumesPaginator(client, params)
		},
		OutputMapper: VolumeOutputMapper,
	}
}