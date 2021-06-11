package deploy

import (
	"context"
	"fmt"

	"github.com/assemblyai/drone-deploy-ecs/pkg/types"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling"
	astypes "github.com/aws/aws-sdk-go-v2/service/applicationautoscaling/types"
)

// TODO add logging
func AppAutoscalingTargetExists() bool {

	return true
}

func GetServiceMaxCount(ctx context.Context, c types.AppAutoscalingClient, cluster string, service string) (int32, error) {
	r, err := getScalableTarget(ctx, c, cluster, service)

	if err != nil {
		return 0, err
	}

	return *r.MaxCapacity, nil
}

func ScaleDown(c types.ECSClient, desiredCount int32, minCount int32, maxCount int32) {
	// Set max, desired, min to 0
}

//
func ScaleUp(c types.ECSClient, count int32, maxCount int32) {
	// Set max count to maxCount
}

func setAppAutoscalingMaxCount(ctx context.Context, c types.AppAutoscalingClient, service string, cluster string, maxCount int32) error {

	p := applicationautoscaling.RegisterScalableTargetInput{
		ResourceId:  aws.String(fmt.Sprintf("service/%s/%s", cluster, service)),
		MaxCapacity: &maxCount,
	}

	_, err := c.RegisterScalableTarget(ctx, &p)

	if err != nil {
		return err
	}

	return nil
}

func setserviceDesiredCount() {

}

func getScalableTarget(ctx context.Context, c types.AppAutoscalingClient, cluster string, service string) (*astypes.ScalableTarget, error) {
	resourceID := fmt.Sprintf("service/%s/%s", cluster, service)

	p := applicationautoscaling.DescribeScalableTargetsInput{
		ServiceNamespace: "ecs",
		ResourceIds:      []string{resourceID},
	}

	r, err := c.DescribeScalableTargets(ctx, &p, nil)

	if err != nil {
		return nil, err
	}

	return &r.ScalableTargets[0], nil
}
