package kms

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/overmindtech/aws-source/adapters"

	"github.com/aws/aws-sdk-go-v2/service/kms"
)

type testClient struct{}

func (t testClient) DescribeKey(ctx context.Context, params *kms.DescribeKeyInput, optFns ...func(*kms.Options)) (*kms.DescribeKeyOutput, error) {
	return &kms.DescribeKeyOutput{
		KeyMetadata: &types.KeyMetadata{
			AWSAccountId:          adapters.PtrString("846764612917"),
			KeyId:                 adapters.PtrString("b8a9477d-836c-491f-857e-07937918959b"),
			Arn:                   adapters.PtrString("arn:aws:kms:us-west-2:846764612917:key/b8a9477d-836c-491f-857e-07937918959b"),
			CreationDate:          adapters.PtrTime(time.Date(2017, 6, 30, 21, 44, 32, 140000000, time.UTC)),
			Enabled:               true,
			Description:           adapters.PtrString("Default KMS key that protects my S3 objects when no other key is defined"),
			KeyUsage:              types.KeyUsageTypeEncryptDecrypt,
			KeyState:              types.KeyStateEnabled,
			Origin:                types.OriginTypeAwsKms,
			KeyManager:            types.KeyManagerTypeAws,
			CustomerMasterKeySpec: types.CustomerMasterKeySpecSymmetricDefault,
			EncryptionAlgorithms: []types.EncryptionAlgorithmSpec{
				types.EncryptionAlgorithmSpecSymmetricDefault,
			},
		},
	}, nil
}

func (t testClient) ListKeys(context.Context, *kms.ListKeysInput, ...func(*kms.Options)) (*kms.ListKeysOutput, error) {
	return &kms.ListKeysOutput{
		Keys: []types.KeyListEntry{
			{
				KeyArn: adapters.PtrString("arn:aws:kms:us-west-2:111122223333:key/1234abcd-12ab-34cd-56ef-1234567890ab"),
				KeyId:  adapters.PtrString("1234abcd-12ab-34cd-56ef-1234567890ab"),
			},
			{
				KeyArn: adapters.PtrString("arn:aws:kms:us-west-2:111122223333:key/0987dcba-09fe-87dc-65ba-ab0987654321"),
				KeyId:  adapters.PtrString("0987dcba-09fe-87dc-65ba-ab0987654321"),
			},
			{
				KeyArn: adapters.PtrString("arn:aws:kms:us-east-2:111122223333:key/1a2b3c4d-5e6f-1a2b-3c4d-5e6f1a2b3c4d"),
				KeyId:  adapters.PtrString("1a2b3c4d-5e6f-1a2b-3c4d-5e6f1a2b3c4d"),
			},
		},
	}, nil
}

func (t testClient) ListResourceTags(context.Context, *kms.ListResourceTagsInput, ...func(*kms.Options)) (*kms.ListResourceTagsOutput, error) {
	return &kms.ListResourceTagsOutput{
		Tags: []types.Tag{
			{
				TagKey:   adapters.PtrString("Dept"),
				TagValue: adapters.PtrString("IT"),
			},
			{
				TagKey:   adapters.PtrString("Purpose"),
				TagValue: adapters.PtrString("Test"),
			},
			{
				TagKey:   adapters.PtrString("Name"),
				TagValue: adapters.PtrString("Test"),
			},
		},
	}, nil
}

func TestGetFunc(t *testing.T) {
	ctx := context.Background()
	cli := testClient{}

	item, err := getFunc(ctx, cli, "scope", &kms.DescribeKeyInput{
		KeyId: adapters.PtrString("1234abcd-12ab-34cd-56ef-1234567890ab"),
	})
	if err != nil {
		t.Fatal(err)
	}

	if err = item.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestNewKeyAdapter(t *testing.T) {
	config, account, region := adapters.GetAutoConfig(t)
	client := kms.NewFromConfig(config)

	adapter := NewKeyAdapter(client, account, region)

	test := adapters.E2ETest{
		Adapter: adapter,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
