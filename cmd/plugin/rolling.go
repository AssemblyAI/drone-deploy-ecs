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

func rolling(e types.ECSClient, cluster string, container string, image string, maxDeployChecks int) error {
	services := getServiceNames(os.Getenv("PLUGIN_SERVICE"))

	// Retrieve the task definition that the first service is using. The first service may be the only service
	// In the event that len(services) > 1, we can reasonably assume that all services in the array use the same task definition
	// That's the entire point of this feature
	td, err := deploy.GetServiceRunningTaskDefinition(context.TODO(), e, services[0], cluster)

	if err != nil {
		log.Println("Failing because of an error determining the currently in-use task definition:", err.Error())
		return errors.New("deploy failed")
	}

	currTD, err := deploy.RetrieveTaskDefinition(context.TODO(), e, td)

	if err != nil {
		log.Println("Failing because of an error retrieving the currently in-use task definition:", err.Error())
		return errors.New("deploy failed")
	}

	newTD, err := deploy.CreateNewTaskDefinitionRevision(context.TODO(), e, currTD, container, image)

	if err != nil {
		log.Println("Failing because of an error retrieving the creating a new task definition revision:", err.Error())
		return errors.New("deploy failed")
	}

	log.Println("Created new task definition revision", newTD.Revision)

	for _, service := range services {
		log.Printf("Starting deployment for service '%s'\n", service)

		deploymentOK, _ := release(e, service, cluster, maxDeployChecks, *newTD.TaskDefinitionArn)

		if !deploymentOK {

			if disableRollbacks {
				log.Println("Deployment failed but rollbacks are disabled. If the service has ECS Circuit Breaker enabled, the circuit breaker should handle rolling back.")
				return errors.New("deploy failed")
			} else {
				log.Println("Rolling back failed deployment for service", service)
				rollbackOK, _ := release(e, service, cluster, maxDeployChecks, *currTD.TaskDefinitionArn)

				if !rollbackOK {
					log.Println("Error rolling back")
				}
				return errors.New("deploy failed")
			}
		}

		log.Printf("Deployment succeeded for service '%s'\n", service)

	}

	return nil
}
