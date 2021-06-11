package main

import (
	"context"
	"errors"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/assemblyai/drone-deploy-ecs/pkg/deploy"
	"github.com/assemblyai/drone-deploy-ecs/pkg/types"
)

const (
	defaultMaxChecksUntilFailed = 60 // 10 second between checks + 60 checks = 600 seconds = 10 minutes
)

/*
Blue / Green

QUESTIONS:
- What do we do about autoscaling?
  - Need to figure out how to set max count to 0

When we do a blue/green deployment, we need to discover which service is "green". As far as the plugin is concerned, the service with 0 replicas is green

If we can't decide which service is blue and which is green, we should exit with a reconcile error

After we decide which service to update first, we need to modify the task definition. This is the same as a rolling deployment.

Now that we know which service is blue, we need to figure out how many replicas it has. We'll set green to have the same

After the task definition is updated, we need to set the green service to use the new TD version

We'll continue to use the ECS service deployment status for deciding if the deployment is working or not

Once the green service is healthy, we can scale down blue
*/

func determineBlueGreen(e types.ECSClient, blueService string, greenService string, cluster string) (string, string, error) {
	blueCount, err := deploy.GetServiceDesiredCount(context.Background(), e, blueService, cluster)

	if err != nil {
		log.Println("Error retrieving desired count for blue service", err.Error())
		return "", "", errors.New("deploy failed")
	}

	greenCount, err := deploy.GetServiceDesiredCount(context.Background(), e, greenService, cluster)

	if err != nil {
		log.Println("Error retrieving desired count for blue service", err.Error())
		return "", "", errors.New("deploy failed")
	}

	if blueCount == 0 {
		return blueService, greenService, nil
	}

	if greenCount == 0 {
		return greenService, blueService, nil
	}

	log.Println("Unable to determine which service is blue and which is green")
	log.Printf("Service '%s' has %d desired replicas while service '%s' has %d desired replicas. One of these should be 0", blueService, blueCount, greenService, greenCount)
	return "", "", errors.New("reconcile error")
}

// Return values -> success (bool), error
func release(e types.ECSClient, service string, cluster string, maxDeployChecks int, taskDefinitionARN string) (bool, error) {
	var err error

	deployCounter := 0
	deployFinished := false
	deployFailed := false

	deploymentID, err := deploy.UpdateServiceTaskDefinitionVersion(context.TODO(), e, service, cluster, taskDefinitionARN)

	if err != nil {
		log.Println("Error updating task definition for service", err.Error())
		return true, errors.New("deploy failed")
	}

	log.Println("Started deployment with ID", deploymentID)

	for !deployFinished {
		// Ensure that we haven't hit this limit
		// We want to rollback quickly
		if deployCounter > maxDeployChecks {
			log.Println("Reached max check limit. Will attempt rollback")
			deployFinished = true
			deployFailed = true
			break
		}

		log.Println("Waiting for deployment to complete. Check number:", deployCounter)
		time.Sleep(10 * time.Second)
		deployCounter++

		deployFinished, err = deploy.CheckDeploymentStatus(
			context.TODO(),
			e,
			service,
			cluster,
			deploymentID,
		)

		if err != nil {
			log.Println("Deployment failed: ", err.Error())
			deployFinished = true
			deployFailed = true
			break
		}

	}

	if deployFailed {
		return false, errors.New("deploy failed")
	}

	return true, nil
}

func blueGreen(e types.ECSClient, cluster string, container string, image string) error {
	blueServiceName := os.Getenv("PLUGIN_BLUE_SERVICE_NAME")
	greenServiceName := os.Getenv("PLUGIN_GREEN_SERVICE_NAME")

	determinedBlueService, determinedGreenService, err := determineBlueGreen(e, blueServiceName, greenServiceName, cluster)

	if err != nil {
		return err
	}

	td, err := deploy.GetServiceRunningTaskDefinition(context.TODO(), e, determinedBlueService, cluster)

	if err != nil {
		log.Println("Failing because of an error determining the currently in-use task definition")
		return err
	}

	currTD, err := deploy.RetrieveTaskDefinition(context.TODO(), e, td)

	if err != nil {
		log.Println("Failing because of an error retrieving the currently in-use task definition")
		return err
	}

	newTD, err := deploy.CreateNewTaskDefinitionRevision(context.TODO(), e, currTD, container, image)

	if err != nil {
		log.Println("Failing because of an error retrieving the creating a new task definition revision")
		return err
	}

	log.Println("Created new task definition revision", newTD.Revision)

	currBlueDesiredCount, err := deploy.GetServiceDesiredCount(context.Background(), e, determinedBlueService, cluster)

	if err != nil {
		log.Println("Failing because of an error determining desired count for blue service", err.Error())
		return err
	}

	// There is no deployment so discard it
	_, err = deploy.UpdateServiceTaskDefinitionVersion(context.TODO(), e, determinedGreenService, cluster, *newTD.TaskDefinitionArn)

	if err != nil {
		log.Println("Error updating task definition for service", err.Error())
		return errors.New("deploy failed")
	}

	// Scale up green service to the same count as blue
	// TODO pass the actual max instead of 0
	deploy.ScaleUp(e, currBlueDesiredCount, 0)
	log.Println("Pausing for 10 seconds while ECS schedules", currBlueDesiredCount, " containers")
	time.Sleep(10 * time.Second)

	// Loop and check GreenScaleUpFinished - timeout and fail after maxDeployChecks

	// Scale down

	return nil
}

func rolling(e types.ECSClient, cluster string, container string, image string, maxDeployChecks int) error {
	service := os.Getenv("PLUGIN_SERVICE")

	td, err := deploy.GetServiceRunningTaskDefinition(context.TODO(), e, service, cluster)

	if err != nil {
		log.Println("Failing because of an error determining the currently in-use task definition")
		return errors.New("deploy failed")
	}

	currTD, err := deploy.RetrieveTaskDefinition(context.TODO(), e, td)

	if err != nil {
		log.Println("Failing because of an error retrieving the currently in-use task definition")
		return errors.New("deploy failed")
	}

	newTD, err := deploy.CreateNewTaskDefinitionRevision(context.TODO(), e, currTD, container, image)

	if err != nil {
		log.Println("Failing because of an error retrieving the creating a new task definition revision")
		return errors.New("deploy failed")
	}

	log.Println("Created new task definition revision", newTD.Revision)

	deploymentOK, _ := release(e, service, cluster, maxDeployChecks, *newTD.TaskDefinitionArn)

	if !deploymentOK {
		log.Println("Rolling back failed deployment")
		rollbackOK, _ := release(e, service, cluster, maxDeployChecks, *currTD.TaskDefinitionArn)

		if !rollbackOK {
			log.Println("Error rolling back")
		}
		return errors.New("deploy failed")
	}

	log.Println("Deployment succeeded")
	return nil
}

func main() {
	// Ensure all required env vars are present
	if err := checkEnvVars(); err != nil {
		os.Exit(1)
	}

	e := newECSClient(os.Getenv("PLUGIN_AWS_REGION"))

	var maxDeployChecks int

	if os.Getenv("PLUGIN_MAX_DEPLOY_CHECKS") == "" {
		log.Println("PLUGIN_MAX_DEPLOY_CHECKS environment variable not set. Defaulting to", defaultMaxChecksUntilFailed)
		maxDeployChecks = defaultMaxChecksUntilFailed
	} else {
		convertResult, err := strconv.Atoi(os.Getenv("PLUGIN_MAX_DEPLOY_CHECKS"))
		if err != nil {
			log.Printf("Error converting '%s' to int. Defaulting to 60 checks, which is 10 minutes\n", os.Getenv("PLUGIN_MAX_DEPLOY_CHECKS"))
			maxDeployChecks = defaultMaxChecksUntilFailed
		} else {
			maxDeployChecks = convertResult
		}
	}

	cluster := os.Getenv("PLUGIN_CLUSTER")
	container := os.Getenv("PLUGIN_CONTAINER")
	image := os.Getenv("PLUGIN_IMAGE")

	if os.Getenv("PLUGIN_MODE") == "blue-green" {

	} else {
		if err := rolling(e, cluster, container, image, maxDeployChecks); err != nil {
			os.Exit(1)
		}
	}
}
