package kms

import (
	"context"
	"testing"
	"time"

	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"

	"github.com/aws/aws-sdk-go-v2/service/kms"
)

/*
Example key policy values

{
    "Version" : "2012-10-17",
    "Id" : "key-default-1",
    "Statement" : [
        {
            "Sid" : "Enable IAM User Permissions",
            "Effect" : "Allow",
            "Principal" : {
                "AWS" : "arn:aws:iam::111122223333:root"
            },
            "Action" : "kms:*",
            "Resource" : "*"
            },
            {
            "Sid" : "Allow Use of Key",
            "Effect" : "Allow",
            "Principal" : {
                "AWS" : "arn:aws:iam::111122223333:user/test-user"
            },
            "Action" : [ "kms:Describe", "kms:List" ],
            "Resource" : "*"
        }
    ]
}
*/

type mockKeyPolicyClient struct{}

func (m *mockKeyPolicyClient) GetKeyPolicy(ctx context.Context, params *kms.GetKeyPolicyInput, optFns ...func(*kms.Options)) (*kms.GetKeyPolicyOutput, error) {
	return &kms.GetKeyPolicyOutput{
		Policy: adapters.PtrString(`{
			"Version" : "2012-10-17",
			"Id" : "key-default-1",
			"Statement" : [
				{
					"Sid" : "Enable IAM User Permissions",
					"Effect" : "Allow",
					"Principal" : {
						"AWS" : "arn:aws:iam::111122223333:root"
					},
					"Action" : "kms:*",
					"Resource" : "*"
				},
				{
					"Sid" : "Allow Use of Key",
					"Effect" : "Allow",
					"Principal" : {
						"AWS" : "arn:aws:iam::111122223333:user/test-user"
					},
					"Action" : [ "kms:Describe", "kms:List" ],
					"Resource" : "*"
				}
			]
		}`),
		PolicyName: adapters.PtrString("default"),
	}, nil
}

func (m *mockKeyPolicyClient) ListKeyPolicies(ctx context.Context, params *kms.ListKeyPoliciesInput, optFns ...func(*kms.Options)) (*kms.ListKeyPoliciesOutput, error) {
	return &kms.ListKeyPoliciesOutput{
		PolicyNames: []string{"default"},
	}, nil
}

func TestGetKeyPolicyFunc(t *testing.T) {
	ctx := context.Background()
	cli := &mockKeyPolicyClient{}

	item, err := getKeyPolicyFunc(ctx, cli, "scope", &kms.GetKeyPolicyInput{
		KeyId: adapters.PtrString("1234abcd-12ab-34cd-56ef-1234567890ab"),
	})
	if err != nil {
		t.Fatal(err)
	}

	if err = item.Validate(); err != nil {
		t.Fatal(err)
	}

	tests := adapters.QueryTests{
		{
			ExpectedType:   "kms-key",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "1234abcd-12ab-34cd-56ef-1234567890ab",
			ExpectedScope:  "scope",
		},
	}

	tests.Execute(t, item)
}

func TestNewKeyPolicyAdapter(t *testing.T) {
	config, account, region := adapters.GetAutoConfig(t)

	client := kms.NewFromConfig(config)

	adapter := NewKeyPolicyAdapter(client, account, region)

	test := adapters.E2ETest{
		Adapter:  adapter,
		Timeout:  10 * time.Second,
		SkipList: true,
	}

	test.Run(t)
}
