package s3

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"
	"github.com/overmindtech/sdpcache"
)

func TestS3SearchImpl(t *testing.T) {
	cache := sdpcache.NewCache()
	t.Run("with a good ARN", func(t *testing.T) {
		items, err := searchImpl(context.Background(), cache, TestS3Client{}, "account-id.region", "arn:partition:service:region:account-id:resource-type:resource-id", false)

		if err != nil {
			t.Error(err)
		}
		if len(items) != 1 {
			t.Errorf("expected 1 item, got %v", len(items))
		}
	})

	t.Run("with a bad ARN", func(t *testing.T) {
		_, err := searchImpl(context.Background(), cache, TestS3Client{}, "account-id.region", "foo", false)

		if err == nil {
			t.Error("expected error")
		} else {
			var ire *sdp.QueryError
			if errors.As(err, &ire) {
				if ire.GetErrorType() != sdp.QueryError_OTHER {
					t.Errorf("expected error type to be OTHER, got %v", ire.GetErrorType().String())
				}
			} else {
				t.Errorf("expected item request error, got %T", err)
			}
		}
	})

	t.Run("with an ARN in another scope", func(t *testing.T) {
		_, err := searchImpl(context.Background(), cache, TestS3Client{}, "account-id.region", "arn:partition:service:region:account-id-2:resource-type:resource-id", false)

		if err == nil {
			t.Error("expected error")
		} else {
			var ire *sdp.QueryError
			if errors.As(err, &ire) {
				if ire.GetErrorType() != sdp.QueryError_NOSCOPE {
					t.Errorf("expected error type to be OTHER, got %v", ire.GetErrorType().String())
				}
			} else {
				t.Errorf("expected item request error, got %T", err)
			}
		}
	})
}

func TestS3ListImpl(t *testing.T) {
	cache := sdpcache.NewCache()
	items, err := listImpl(context.Background(), cache, TestS3Client{}, "foo", false)

	if err != nil {
		t.Error(err)
	}
	if len(items) != 1 {
		t.Errorf("expected 1 item, got %v", len(items))
	}
}

func TestS3GetImpl(t *testing.T) {
	cache := sdpcache.NewCache()
	item, err := getImpl(context.Background(), cache, TestS3Client{}, "foo", "bar", false)

	if err != nil {
		t.Fatal(err)
	}

	tests := adapters.QueryTests{
		{
			ExpectedType:   "http",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "https://hostname",
			ExpectedScope:  "global",
		},
		{
			ExpectedType:   "lambda-function",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:partition:service:region:account-id:resource-type:resource-id",
			ExpectedScope:  "account-id.region",
		},
		{
			ExpectedType:   "sqs-queue",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:partition:service:region:account-id:resource-type:resource-id",
			ExpectedScope:  "account-id.region",
		},
		{
			ExpectedType:   "sns-topic",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:partition:service:region:account-id:resource-type:resource-id",
			ExpectedScope:  "account-id.region",
		},
		{
			ExpectedType:   "s3-bucket",
			ExpectedMethod: sdp.QueryMethod_GET,
			ExpectedQuery:  "bucket",
			ExpectedScope:  "foo",
		},
		{
			ExpectedType:   "s3-bucket",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:partition:service:region:account-id:resource-type:resource-id",
			ExpectedScope:  "account-id.region",
		},
		{
			ExpectedType:   "s3-bucket",
			ExpectedMethod: sdp.QueryMethod_SEARCH,
			ExpectedQuery:  "arn:partition:service:region:account-id:resource-type:resource-id",
			ExpectedScope:  "account-id.region",
		},
	}

	tests.Execute(t, item)
}

func TestS3SourceCaching(t *testing.T) {
	cache := sdpcache.NewCache()
	first, err := getImpl(context.Background(), cache, TestS3Client{}, "foo", "bar", false)
	if err != nil {
		t.Fatal(err)
	}
	if first == nil {
		t.Fatal("expected first item")
	}

	second, err := getImpl(context.Background(), cache, TestS3FailClient{}, "foo", "bar", false)
	if err != nil {
		t.Fatal(err)
	}
	if second == nil {
		t.Fatal("expected second item")
	}

	third, err := getImpl(context.Background(), cache, TestS3Client{}, "foo", "bar", true)
	if err != nil {
		t.Fatal(err)
	}
	if third == nil {
		t.Fatal("expected third item")
	}

	if third == second {
		t.Errorf("expected third item (%v) to be different to second item (%v)", third, second)
	}
}

var owner = types.Owner{
	DisplayName: adapters.PtrString("dylan"),
	ID:          adapters.PtrString("id"),
}

// TestS3Client A client that returns example data
type TestS3Client struct{}

func (t TestS3Client) ListBuckets(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error) {
	return &s3.ListBucketsOutput{
		Buckets: []types.Bucket{
			{
				CreationDate: adapters.PtrTime(time.Now()),
				Name:         adapters.PtrString("foo"),
			},
		},
		Owner: &owner,
	}, nil
}

func (t TestS3Client) GetBucketAcl(ctx context.Context, params *s3.GetBucketAclInput, optFns ...func(*s3.Options)) (*s3.GetBucketAclOutput, error) {
	return &s3.GetBucketAclOutput{
		Grants: []types.Grant{
			{
				Grantee: &types.Grantee{
					Type:         types.TypeAmazonCustomerByEmail,
					DisplayName:  adapters.PtrString("dylan"),
					EmailAddress: adapters.PtrString("dylan@company.com"),
					ID:           adapters.PtrString("id"),
					URI:          adapters.PtrString("uri"),
				},
			},
		},
		Owner: &owner,
	}, nil
}

func (t TestS3Client) GetBucketAnalyticsConfiguration(ctx context.Context, params *s3.GetBucketAnalyticsConfigurationInput, optFns ...func(*s3.Options)) (*s3.GetBucketAnalyticsConfigurationOutput, error) {
	return &s3.GetBucketAnalyticsConfigurationOutput{
		AnalyticsConfiguration: &types.AnalyticsConfiguration{
			Id: adapters.PtrString("id"),
			StorageClassAnalysis: &types.StorageClassAnalysis{
				DataExport: &types.StorageClassAnalysisDataExport{
					Destination: &types.AnalyticsExportDestination{
						S3BucketDestination: &types.AnalyticsS3BucketDestination{
							Bucket:          adapters.PtrString("arn:partition:service:region:account-id:resource-type:resource-id"),
							Format:          types.AnalyticsS3ExportFileFormatCsv,
							BucketAccountId: adapters.PtrString("id"),
							Prefix:          adapters.PtrString("pre"),
						},
					},
					OutputSchemaVersion: types.StorageClassAnalysisSchemaVersionV1,
				},
			},
		},
	}, nil
}

func (t TestS3Client) GetBucketCors(ctx context.Context, params *s3.GetBucketCorsInput, optFns ...func(*s3.Options)) (*s3.GetBucketCorsOutput, error) {
	return &s3.GetBucketCorsOutput{
		CORSRules: []types.CORSRule{
			{
				AllowedMethods: []string{
					"GET",
				},
				AllowedOrigins: []string{
					"amazon.com",
				},
				AllowedHeaders: []string{
					"Authorization",
				},
				ExposeHeaders: []string{
					"foo",
				},
				ID:            adapters.PtrString("id"),
				MaxAgeSeconds: adapters.PtrInt32(10),
			},
		},
	}, nil
}

func (t TestS3Client) GetBucketEncryption(ctx context.Context, params *s3.GetBucketEncryptionInput, optFns ...func(*s3.Options)) (*s3.GetBucketEncryptionOutput, error) {
	return &s3.GetBucketEncryptionOutput{
		ServerSideEncryptionConfiguration: &types.ServerSideEncryptionConfiguration{
			Rules: []types.ServerSideEncryptionRule{
				{
					ApplyServerSideEncryptionByDefault: &types.ServerSideEncryptionByDefault{
						SSEAlgorithm:   types.ServerSideEncryptionAes256,
						KMSMasterKeyID: adapters.PtrString("id"),
					},
					BucketKeyEnabled: adapters.PtrBool(true),
				},
			},
		},
	}, nil
}

func (t TestS3Client) GetBucketIntelligentTieringConfiguration(ctx context.Context, params *s3.GetBucketIntelligentTieringConfigurationInput, optFns ...func(*s3.Options)) (*s3.GetBucketIntelligentTieringConfigurationOutput, error) {
	return &s3.GetBucketIntelligentTieringConfigurationOutput{
		IntelligentTieringConfiguration: &types.IntelligentTieringConfiguration{
			Id:     adapters.PtrString("id"),
			Status: types.IntelligentTieringStatusEnabled,
			Tierings: []types.Tiering{
				{
					AccessTier: types.IntelligentTieringAccessTierDeepArchiveAccess,
					Days:       adapters.PtrInt32(100),
				},
			},
			Filter: &types.IntelligentTieringFilter{},
		},
	}, nil
}

func (t TestS3Client) GetBucketInventoryConfiguration(ctx context.Context, params *s3.GetBucketInventoryConfigurationInput, optFns ...func(*s3.Options)) (*s3.GetBucketInventoryConfigurationOutput, error) {
	return &s3.GetBucketInventoryConfigurationOutput{
		InventoryConfiguration: &types.InventoryConfiguration{
			Destination: &types.InventoryDestination{
				S3BucketDestination: &types.InventoryS3BucketDestination{
					Bucket:    adapters.PtrString("arn:partition:service:region:account-id:resource-type:resource-id"),
					Format:    types.InventoryFormatCsv,
					AccountId: adapters.PtrString("id"),
					Encryption: &types.InventoryEncryption{
						SSEKMS: &types.SSEKMS{
							KeyId: adapters.PtrString("key"),
						},
					},
					Prefix: adapters.PtrString("pre"),
				},
			},
			Id:                     adapters.PtrString("id"),
			IncludedObjectVersions: types.InventoryIncludedObjectVersionsAll,
			IsEnabled:              adapters.PtrBool(true),
			Schedule: &types.InventorySchedule{
				Frequency: types.InventoryFrequencyDaily,
			},
		},
	}, nil
}

func (t TestS3Client) GetBucketLifecycleConfiguration(ctx context.Context, params *s3.GetBucketLifecycleConfigurationInput, optFns ...func(*s3.Options)) (*s3.GetBucketLifecycleConfigurationOutput, error) {
	return &s3.GetBucketLifecycleConfigurationOutput{
		Rules: []types.LifecycleRule{
			{
				Status: types.ExpirationStatusEnabled,
				AbortIncompleteMultipartUpload: &types.AbortIncompleteMultipartUpload{
					DaysAfterInitiation: adapters.PtrInt32(1),
				},
				Expiration: &types.LifecycleExpiration{
					Date:                      adapters.PtrTime(time.Now()),
					Days:                      adapters.PtrInt32(3),
					ExpiredObjectDeleteMarker: adapters.PtrBool(true),
				},
				ID: adapters.PtrString("id"),
				NoncurrentVersionExpiration: &types.NoncurrentVersionExpiration{
					NewerNoncurrentVersions: adapters.PtrInt32(3),
					NoncurrentDays:          adapters.PtrInt32(1),
				},
				NoncurrentVersionTransitions: []types.NoncurrentVersionTransition{
					{
						NewerNoncurrentVersions: adapters.PtrInt32(1),
						NoncurrentDays:          adapters.PtrInt32(1),
						StorageClass:            types.TransitionStorageClassGlacierIr,
					},
				},
				Prefix: adapters.PtrString("pre"),
				Transitions: []types.Transition{
					{
						Date:         adapters.PtrTime(time.Now()),
						Days:         adapters.PtrInt32(12),
						StorageClass: types.TransitionStorageClassGlacierIr,
					},
				},
			},
		},
	}, nil
}

func (t TestS3Client) GetBucketLocation(ctx context.Context, params *s3.GetBucketLocationInput, optFns ...func(*s3.Options)) (*s3.GetBucketLocationOutput, error) {
	return &s3.GetBucketLocationOutput{
		LocationConstraint: types.BucketLocationConstraintAfSouth1,
	}, nil
}

func (t TestS3Client) GetBucketLogging(ctx context.Context, params *s3.GetBucketLoggingInput, optFns ...func(*s3.Options)) (*s3.GetBucketLoggingOutput, error) {
	return &s3.GetBucketLoggingOutput{
		LoggingEnabled: &types.LoggingEnabled{
			TargetBucket: adapters.PtrString("bucket"),
			TargetPrefix: adapters.PtrString("pre"),
			TargetGrants: []types.TargetGrant{
				{
					Grantee: &types.Grantee{
						Type: types.TypeGroup,
						ID:   adapters.PtrString("id"),
					},
				},
			},
		},
	}, nil
}

func (t TestS3Client) GetBucketMetricsConfiguration(ctx context.Context, params *s3.GetBucketMetricsConfigurationInput, optFns ...func(*s3.Options)) (*s3.GetBucketMetricsConfigurationOutput, error) {
	return &s3.GetBucketMetricsConfigurationOutput{
		MetricsConfiguration: &types.MetricsConfiguration{
			Id: adapters.PtrString("id"),
		},
	}, nil
}

func (t TestS3Client) GetBucketNotificationConfiguration(ctx context.Context, params *s3.GetBucketNotificationConfigurationInput, optFns ...func(*s3.Options)) (*s3.GetBucketNotificationConfigurationOutput, error) {
	return &s3.GetBucketNotificationConfigurationOutput{
		LambdaFunctionConfigurations: []types.LambdaFunctionConfiguration{
			{
				Events:            []types.Event{},
				LambdaFunctionArn: adapters.PtrString("arn:partition:service:region:account-id:resource-type:resource-id"),
				Id:                adapters.PtrString("id"),
			},
		},
		EventBridgeConfiguration: &types.EventBridgeConfiguration{},
		QueueConfigurations: []types.QueueConfiguration{
			{
				Events:   []types.Event{},
				QueueArn: adapters.PtrString("arn:partition:service:region:account-id:resource-type:resource-id"),
				Filter: &types.NotificationConfigurationFilter{
					Key: &types.S3KeyFilter{
						FilterRules: []types.FilterRule{
							{
								Name:  types.FilterRuleNamePrefix,
								Value: adapters.PtrString("foo"),
							},
						},
					},
				},
				Id: adapters.PtrString("id"),
			},
		},
		TopicConfigurations: []types.TopicConfiguration{
			{
				Events:   []types.Event{},
				TopicArn: adapters.PtrString("arn:partition:service:region:account-id:resource-type:resource-id"),
				Filter: &types.NotificationConfigurationFilter{
					Key: &types.S3KeyFilter{
						FilterRules: []types.FilterRule{
							{
								Name:  types.FilterRuleNameSuffix,
								Value: adapters.PtrString("fix"),
							},
						},
					},
				},
				Id: adapters.PtrString("id"),
			},
		},
	}, nil
}

func (t TestS3Client) GetBucketOwnershipControls(ctx context.Context, params *s3.GetBucketOwnershipControlsInput, optFns ...func(*s3.Options)) (*s3.GetBucketOwnershipControlsOutput, error) {
	return &s3.GetBucketOwnershipControlsOutput{
		OwnershipControls: &types.OwnershipControls{
			Rules: []types.OwnershipControlsRule{
				{
					ObjectOwnership: types.ObjectOwnershipBucketOwnerPreferred,
				},
			},
		},
	}, nil
}

func (t TestS3Client) GetBucketPolicy(ctx context.Context, params *s3.GetBucketPolicyInput, optFns ...func(*s3.Options)) (*s3.GetBucketPolicyOutput, error) {
	return &s3.GetBucketPolicyOutput{
		Policy: adapters.PtrString("policy"),
	}, nil
}

func (t TestS3Client) GetBucketPolicyStatus(ctx context.Context, params *s3.GetBucketPolicyStatusInput, optFns ...func(*s3.Options)) (*s3.GetBucketPolicyStatusOutput, error) {
	return &s3.GetBucketPolicyStatusOutput{
		PolicyStatus: &types.PolicyStatus{
			IsPublic: adapters.PtrBool(true),
		},
	}, nil
}

func (t TestS3Client) GetBucketReplication(ctx context.Context, params *s3.GetBucketReplicationInput, optFns ...func(*s3.Options)) (*s3.GetBucketReplicationOutput, error) {
	return &s3.GetBucketReplicationOutput{
		ReplicationConfiguration: &types.ReplicationConfiguration{
			Role: adapters.PtrString("role"),
			Rules: []types.ReplicationRule{
				{
					Destination: &types.Destination{
						Bucket: adapters.PtrString("bucket"),
						AccessControlTranslation: &types.AccessControlTranslation{
							Owner: types.OwnerOverrideDestination,
						},
						Account: adapters.PtrString("account"),
						EncryptionConfiguration: &types.EncryptionConfiguration{
							ReplicaKmsKeyID: adapters.PtrString("keyId"),
						},
						Metrics: &types.Metrics{
							Status: types.MetricsStatusEnabled,
							EventThreshold: &types.ReplicationTimeValue{
								Minutes: adapters.PtrInt32(1),
							},
						},
						ReplicationTime: &types.ReplicationTime{
							Status: types.ReplicationTimeStatusEnabled,
							Time: &types.ReplicationTimeValue{
								Minutes: adapters.PtrInt32(1),
							},
						},
						StorageClass: types.StorageClassGlacier,
					},
				},
			},
		},
	}, nil
}

func (t TestS3Client) GetBucketRequestPayment(ctx context.Context, params *s3.GetBucketRequestPaymentInput, optFns ...func(*s3.Options)) (*s3.GetBucketRequestPaymentOutput, error) {
	return &s3.GetBucketRequestPaymentOutput{
		Payer: types.PayerRequester,
	}, nil
}

func (t TestS3Client) GetBucketTagging(ctx context.Context, params *s3.GetBucketTaggingInput, optFns ...func(*s3.Options)) (*s3.GetBucketTaggingOutput, error) {
	return &s3.GetBucketTaggingOutput{
		TagSet: []types.Tag{},
	}, nil
}

func (t TestS3Client) GetBucketVersioning(ctx context.Context, params *s3.GetBucketVersioningInput, optFns ...func(*s3.Options)) (*s3.GetBucketVersioningOutput, error) {
	return &s3.GetBucketVersioningOutput{
		MFADelete: types.MFADeleteStatusEnabled,
		Status:    types.BucketVersioningStatusSuspended,
	}, nil
}

func (t TestS3Client) GetBucketWebsite(ctx context.Context, params *s3.GetBucketWebsiteInput, optFns ...func(*s3.Options)) (*s3.GetBucketWebsiteOutput, error) {
	return &s3.GetBucketWebsiteOutput{
		ErrorDocument: &types.ErrorDocument{
			Key: adapters.PtrString("key"),
		},
		IndexDocument: &types.IndexDocument{
			Suffix: adapters.PtrString("html"),
		},
		RedirectAllRequestsTo: &types.RedirectAllRequestsTo{
			HostName: adapters.PtrString("hostname"),
			Protocol: types.ProtocolHttps,
		},
		RoutingRules: []types.RoutingRule{
			{
				Redirect: &types.Redirect{
					HostName:             adapters.PtrString("hostname"),
					HttpRedirectCode:     adapters.PtrString("303"),
					Protocol:             types.ProtocolHttp,
					ReplaceKeyPrefixWith: adapters.PtrString("pre"),
					ReplaceKeyWith:       adapters.PtrString("key"),
				},
			},
		},
	}, nil
}

type TestS3FailClient struct{}

func (t TestS3FailClient) ListBuckets(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error) {
	return nil, errors.New("failed to list buckets")
}

func (t TestS3FailClient) GetBucketAcl(ctx context.Context, params *s3.GetBucketAclInput, optFns ...func(*s3.Options)) (*s3.GetBucketAclOutput, error) {
	return nil, errors.New("failed to get bucket ACL")
}
func (t TestS3FailClient) GetBucketAnalyticsConfiguration(ctx context.Context, params *s3.GetBucketAnalyticsConfigurationInput, optFns ...func(*s3.Options)) (*s3.GetBucketAnalyticsConfigurationOutput, error) {
	return nil, errors.New("failed to get bucket ACL")
}

func (t TestS3FailClient) GetBucketCors(ctx context.Context, params *s3.GetBucketCorsInput, optFns ...func(*s3.Options)) (*s3.GetBucketCorsOutput, error) {
	return nil, errors.New("failed to get bucket CORS")
}

func (t TestS3FailClient) GetBucketEncryption(ctx context.Context, params *s3.GetBucketEncryptionInput, optFns ...func(*s3.Options)) (*s3.GetBucketEncryptionOutput, error) {
	return nil, errors.New("failed to get bucket CORS")
}

func (t TestS3FailClient) GetBucketIntelligentTieringConfiguration(ctx context.Context, params *s3.GetBucketIntelligentTieringConfigurationInput, optFns ...func(*s3.Options)) (*s3.GetBucketIntelligentTieringConfigurationOutput, error) {
	return nil, errors.New("failed to get bucket CORS")
}

func (t TestS3FailClient) GetBucketInventoryConfiguration(ctx context.Context, params *s3.GetBucketInventoryConfigurationInput, optFns ...func(*s3.Options)) (*s3.GetBucketInventoryConfigurationOutput, error) {
	return nil, errors.New("failed to get bucket CORS")
}

func (t TestS3FailClient) GetBucketLifecycleConfiguration(ctx context.Context, params *s3.GetBucketLifecycleConfigurationInput, optFns ...func(*s3.Options)) (*s3.GetBucketLifecycleConfigurationOutput, error) {
	return nil, errors.New("failed to get bucket lifecycle configuration")
}

func (t TestS3FailClient) GetBucketLocation(ctx context.Context, params *s3.GetBucketLocationInput, optFns ...func(*s3.Options)) (*s3.GetBucketLocationOutput, error) {
	return nil, errors.New("failed to get bucket location")
}

func (t TestS3FailClient) GetBucketLogging(ctx context.Context, params *s3.GetBucketLoggingInput, optFns ...func(*s3.Options)) (*s3.GetBucketLoggingOutput, error) {
	return nil, errors.New("failed to get bucket logging")
}

func (t TestS3FailClient) GetBucketMetricsConfiguration(ctx context.Context, params *s3.GetBucketMetricsConfigurationInput, optFns ...func(*s3.Options)) (*s3.GetBucketMetricsConfigurationOutput, error) {
	return nil, errors.New("failed to get bucket logging")
}

func (t TestS3FailClient) GetBucketNotificationConfiguration(ctx context.Context, params *s3.GetBucketNotificationConfigurationInput, optFns ...func(*s3.Options)) (*s3.GetBucketNotificationConfigurationOutput, error) {
	return nil, errors.New("failed to get bucket notification configuration")
}

func (t TestS3FailClient) GetBucketOwnershipControls(ctx context.Context, params *s3.GetBucketOwnershipControlsInput, optFns ...func(*s3.Options)) (*s3.GetBucketOwnershipControlsOutput, error) {
	return nil, errors.New("failed to get bucket policy")
}

func (t TestS3FailClient) GetBucketPolicy(ctx context.Context, params *s3.GetBucketPolicyInput, optFns ...func(*s3.Options)) (*s3.GetBucketPolicyOutput, error) {
	return nil, errors.New("failed to get bucket policy")
}

func (t TestS3FailClient) GetBucketPolicyStatus(ctx context.Context, params *s3.GetBucketPolicyStatusInput, optFns ...func(*s3.Options)) (*s3.GetBucketPolicyStatusOutput, error) {
	return nil, errors.New("failed to get bucket policy")
}

func (t TestS3FailClient) GetBucketReplication(ctx context.Context, params *s3.GetBucketReplicationInput, optFns ...func(*s3.Options)) (*s3.GetBucketReplicationOutput, error) {
	return nil, errors.New("failed to get bucket replication")
}

func (t TestS3FailClient) GetBucketRequestPayment(ctx context.Context, params *s3.GetBucketRequestPaymentInput, optFns ...func(*s3.Options)) (*s3.GetBucketRequestPaymentOutput, error) {
	return nil, errors.New("failed to get bucket request payment")
}

func (t TestS3FailClient) GetBucketTagging(ctx context.Context, params *s3.GetBucketTaggingInput, optFns ...func(*s3.Options)) (*s3.GetBucketTaggingOutput, error) {
	return nil, errors.New("failed to get bucket tagging")
}

func (t TestS3FailClient) GetBucketVersioning(ctx context.Context, params *s3.GetBucketVersioningInput, optFns ...func(*s3.Options)) (*s3.GetBucketVersioningOutput, error) {
	return nil, errors.New("failed to get bucket versioning")
}

func (t TestS3FailClient) GetBucketWebsite(ctx context.Context, params *s3.GetBucketWebsiteInput, optFns ...func(*s3.Options)) (*s3.GetBucketWebsiteOutput, error) {
	return nil, errors.New("failed to get bucket website")
}

func (t TestS3FailClient) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	return nil, errors.New("failed to get object")
}

func (t TestS3FailClient) HeadBucket(ctx context.Context, params *s3.HeadBucketInput, optFns ...func(*s3.Options)) (*s3.HeadBucketOutput, error) {
	return nil, errors.New("failed to head bucket")
}

func (t TestS3FailClient) HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
	return nil, errors.New("failed to head object")
}

func (t TestS3FailClient) PutBucketAcl(ctx context.Context, params *s3.PutBucketAclInput, optFns ...func(*s3.Options)) (*s3.PutBucketAclOutput, error) {
	return nil, errors.New("failed to put bucket ACL")
}

func (t TestS3FailClient) PutBucketCors(ctx context.Context, params *s3.PutBucketCorsInput, optFns ...func(*s3.Options)) (*s3.PutBucketCorsOutput, error) {
	return nil, errors.New("failed to put bucket CORS")
}

func (t TestS3FailClient) PutBucketLifecycleConfiguration(ctx context.Context, params *s3.PutBucketLifecycleConfigurationInput, optFns ...func(*s3.Options)) (*s3.PutBucketLifecycleConfigurationOutput, error) {
	return nil, errors.New("failed to put bucket lifecycle configuration")
}

func (t TestS3FailClient) PutBucketLogging(ctx context.Context, params *s3.PutBucketLoggingInput, optFns ...func(*s3.Options)) (*s3.PutBucketLoggingOutput, error) {
	return nil, errors.New("failed to put bucket logging")
}

func (t TestS3FailClient) PutBucketNotificationConfiguration(ctx context.Context, params *s3.PutBucketNotificationConfigurationInput, optFns ...func(*s3.Options)) (*s3.PutBucketNotificationConfigurationOutput, error) {
	return nil, errors.New("failed to put bucket notification configuration")
}

func (t TestS3FailClient) PutBucketPolicy(ctx context.Context, params *s3.PutBucketPolicyInput, optFns ...func(*s3.Options)) (*s3.PutBucketPolicyOutput, error) {
	return nil, errors.New("failed to put bucket policy")
}

func (t TestS3FailClient) PutBucketReplication(ctx context.Context, params *s3.PutBucketReplicationInput, optFns ...func(*s3.Options)) (*s3.PutBucketReplicationOutput, error) {
	return nil, errors.New("failed to put bucket replication")
}

func (t TestS3FailClient) PutBucketRequestPayment(ctx context.Context, params *s3.PutBucketRequestPaymentInput, optFns ...func(*s3.Options)) (*s3.PutBucketRequestPaymentOutput, error) {
	return nil, errors.New("failed to put bucket request payment")
}

func (t TestS3FailClient) PutBucketTagging(ctx context.Context, params *s3.PutBucketTaggingInput, optFns ...func(*s3.Options)) (*s3.PutBucketTaggingOutput, error) {
	return nil, errors.New("failed to put bucket tagging")
}

func (t TestS3FailClient) PutBucketVersioning(ctx context.Context, params *s3.PutBucketVersioningInput, optFns ...func(*s3.Options)) (*s3.PutBucketVersioningOutput, error) {
	return nil, errors.New("failed to put bucket versioning")
}

func (t TestS3FailClient) PutBucketWebsite(ctx context.Context, params *s3.PutBucketWebsiteInput, optFns ...func(*s3.Options)) (*s3.PutBucketWebsiteOutput, error) {
	return nil, errors.New("failed to put bucket website")
}

func (t TestS3FailClient) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	return nil, errors.New("failed to put object")
}

func TestNewS3Adapter(t *testing.T) {
	config, account, _ := adapters.GetAutoConfig(t)

	adapter := NewS3Adapter(config, account)

	test := adapters.E2ETest{
		Adapter: adapter,
		Timeout: 10 * time.Second,
	}

	test.Run(t)
}
