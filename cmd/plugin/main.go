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
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
)

const (
	defaultMaxChecksUntilFailed = 60 // 10 second between checks + 60 checks = 600 seconds = 10 minutes
)

func checkEnvVars() error {
	requiredVars := []string{
		"PLUGIN_AWS_REGION",
		"PLUGIN_SERVICE",
		"PLUGIN_CLUSTER",
		"PLUGIN_CONTAINER",
		"PLUGIN_IMAGE",
	}

	for _, v := range requiredVars {
		if os.Getenv(v) == "" {
			log.Printf("Required environment variable '%s' is missing\n", v)
			return errors.New("env var not set")
		}
	}

	return nil
}

func newECSClient(region string) *ecs.Client {
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(region),
	)

	if err != nil {
		log.Fatalf("Failed to load SDK configuration, %v", err)
	}

	return ecs.NewFromConfig(cfg)
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
		}

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
		}

		log.Println("Waiting for deployment to complete. Check number:", deployCounter)
		time.Sleep(10 * time.Second)
		deployCounter++
	}

	if deployFailed {
		return false, errors.New("deploy failed")
	}

	return true, nil
}

func main() {
	// Ensure all required env vars are present
	if err := checkEnvVars(); err != nil {
		os.Exit(1)
	}

	e := newECSClient(os.Getenv("PLUGIN_AWS_REGION"))

	var maxDeployChecks int

	service := os.Getenv("PLUGIN_SERVICE")
	cluster := os.Getenv("PLUGIN_CLUSTER")
	container := os.Getenv("PLUGIN_CONTAINER")
	image := os.Getenv("PLUGIN_IMAGE")

	if os.Getenv("PLUGIN_MAX_DEPLOY_CHECKS") == "" {
		log.Println("PLUGIN_MAX_DEPLOY_CHECKS environment variable not set. Defaulting to", defaultMaxChecksUntilFailed)
		maxDeployChecks = defaultMaxChecksUntilFailed
	} else {
		convertResult, err := strconv.Atoi(os.Getenv("PLUGIN_MAX_DEPLOY_CHECKS"))
		if err != nil {
			log.Printf("Error converting '%s' to int. Default to 60 checks, which is 10 minutes\n", os.Getenv("PLUGIN_MAX_DEPLOY_CHECKS"))
			maxDeployChecks = defaultMaxChecksUntilFailed
		} else {
			maxDeployChecks = convertResult
		}
	}

	td, err := deploy.GetServiceRunningTaskDefinition(context.TODO(), e, service, cluster)

	if err != nil {
		log.Println("Failing because of an error determining the currently in-use task definition")
		os.Exit(1)
	}

	currTD, err := deploy.RetrieveTaskDefinition(context.TODO(), e, td)

	if err != nil {
		log.Println("Failing because of an error retrieving the currently in-use task definition")
		os.Exit(1)
	}

	newTD, err := deploy.CreateNewTaskDefinitionRevision(context.TODO(), e, currTD, container, image)

	if err != nil {
		log.Println("Failing because of an error retrieving the creating a new task definition revision")
		os.Exit(1)
	}

	log.Println("Created new task definition revision", newTD.Revision)

	deploymentOK, err := release(e, service, cluster, maxDeployChecks, *newTD.TaskDefinitionArn)

	if !deploymentOK {
		log.Println("Rolling back failed deployment")
		rollbackOK, _ := release(e, service, cluster, maxDeployChecks, *currTD.TaskDefinitionArn)

		if !rollbackOK {
			log.Println("Error rolling back")
		}
		// Mark build as failed because the initial deployment failed
		os.Exit(1)
	}

	log.Println("Deployment succeeded")

}
