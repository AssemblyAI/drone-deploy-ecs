package main

import (
	"log"
	"os"
	"strconv"

	"github.com/assemblyai/drone-deploy-ecs/pkg/deploy"
)

const (
	defaultMaxChecksUntilFailed = 60 // 10 second between checks + 60 checks = 600 seconds = 10 minutes
)

func main() {
	// Ensure all required env vars are present
	if err := checkEnvVars(); err != nil {
		os.Exit(1)
	}

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

	dc := deploy.DeployConfig{
		ECS:            newECSClient(os.Getenv("PLUGIN_AWS_REGION")),
		AppAutoscaling: newAppAutoscalingClient(os.Getenv("PLUGIN_AWS_REGION")),
		Cluster:        os.Getenv("PLUGIN_CLUSTER"),
		Container:      os.Getenv("PLUGIN_CONTAINER"),
		Image:          os.Getenv("PLUGIN_IMAGE"),
	}

	if os.Getenv("PLUGIN_MODE") == "blue-green" {
		if err := checkBlueGreenVars(); err != nil {
			os.Exit(1)
		}
		if err := blueGreen(dc, maxDeployChecks); err != nil {
			os.Exit(1)
		}
	} else {
		if err := rolling(dc.ECS, dc.Cluster, dc.Container, dc.Image, maxDeployChecks); err != nil {
			os.Exit(1)
		}
	}
}
