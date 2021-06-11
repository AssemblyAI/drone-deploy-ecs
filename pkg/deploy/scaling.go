package deploy

import (
	"context"
	"fmt"
	"log"

	"github.com/assemblyai/drone-deploy-ecs/pkg/types"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling"
	astypes "github.com/aws/aws-sdk-go-v2/service/applicationautoscaling/types"
)

func AppAutoscalingTargetExists(ctx context.Context, c types.AppAutoscalingClient, cluster string, service string) (bool, error) {
	_, err := getScalableTarget(ctx, c, cluster, service)

	if err == nil {
		return true, nil
	}

	if _, ok := err.(*ErrNoResults); ok {
		return false, nil
	} else {
		return false, err
	}

}

func GetServiceMinMaxCount(ctx context.Context, c types.AppAutoscalingClient, cluster string, service string) (int32, int32, error) {
	r, err := getScalableTarget(ctx, c, cluster, service)

	if err != nil {
		return 0, 0, err
	}

	return *r.MaxCapacity, *r.MinCapacity, nil
}

func (c DeployConfig) ScaleDown(desiredCount int32, minCount int32, maxCount int32, service string, serviceUsesAppAutoscaling bool) error {
	log.Println("Setting desired count to", desiredCount, "for service", service)
	err := setECSServiceDesiredCount(c.ECS, service, c.Cluster, desiredCount)

	if err != nil {
		return err
	}

	if serviceUsesAppAutoscaling {
		log.Println("Setting max count to", maxCount, "for service", service)
		return setAppAutoscalingCounts(context.Background(), c.AppAutoscaling, service, c.Cluster, maxCount, 0)
	}

	return nil
}

func (c DeployConfig) ScaleUp(desiredCount int32, minCount int32, maxCount int32, service string) error {
	if maxCount == -1 {
		log.Println("Setting desired count to", desiredCount, "for service", service)
		return setECSServiceDesiredCount(c.ECS, service, c.Cluster, desiredCount)
	} else {
		err := setAppAutoscalingCounts(context.Background(), c.AppAutoscaling, service, c.Cluster, maxCount, minCount)

		if err != nil {
			return err
		}

		err = setECSServiceDesiredCount(c.ECS, service, c.Cluster, desiredCount)

		return err
	}

}

func setAppAutoscalingCounts(
	ctx context.Context, c types.AppAutoscalingClient, service string, cluster string, maxCount int32, minCount int32) error {

	p := applicationautoscaling.RegisterScalableTargetInput{
		ResourceId:        aws.String(fmt.Sprintf("service/%s/%s", cluster, service)),
		ServiceNamespace:  astypes.ServiceNamespaceEcs,
		MaxCapacity:       &maxCount,
		MinCapacity:       &minCount,
		ScalableDimension: astypes.ScalableDimensionECSServiceDesiredCount,
	}

	_, err := c.RegisterScalableTarget(ctx, &p)

	if err != nil {
		return err
	}

	return nil
}

func getScalableTarget(ctx context.Context, c types.AppAutoscalingClient, cluster string, service string) (*astypes.ScalableTarget, error) {
	resourceID := fmt.Sprintf("service/%s/%s", cluster, service)

	p := applicationautoscaling.DescribeScalableTargetsInput{
		ServiceNamespace: astypes.ServiceNamespaceEcs,
		ResourceIds:      []string{resourceID},
	}

	r, err := c.DescribeScalableTargets(ctx, &p)

	if err != nil {
		return nil, err
	}

	if len(r.ScalableTargets) == 0 {
		return nil, &ErrNoResults{Message: "No scalable targets"}
	}

	return &r.ScalableTargets[0], nil
}
