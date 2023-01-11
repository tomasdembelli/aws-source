package lambda

import (
	"context"
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/overmindtech/aws-source/sources"
	"github.com/overmindtech/sdp-go"
)

type FunctionDetails struct {
	Code               *types.FunctionCodeLocation
	Concurrency        *types.Concurrency
	Configuration      *types.FunctionConfiguration
	UrlConfigs         []*types.FunctionUrlConfig
	EventInvokeConfigs []*types.FunctionEventInvokeConfig
	Tags               map[string]string
}

// FunctionGetFunc Gets the details of a specific lambda function
func FunctionGetFunc(ctx context.Context, client LambdaClient, scope string, input *lambda.GetFunctionInput) (*sdp.Item, error) {
	out, err := client.GetFunction(ctx, input)

	if err != nil {
		return nil, err
	}

	if out.Configuration == nil {
		return nil, errors.New("function has nil configuration")
	}

	if out.Configuration.FunctionName == nil {
		return nil, errors.New("function has empty name")
	}

	function := FunctionDetails{
		Code:          out.Code,
		Concurrency:   out.Concurrency,
		Configuration: out.Configuration,
		Tags:          out.Tags,
	}

	// Get details of all URL configs
	urlConfigs := lambda.NewListFunctionUrlConfigsPaginator(client, &lambda.ListFunctionUrlConfigsInput{
		FunctionName: out.Configuration.FunctionName,
	})

	var urlOut *lambda.ListFunctionUrlConfigsOutput

	for urlConfigs.HasMorePages() {
		urlOut, err = urlConfigs.NextPage(ctx)

		if err != nil {
			continue
		}

		for _, config := range urlOut.FunctionUrlConfigs {
			function.UrlConfigs = append(function.UrlConfigs, &config)
		}
	}

	// Get details of event configs
	eventConfigs := lambda.NewListFunctionEventInvokeConfigsPaginator(client, &lambda.ListFunctionEventInvokeConfigsInput{
		FunctionName: out.Configuration.FunctionName,
	})

	var eventOut *lambda.ListFunctionEventInvokeConfigsOutput

	for eventConfigs.HasMorePages() {
		eventOut, err = eventConfigs.NextPage(ctx)

		if err != nil {
			continue
		}

		for _, event := range eventOut.FunctionEventInvokeConfigs {
			function.EventInvokeConfigs = append(function.EventInvokeConfigs, &event)
		}
	}

	attributes, err := sources.ToAttributesCase(function, "resultMetadata")

	if err != nil {
		return nil, err
	}

	err = attributes.Set("name", *out.Configuration.FunctionName)

	if err != nil {
		return nil, err
	}

	item := sdp.Item{
		Type:            "lambda-function",
		UniqueAttribute: "name",
		Attributes:      attributes,
		Scope:           scope,
	}

	if function.Code != nil {
		if function.Code.Location != nil {
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Type:   "http",
				Method: sdp.RequestMethod_GET,
				Query:  *function.Code.Location,
				Scope:  "global",
			})
		}

		if function.Code.ImageUri != nil {
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Type:   "http",
				Method: sdp.RequestMethod_GET,
				Query:  *function.Code.ImageUri,
				Scope:  "global",
			})
		}

		if function.Code.ResolvedImageUri != nil {
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Type:   "http",
				Method: sdp.RequestMethod_GET,
				Query:  *function.Code.ResolvedImageUri,
				Scope:  "global",
			})
		}
	}

	var a *sources.ARN

	if function.Configuration != nil {
		if function.Configuration.Role != nil {
			if a, err = sources.ParseARN(*function.Configuration.Role); err == nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "iam-role",
					Method: sdp.RequestMethod_SEARCH,
					Query:  *function.Configuration.Role,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				})
			}
		}

		if function.Configuration.DeadLetterConfig != nil {
			if function.Configuration.DeadLetterConfig.TargetArn != nil {
				if req, err := GetEventLinkedItem(*function.Configuration.DeadLetterConfig.TargetArn); err == nil {
					item.LinkedItemRequests = append(item.LinkedItemRequests, req)
				}
			}
		}

		for _, fsConfig := range function.Configuration.FileSystemConfigs {
			if fsConfig.Arn != nil {
				if a, err = sources.ParseARN(*fsConfig.Arn); err == nil {
					item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
						Type:   "efs-access-point",
						Method: sdp.RequestMethod_SEARCH,
						Query:  *fsConfig.Arn,
						Scope:  sources.FormatScope(a.AccountID, a.Region),
					})
				}
			}
		}

		if function.Configuration.KMSKeyArn != nil {
			if a, err = sources.ParseARN(*function.Configuration.KMSKeyArn); err == nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "kms-key",
					Method: sdp.RequestMethod_SEARCH,
					Query:  *function.Configuration.KMSKeyArn,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				})
			}
		}

		for _, layer := range function.Configuration.Layers {
			if layer.Arn != nil {
				if a, err = sources.ParseARN(*layer.Arn); err == nil {
					// Strip the leading "layer:"
					name := strings.TrimPrefix(a.Resource, "layer:")
					item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
						Type:   "lambda-layer-version",
						Method: sdp.RequestMethod_GET,
						Query:  name,
						Scope:  sources.FormatScope(a.AccountID, a.Region),
					})
				}
			}

			if layer.SigningJobArn != nil {
				if a, err = sources.ParseARN(*layer.SigningJobArn); err == nil {
					item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
						Type:   "signer-signing-job",
						Method: sdp.RequestMethod_SEARCH,
						Query:  *layer.SigningJobArn,
						Scope:  sources.FormatScope(a.AccountID, a.Region),
					})
				}
			}

			if layer.SigningProfileVersionArn != nil {
				if a, err = sources.ParseARN(*layer.SigningProfileVersionArn); err == nil {
					item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
						Type:   "signer-signing-profile",
						Method: sdp.RequestMethod_SEARCH,
						Query:  *layer.SigningProfileVersionArn,
						Scope:  sources.FormatScope(a.AccountID, a.Region),
					})
				}
			}
		}

		if function.Configuration.MasterArn != nil {
			if a, err = sources.ParseARN(*function.Configuration.MasterArn); err == nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "lambda-function",
					Method: sdp.RequestMethod_SEARCH,
					Query:  *function.Configuration.MasterArn,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				})
			}
		}

		if function.Configuration.SigningJobArn != nil {
			if a, err = sources.ParseARN(*function.Configuration.SigningJobArn); err == nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "signer-signing-job",
					Method: sdp.RequestMethod_SEARCH,
					Query:  *function.Configuration.SigningJobArn,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				})
			}
		}

		if function.Configuration.SigningProfileVersionArn != nil {
			if a, err = sources.ParseARN(*function.Configuration.SigningProfileVersionArn); err == nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "signer-signing-profile",
					Method: sdp.RequestMethod_SEARCH,
					Query:  *function.Configuration.SigningProfileVersionArn,
					Scope:  sources.FormatScope(a.AccountID, a.Region),
				})
			}
		}

		if function.Configuration.VpcConfig != nil {
			for _, id := range function.Configuration.VpcConfig.SecurityGroupIds {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "ec2-security-group",
					Method: sdp.RequestMethod_GET,
					Query:  id,
					Scope:  scope,
				})
			}

			for _, id := range function.Configuration.VpcConfig.SubnetIds {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "ec2-subnet",
					Method: sdp.RequestMethod_GET,
					Query:  id,
					Scope:  scope,
				})
			}

			if function.Configuration.VpcConfig.VpcId != nil {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Type:   "ec2-vpc",
					Method: sdp.RequestMethod_GET,
					Query:  *function.Configuration.VpcConfig.VpcId,
					Scope:  scope,
				})
			}
		}
	}

	for _, config := range function.UrlConfigs {
		if config.FunctionUrl != nil {
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Type:   "http",
				Method: sdp.RequestMethod_GET,
				Query:  *config.FunctionUrl,
				Scope:  "global",
			})
		}
	}

	for _, config := range function.EventInvokeConfigs {
		if config.DestinationConfig != nil {
			if config.DestinationConfig.OnFailure != nil {
				if config.DestinationConfig.OnFailure.Destination != nil {
					lir, err := GetEventLinkedItem(*config.DestinationConfig.OnFailure.Destination)

					if err == nil {
						item.LinkedItemRequests = append(item.LinkedItemRequests, lir)
					}
				}
			}

			if config.DestinationConfig.OnSuccess != nil {
				if config.DestinationConfig.OnSuccess.Destination != nil {
					lir, err := GetEventLinkedItem(*config.DestinationConfig.OnSuccess.Destination)

					if err == nil {
						item.LinkedItemRequests = append(item.LinkedItemRequests, lir)
					}

				}
			}
		}
	}

	return &item, nil
}

// GetEventLinkedItem Gets the linked item request for a given destination ARN
func GetEventLinkedItem(destinationARN string) (*sdp.ItemRequest, error) {
	parsed, err := sources.ParseARN(destinationARN)

	if err != nil {
		return nil, err
	}

	scope := sources.FormatScope(parsed.AccountID, parsed.Region)

	switch parsed.Service {
	case "sns":
		// In this case it's an SNS topic
		return &sdp.ItemRequest{
			Type:   "sns-topic",
			Method: sdp.RequestMethod_SEARCH,
			Query:  destinationARN,
			Scope:  scope,
		}, nil
	case "sqs":
		return &sdp.ItemRequest{
			Type:   "sqs-queue",
			Method: sdp.RequestMethod_SEARCH,
			Query:  destinationARN,
			Scope:  scope,
		}, nil
	case "lambda":
		return &sdp.ItemRequest{
			Type:   "lambda-function",
			Method: sdp.RequestMethod_SEARCH,
			Query:  destinationARN,
			Scope:  scope,
		}, nil
	case "events":
		return &sdp.ItemRequest{
			Type:   "events-event-bus",
			Method: sdp.RequestMethod_SEARCH,
			Query:  destinationARN,
			Scope:  scope,
		}, nil
	}

	return nil, errors.New("could not find matching request")
}

func NewFunctionSource(config aws.Config, accountID string, region string) *sources.AlwaysGetSource[*lambda.ListFunctionsInput, *lambda.ListFunctionsOutput, *lambda.GetFunctionInput, *lambda.GetFunctionOutput, LambdaClient, *lambda.Options] {
	return &sources.AlwaysGetSource[*lambda.ListFunctionsInput, *lambda.ListFunctionsOutput, *lambda.GetFunctionInput, *lambda.GetFunctionOutput, LambdaClient, *lambda.Options]{
		ItemType:  "lambda-function",
		Client:    lambda.NewFromConfig(config),
		AccountID: accountID,
		Region:    region,
		ListInput: &lambda.ListFunctionsInput{},
		GetFunc:   FunctionGetFunc,
		GetInputMapper: func(scope, query string) *lambda.GetFunctionInput {
			return &lambda.GetFunctionInput{
				FunctionName: &query,
			}
		},
		ListFuncPaginatorBuilder: func(client LambdaClient, input *lambda.ListFunctionsInput) sources.Paginator[*lambda.ListFunctionsOutput, *lambda.Options] {
			return lambda.NewListFunctionsPaginator(client, input)
		},
		ListFuncOutputMapper: func(output *lambda.ListFunctionsOutput, input *lambda.ListFunctionsInput) ([]*lambda.GetFunctionInput, error) {
			inputs := make([]*lambda.GetFunctionInput, len(output.Functions))

			for i, f := range output.Functions {
				inputs[i] = &lambda.GetFunctionInput{
					FunctionName: f.FunctionName,
				}
			}

			return inputs, nil
		},
	}
}