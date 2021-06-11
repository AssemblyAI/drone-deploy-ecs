package main

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/assemblyai/drone-deploy-ecs/pkg/deploy"
	"github.com/assemblyai/drone-deploy-ecs/pkg/types"
)

/*
Blue / Green

When we do a blue/green deployment, we need to discover which service is "green". As far as the plugin is concerned, the service with 0 replicas is green

If we can't decide which service is blue and which is green, we should exit with a reconcile error

After we decide which service to update first, we need to modify the task definition. This is the same as a rolling deployment.

Now that we know which service is blue, we need to figure out how many replicas it has. We'll set green to have the same

After the task definition is updated, we need to set the green service to use the new TD version

Once green is using new task definition, we need to check blue for an app autoscaling target. If it exists,
we need to set green to have the same max. Otherwise we'll just work off desired count

Once we've done that, set green's desired count to blue's and watch until green's running count == desired count. This step will need to
have a timeout on it. If we reach the timeout, scale green back down and fail the deployment

Once green is up and running, scale down blue. If there's an autoscaling target, scale down by decrementing the max, otherwise just
work off of the desired count


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
