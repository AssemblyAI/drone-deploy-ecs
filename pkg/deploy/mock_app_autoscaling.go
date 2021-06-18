package deploy

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling"
	astypes "github.com/aws/aws-sdk-go-v2/service/applicationautoscaling/types"
	"github.com/aws/smithy-go/middleware"
)

type MockAppAutoscalingClient struct {
	TestingT     *testing.T
	WantError    bool
	TargetExists bool
}

func (c MockAppAutoscalingClient) DescribeScalableTargets(ctx context.Context, params *applicationautoscaling.DescribeScalableTargetsInput, optFns ...func(*applicationautoscaling.Options)) (*applicationautoscaling.DescribeScalableTargetsOutput, error) {
	if c.WantError {
		return nil, errors.New("error")
	}

	if !c.TargetExists {
		return &applicationautoscaling.DescribeScalableTargetsOutput{ScalableTargets: []astypes.ScalableTarget{}}, nil
	}

	out := applicationautoscaling.DescribeScalableTargetsOutput{
		NextToken: new(string),
		ScalableTargets: []astypes.ScalableTarget{
			{
				MaxCapacity:       aws.Int32(20),
				MinCapacity:       aws.Int32(4),
				ResourceId:        new(string),
				RoleARN:           new(string),
				ScalableDimension: astypes.ScalableDimensionECSServiceDesiredCount,
				ServiceNamespace:  astypes.ServiceNamespaceEcs,
				SuspendedState:    &astypes.SuspendedState{},
			},
		},
		ResultMetadata: middleware.Metadata{},
	}

	return &out, nil
}

func (c MockAppAutoscalingClient) RegisterScalableTarget(ctx context.Context, params *applicationautoscaling.RegisterScalableTargetInput, optFns ...func(*applicationautoscaling.Options)) (*applicationautoscaling.RegisterScalableTargetOutput, error) {
	if c.WantError {
		return nil, errors.New("error")
	}

	out := applicationautoscaling.RegisterScalableTargetOutput{
		ResultMetadata: middleware.Metadata{},
	}

	return &out, nil
}
