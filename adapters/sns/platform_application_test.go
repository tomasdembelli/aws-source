package sns

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/overmindtech/aws-source/adapters"
)

type mockPlatformApplicationClient struct{}

func (m mockPlatformApplicationClient) ListTagsForResource(ctx context.Context, input *sns.ListTagsForResourceInput, f ...func(*sns.Options)) (*sns.ListTagsForResourceOutput, error) {
	return &sns.ListTagsForResourceOutput{
		Tags: []types.Tag{
			{Key: adapters.PtrString("tag1"), Value: adapters.PtrString("value1")},
			{Key: adapters.PtrString("tag2"), Value: adapters.PtrString("value2")},
		},
	}, nil
}

func (m mockPlatformApplicationClient) GetPlatformApplicationAttributes(ctx context.Context, params *sns.GetPlatformApplicationAttributesInput, optFns ...func(*sns.Options)) (*sns.GetPlatformApplicationAttributesOutput, error) {
	return &sns.GetPlatformApplicationAttributesOutput{
		Attributes: map[string]string{
			"Enabled":                   "true",
			"SuccessFeedbackSampleRate": "100",
		},
	}, nil
}

func (m mockPlatformApplicationClient) ListPlatformApplications(ctx context.Context, params *sns.ListPlatformApplicationsInput, optFns ...func(*sns.Options)) (*sns.ListPlatformApplicationsOutput, error) {
	return &sns.ListPlatformApplicationsOutput{
		PlatformApplications: []types.PlatformApplication{
			{
				PlatformApplicationArn: adapters.PtrString("arn:aws:sns:us-west-2:123456789012:app/ADM/MyApplication"),
				Attributes: map[string]string{
					"SuccessFeedbackSampleRate": "100",
					"Enabled":                   "true",
				},
			},
			{
				PlatformApplicationArn: adapters.PtrString("arn:aws:sns:us-west-2:123456789012:app/MPNS/MyOtherApplication"),
				Attributes: map[string]string{
					"SuccessFeedbackSampleRate": "100",
					"Enabled":                   "true",
				},
			},
		},
	}, nil
}

func TestGetPlatformApplicationFunc(t *testing.T) {
	ctx := context.Background()
	cli := mockPlatformApplicationClient{}

	item, err := getPlatformApplicationFunc(ctx, cli, "scope", &sns.GetPlatformApplicationAttributesInput{
		PlatformApplicationArn: adapters.PtrString("arn:aws:sns:us-west-2:123456789012:my-topic"),
	})
	if err != nil {
		t.Fatal(err)
	}

	if err = item.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestNewPlatformApplicationAdapter(t *testing.T) {
	client, account, region := GetAutoConfig(t)

	adapter := NewPlatformApplicationAdapter(client, account, region)

	test := adapters.E2ETest{
		Adapter: adapter,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
