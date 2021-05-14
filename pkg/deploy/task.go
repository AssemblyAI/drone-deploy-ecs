package deploy

import (
	"context"
	"errors"
	"log"

	"github.com/assemblyai/drone-deploy-ecs/pkg/types"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

// RetrieveTaskDefinition saves the current task definition for rollbacks
func RetrieveTaskDefinition(ctx context.Context, c types.ECSClient, taskDefinitionARN string) (ecstypes.TaskDefinition, error) {
	td := ecstypes.TaskDefinition{}

	i := ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(taskDefinitionARN),
	}

	out, err := c.DescribeTaskDefinition(
		ctx,
		&i,
	)

	td = *out.TaskDefinition

	if err != nil {
		log.Println("Error describing task definition: ", err.Error())
		return td, err
	}

	return td, nil
}

func CreateNewTaskDefinitionRevision(ctx context.Context, c types.ECSClient, taskDefintion ecstypes.TaskDefinition, containerName string, newImage string) (*ecstypes.TaskDefinition, error) {
	updatedContainers, err := updateImage(taskDefintion.ContainerDefinitions, containerName, newImage)

	if err != nil {
		return nil, err
	}

	i := ecs.RegisterTaskDefinitionInput{
		ContainerDefinitions:    updatedContainers,
		Family:                  taskDefintion.Family,
		Cpu:                     taskDefintion.Cpu,
		EphemeralStorage:        taskDefintion.EphemeralStorage,
		ExecutionRoleArn:        taskDefintion.ExecutionRoleArn,
		InferenceAccelerators:   taskDefintion.InferenceAccelerators,
		IpcMode:                 taskDefintion.IpcMode,
		Memory:                  taskDefintion.Memory,
		NetworkMode:             taskDefintion.NetworkMode,
		PidMode:                 taskDefintion.PidMode,
		PlacementConstraints:    taskDefintion.PlacementConstraints,
		ProxyConfiguration:      taskDefintion.ProxyConfiguration,
		RequiresCompatibilities: taskDefintion.RequiresCompatibilities,
		TaskRoleArn:             taskDefintion.TaskRoleArn,
		Volumes:                 taskDefintion.Volumes,
	}

	out, err := c.RegisterTaskDefinition(
		ctx,
		&i,
	)

	if err != nil {
		log.Println("Error registering task definition: ", err.Error())
		return nil, err
	}

	return out.TaskDefinition, nil

}

func updateImage(containers []ecstypes.ContainerDefinition, containerName string, newImage string) ([]ecstypes.ContainerDefinition, error) {
	var resp []ecstypes.ContainerDefinition

	containerFound := false

	for idx, container := range containers {
		if *container.Name == containerName {
			log.Printf("Updating container '%s' from '%s' to '%s'\n", *container.Name, *container.Image, newImage)
			resp = append(resp, container)
			resp[idx].Image = &newImage
			containerFound = true
		} else {
			resp = append(resp, container)
		}
	}

	if !containerFound {
		log.Printf("Container '%s' not found. Cannot proceed.\n", containerName)
		return resp, errors.New("container not found")
	}

	return resp, nil
}
