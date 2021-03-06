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

func GetServiceDesiredCount(ctx context.Context, c types.ECSClient, service string, cluster string) (int32, error) {
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
		return 0, err
	}

	return out.Services[0].DesiredCount, nil
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
// TODO remove deploymentID
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

	if out.Services[0].Deployments[0].FailedTasks > 0 {
		log.Printf("Deployment %d has failed tasks\n", out.Services[0].Deployments[0].FailedTasks)
		// This is helpful for debugging
		// ECS will clear all stopped tasks after a deployment finishes
		showFailedTasks(c, service, cluster, deploymentID)
		return true, errors.New("deployment failed")
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

func showFailedTasks(c types.ECSClient, service string, cluster string, deploymentID string) {
	input := ecs.ListTasksInput{
		Cluster:       &cluster,
		DesiredStatus: "STOPPED",
		StartedBy:     &deploymentID,
	}

	log.Printf("Checking failed tasks for service '%s'", service)

	resp, err := c.ListTasks(context.TODO(), &input)

	if err != nil {
		log.Printf("Error listing tasks: %s", err.Error())
		return
	}

	tasks, err := c.DescribeTasks(
		context.TODO(),
		&ecs.DescribeTasksInput{
			Tasks:   resp.TaskArns,
			Cluster: &cluster,
		},
	)

	if err != nil {
		log.Printf("Error describing tasks: %s", err.Error())
		return
	}

	for _, task := range tasks.Tasks {
		stoppedReason := task.StoppedReason
		taskARN := task.TaskArn

		log.Printf("Task '%s' failure reason: %s", *taskARN, *stoppedReason)
	}

}

func setECSServiceDesiredCount(c types.ECSClient, service string, cluster string, desiredCount int32) error {

	p := ecs.UpdateServiceInput{
		Service:      &service,
		DesiredCount: &desiredCount,
		Cluster:      &cluster,
	}

	// TODO use provided context
	_, err := c.UpdateService(context.Background(), &p)

	return err
}

// TODO update mock client so we can test this
func (c DeployConfig) GreenScaleUpFinished(ctx context.Context, service string) (bool, error) {
	i := ecs.DescribeServicesInput{
		Services: []string{service},
		Cluster:  aws.String(c.Cluster),
	}

	out, err := c.ECS.DescribeServices(
		ctx,
		&i,
	)

	if err != nil {
		log.Println("Error describing service: ", err.Error())
		return true, err
	}

	if out.Services[0].RunningCount != out.Services[0].DesiredCount {
		return false, nil
	}

	return true, nil
}
