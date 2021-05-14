package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/assemblyai/drone-deploy-ecs/pkg/deploy"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
)

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

func main() {
	e := newECSClient(os.Getenv("PLUGIN_AWS_REGION"))

	service := os.Getenv("PLUGIN_SERVICE")
	cluster := os.Getenv("PLUGIN_CLUSTER")
	container := os.Getenv("PLUGIN_CONTAINER")
	image := os.Getenv("PLUGIN_IMAGE")

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

	deploymentID, err := deploy.UpdateServiceTaskDefinitionVersion(context.TODO(), e, service, cluster, *newTD.TaskDefinitionArn)

	if err != nil {
		panic(err)
	}

	deployFinished := false

	log.Println("Deployment begun")

	for !deployFinished {
		deployFinished, err = deploy.CheckDeploymentStatus(
			context.TODO(),
			e,
			service,
			cluster,
			deploymentID,
		)

		if err != nil {
			log.Println("Deployment failed: ", err.Error())
			os.Exit(1)
		}

		log.Println("Waiting for deployment to complete")
		time.Sleep(5 * time.Second)
	}

	log.Println("Deployment complete. Successfully updated service")

}
