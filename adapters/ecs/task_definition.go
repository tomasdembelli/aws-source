package ecs

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/overmindtech/aws-source/adapters"
	"github.com/overmindtech/sdp-go"
)

// TaskDefinitionIncludeFields Fields that we want included by default
var TaskDefinitionIncludeFields = []types.TaskDefinitionField{
	types.TaskDefinitionFieldTags,
}

func taskDefinitionGetFunc(ctx context.Context, client ECSClient, scope string, input *ecs.DescribeTaskDefinitionInput) (*sdp.Item, error) {
	out, err := client.DescribeTaskDefinition(ctx, input)

	if err != nil {
		return nil, err
	}

	if out.TaskDefinition == nil {
		return nil, errors.New("task definition is nil")
	}

	td := out.TaskDefinition

	attributes, err := adapters.ToAttributesWithExclude(td)

	if err != nil {
		return nil, err
	}

	// Set a custom attribute that we will use for a unique attribute in the
	// format: {family}:{revision}
	if td.Family == nil {
		return nil, errors.New("task definition family was nil")
	}

	item := sdp.Item{
		Type:            "ecs-task-definition",
		UniqueAttribute: "Family",
		Attributes:      attributes,
		Scope:           scope,
		Tags:            tagsToMap(out.Tags),
	}

	switch td.Status {
	case types.TaskDefinitionStatusActive:
		item.Health = sdp.Health_HEALTH_OK.Enum()
	case types.TaskDefinitionStatusInactive:
		item.Health = nil
	case types.TaskDefinitionStatusDeleteInProgress:
		item.Health = sdp.Health_HEALTH_WARNING.Enum()
	}

	var a *adapters.ARN
	var link *sdp.LinkedItemQuery

	for _, cd := range td.ContainerDefinitions {
		for _, secret := range cd.Secrets {
			link = getSecretLinkedItem(secret)

			if link != nil {
				item.LinkedItemQueries = append(item.LinkedItemQueries, link)
			}
		}

		if cd.LogConfiguration != nil {
			for _, secret := range cd.LogConfiguration.SecretOptions {
				link = getSecretLinkedItem(secret)

				if link != nil {
					item.LinkedItemQueries = append(item.LinkedItemQueries, link)
				}
			}
		}

		newQueries, err := sdp.ExtractLinksFrom(cd.Environment)
		if err == nil {
			item.LinkedItemQueries = append(item.LinkedItemQueries, newQueries...)
		}
	}

	if td.ExecutionRoleArn != nil {
		if a, err = adapters.ParseARN(*td.ExecutionRoleArn); err == nil {
			// +overmind:link iam-role
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "iam-role",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *td.ExecutionRoleArn,
					Scope:  adapters.FormatScope(a.AccountID, a.Region),
				},
				BlastPropagation: &sdp.BlastPropagation{
					// The role can affect the task definition
					In: true,
					// The task definition can't affect the role
					Out: false,
				},
			})
		}
	}

	if td.TaskRoleArn != nil {
		if a, err = adapters.ParseARN(*td.TaskRoleArn); err == nil {
			// +overmind:link iam-role
			item.LinkedItemQueries = append(item.LinkedItemQueries, &sdp.LinkedItemQuery{
				Query: &sdp.Query{
					Type:   "iam-role",
					Method: sdp.QueryMethod_SEARCH,
					Query:  *td.TaskRoleArn,
					Scope:  adapters.FormatScope(a.AccountID, a.Region),
				},
				BlastPropagation: &sdp.BlastPropagation{
					// The role can affect the task definition
					In: true,
					// The task definition can't affect the role
					Out: false,
				},
			})
		}
	}

	return &item, nil
}

// getSecretLinkedItem Converts a `types.Secret` to the linked item that the
// secret is related to, if relevant
func getSecretLinkedItem(secret types.Secret) *sdp.LinkedItemQuery {
	if secret.ValueFrom != nil {
		if a, err := adapters.ParseARN(*secret.ValueFrom); err == nil {
			// The secret can refer to either something from secrets
			// manager or SSN, so handle this
			secretScope := adapters.FormatScope(a.AccountID, a.Region)

			switch a.Service {
			case "secretsmanager":
				// +overmind:link secretsmanager-secret
				return &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "secretsmanager-secret",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *secret.ValueFrom,
						Scope:  secretScope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						// The secret can affect the task definition
						In: true,
						// The task definition can't affect the secret
						Out: false,
					},
				}
			case "ssm":
				// +overmind:link ssm-parameter
				return &sdp.LinkedItemQuery{
					Query: &sdp.Query{
						Type:   "ssm-parameter",
						Method: sdp.QueryMethod_SEARCH,
						Query:  *secret.ValueFrom,
						Scope:  secretScope,
					},
					BlastPropagation: &sdp.BlastPropagation{
						// The secret can affect the task definition
						In: true,
						// The task definition can't affect the secret
						Out: false,
					},
				}
			}
		}
	}

	return nil
}

//go:generate docgen ../../docs-data
// +overmind:type ecs-task-definition
// +overmind:descriptiveType Task Definition
// +overmind:get Get a task definition by revision name ({family}:{revision})
// +overmind:list List all task definitions
// +overmind:search Search for task definitions by ARN
// +overmind:group AWS
// +overmind:terraform:queryMap aws_ecs_task_definition.family

func NewTaskDefinitionAdapter(client ECSClient, accountID string, region string) *adapters.AlwaysGetAdapter[*ecs.ListTaskDefinitionsInput, *ecs.ListTaskDefinitionsOutput, *ecs.DescribeTaskDefinitionInput, *ecs.DescribeTaskDefinitionOutput, ECSClient, *ecs.Options] {
	return &adapters.AlwaysGetAdapter[*ecs.ListTaskDefinitionsInput, *ecs.ListTaskDefinitionsOutput, *ecs.DescribeTaskDefinitionInput, *ecs.DescribeTaskDefinitionOutput, ECSClient, *ecs.Options]{
		ItemType:        "ecs-task-definition",
		Client:          client,
		AccountID:       accountID,
		Region:          region,
		GetFunc:         taskDefinitionGetFunc,
		ListInput:       &ecs.ListTaskDefinitionsInput{},
		AdapterMetadata: TaskDefinitionMetadata(),
		GetInputMapper: func(scope, query string) *ecs.DescribeTaskDefinitionInput {
			// AWS actually supports "family:revision" format as an input here
			// so we can just push it in directly
			return &ecs.DescribeTaskDefinitionInput{
				TaskDefinition: adapters.PtrString(query),
			}
		},
		ListFuncPaginatorBuilder: func(client ECSClient, input *ecs.ListTaskDefinitionsInput) adapters.Paginator[*ecs.ListTaskDefinitionsOutput, *ecs.Options] {
			return ecs.NewListTaskDefinitionsPaginator(client, input)
		},
		ListFuncOutputMapper: func(output *ecs.ListTaskDefinitionsOutput, input *ecs.ListTaskDefinitionsInput) ([]*ecs.DescribeTaskDefinitionInput, error) {
			getInputs := make([](*ecs.DescribeTaskDefinitionInput), 0)

			for _, arn := range output.TaskDefinitionArns {
				if a, err := adapters.ParseARN(arn); err == nil {
					getInputs = append(getInputs, &ecs.DescribeTaskDefinitionInput{
						TaskDefinition: adapters.PtrString(a.ResourceID()),
					})
				}
			}

			return getInputs, nil
		},
	}
}

func TaskDefinitionMetadata() sdp.AdapterMetadata {
	return sdp.AdapterMetadata{
		Type:            "ecs-task-definition",
		DescriptiveName: "Task Definition",
		SupportedQueryMethods: &sdp.AdapterSupportedQueryMethods{
			Get:               true,
			List:              true,
			Search:            true,
			GetDescription:    "Get a task definition by revision name ({family}:{revision})",
			ListDescription:   "List all task definitions",
			SearchDescription: "Search for task definitions by ARN",
		},
		TerraformMappings: []*sdp.TerraformMapping{
			{TerraformQueryMap: "aws_ecs_task_definition.family"},
		},
		PotentialLinks: []string{"iam-role", "secretsmanager-secret", "ssm-parameter"},
		Category:       sdp.AdapterCategory_ADAPTER_CATEGORY_COMPUTE_APPLICATION,
	}
}
