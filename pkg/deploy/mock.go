package deploy

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"gotest.tools/assert"
)

const (
	testTDARN string = "arn:aws:ecs:us-west-2:123456789012:task-definition/amazon-ecs-sample:1"
)

type MockECSClient struct {
	DeploymentState ecstypes.DeploymentRolloutState
	TestingT        *testing.T
	WantError       bool
}

func (c MockECSClient) DescribeTaskDefinition(ctx context.Context, params *ecs.DescribeTaskDefinitionInput, optFns ...func(*ecs.Options)) (*ecs.DescribeTaskDefinitionOutput, error) {

	if c.WantError {
		return nil, errors.New("error")
	}
	td := ecstypes.TaskDefinition{
		ContainerDefinitions: []ecstypes.ContainerDefinition{
			{
				Name:  aws.String("app"),
				Image: aws.String("some/image:1.0"),
			},
			{
				Name:  aws.String("sidecar"),
				Image: aws.String("datadog/agent:7"),
			},
		},
	}

	out := ecs.DescribeTaskDefinitionOutput{
		TaskDefinition: &td,
	}

	return &out, nil
}

func (c MockECSClient) RegisterTaskDefinition(ctx context.Context, params *ecs.RegisterTaskDefinitionInput, optFns ...func(*ecs.Options)) (*ecs.RegisterTaskDefinitionOutput, error) {
	out := ecs.RegisterTaskDefinitionOutput{
		Tags: []ecstypes.Tag{},
		TaskDefinition: &ecstypes.TaskDefinition{
			Compatibilities:         []ecstypes.Compatibility{},
			ContainerDefinitions:    params.ContainerDefinitions,
			Cpu:                     params.Cpu,
			EphemeralStorage:        params.EphemeralStorage,
			ExecutionRoleArn:        params.ExecutionRoleArn,
			Family:                  params.Family,
			InferenceAccelerators:   params.InferenceAccelerators,
			IpcMode:                 params.IpcMode,
			Memory:                  params.Memory,
			NetworkMode:             params.NetworkMode,
			PidMode:                 params.PidMode,
			PlacementConstraints:    params.PlacementConstraints,
			ProxyConfiguration:      params.ProxyConfiguration,
			RequiresCompatibilities: params.RequiresCompatibilities,
			Status:                  ecstypes.TaskDefinitionStatus(c.DeploymentState),
			TaskDefinitionArn:       aws.String(testTDARN),
			TaskRoleArn:             params.TaskRoleArn,
			Volumes:                 params.Volumes,
		},
	}

	return &out, nil
}

func (c MockECSClient) DescribeServices(ctx context.Context, params *ecs.DescribeServicesInput, optFns ...func(*ecs.Options)) (*ecs.DescribeServicesOutput, error) {

	assert.Equal(c.TestingT, *params.Cluster, "test-cluster")
	assert.Equal(c.TestingT, params.Services[0], "test-service")

	d := []ecstypes.Deployment{
		{
			DesiredCount: 2,
			Id:           aws.String("some-deployment-id"),
			RolloutState: c.DeploymentState,
			Status:       aws.String("PRIMARY"),
		},
	}

	s := []ecstypes.Service{
		{
			ServiceName:    aws.String("ci-cluster"),
			Status:         aws.String("ACTIVE"),
			Deployments:    d,
			TaskDefinition: aws.String(testTDARN),
		},
	}

	out := ecs.DescribeServicesOutput{
		Failures: []ecstypes.Failure{},
		Services: s,
	}

	return &out, nil
}

func (c MockECSClient) UpdateService(ctx context.Context, params *ecs.UpdateServiceInput, optFns ...func(*ecs.Options)) (*ecs.UpdateServiceOutput, error) {

	if c.WantError {
		return nil, errors.New("error")
	}

	out := ecs.UpdateServiceOutput{
		Service: &ecstypes.Service{
			CapacityProviderStrategy: []ecstypes.CapacityProviderStrategyItem{},
			ClusterArn:               new(string),
			CreatedAt:                &time.Time{},
			CreatedBy:                new(string),
			DeploymentConfiguration:  &ecstypes.DeploymentConfiguration{},
			DeploymentController:     &ecstypes.DeploymentController{},
			Deployments: []ecstypes.Deployment{
				{
					CapacityProviderStrategy: []ecstypes.CapacityProviderStrategyItem{},
					CreatedAt:                &time.Time{},
					DesiredCount:             0,
					FailedTasks:              0,
					Id:                       aws.String("test-deployment"),
					LaunchType:               "",
					NetworkConfiguration:     &ecstypes.NetworkConfiguration{},
					PendingCount:             0,
					PlatformVersion:          new(string),
					RolloutState:             "",
					RolloutStateReason:       new(string),
					RunningCount:             0,
					Status:                   new(string),
					TaskDefinition:           new(string),
					UpdatedAt:                &time.Time{},
				},
			},
			DesiredCount:                  0,
			EnableECSManagedTags:          false,
			EnableExecuteCommand:          false,
			Events:                        []ecstypes.ServiceEvent{},
			HealthCheckGracePeriodSeconds: new(int32),
			LaunchType:                    "",
			LoadBalancers:                 []ecstypes.LoadBalancer{},
			NetworkConfiguration:          &ecstypes.NetworkConfiguration{},
			PendingCount:                  0,
			PlacementConstraints:          []ecstypes.PlacementConstraint{},
			PlacementStrategy:             []ecstypes.PlacementStrategy{},
			PlatformVersion:               new(string),
			PropagateTags:                 "",
			RoleArn:                       new(string),
			RunningCount:                  0,
			SchedulingStrategy:            "",
			ServiceArn:                    new(string),
			ServiceName:                   new(string),
			ServiceRegistries:             []ecstypes.ServiceRegistry{},
			Status:                        new(string),
			Tags:                          []ecstypes.Tag{},
			TaskDefinition:                new(string),
			TaskSets:                      []ecstypes.TaskSet{},
		},
	}

	return &out, nil
}
