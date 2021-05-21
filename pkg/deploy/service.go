package deploy

import (
	"context"
	"errors"
	"log"

	"github.com/assemblyai/drone-deploy-ecs/pkg/types"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
)

func GetServiceRunningTaskDefinition(ctx context.Context, c types.ECSClient, service string, cluster string) (string, error) {
	i := ecs.DescribeServicesInput{
		Services: []string{service},
		Cluster:  aws.String(cluster),
	}

	out, err := c.DescribeServices(
		ctx,
		&i,
	)

	if err != nil {
		log.Println("Error describing service: ", err.Error())
		return "", err
	}

	return *out.Services[0].TaskDefinition, nil
}

func UpdateServiceTaskDefinitionVersion(ctx context.Context, c types.ECSClient, service string, cluster string, taskDefinitonARN string) (string, error) {

	i := ecs.UpdateServiceInput{
		Service:        aws.String(service),
		Cluster:        aws.String(cluster),
		TaskDefinition: aws.String(taskDefinitonARN),
	}

	out, err := c.UpdateService(
		ctx,
		&i,
	)

	if err != nil {
		log.Println("Error updating service: ", err.Error())
		return "", err
	}

	// The first item in the array should be the deployment we just created
	return *out.Service.Deployments[0].Id, nil
}

// CheckDeploymentStatus returns true if a deployment has finished (either success or failure) and false if the deployment is in progress
func CheckDeploymentStatus(ctx context.Context, c types.ECSClient, service string, cluster string, deploymentID string) (bool, error) {
	i := ecs.DescribeServicesInput{
		Services: []string{service},
		Cluster:  aws.String(cluster),
	}

	out, err := c.DescribeServices(
		ctx,
		&i,
	)

	if err != nil {
		log.Println("Error describing service: ", err.Error())
		return true, err
	}

	if out.Services[0].Deployments[0].RolloutState == "IN_PROGRESS" {
		return false, nil
	} else if out.Services[0].Deployments[0].RolloutState == "COMPLETED" {
		return true, nil
	} else {
		// The only other status is FAILED
		return true, errors.New("deployment failed")
	}

}
