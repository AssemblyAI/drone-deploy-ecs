package main

import (
	"context"
	"log"
	"os"
	"strconv"

	"github.com/assemblyai/drone-deploy-ecs/pkg/deploy"
)

const (
	defaultMaxChecksUntilFailed = 60 // 10 second between checks + 60 checks = 600 seconds = 10 minutes
)

var (
	disableRollbacks bool
	maxDeployChecks  int
)

func main() {
	// Ensure all required env vars are present
	if err := checkEnvVars(); err != nil {
		os.Exit(1)
	}

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

	// Set disable_rollbacks to any string in order to disable them
	if os.Getenv("PLUGIN_DISABLE_ROLLBACKS") == "" {
		log.Println("Rollbacks are enabled. Note: this setting only applies to rolling deployments")
		disableRollbacks = false
	} else {
		log.Println("Rollbacks are disabled. Note: this setting only applies to rolling deployments")
		disableRollbacks = true
	}

	dc := deploy.DeployConfig{
		ECS:            newECSClient(os.Getenv("PLUGIN_AWS_REGION")),
		AppAutoscaling: newAppAutoscalingClient(os.Getenv("PLUGIN_AWS_REGION")),
		Cluster:        os.Getenv("PLUGIN_CLUSTER"),
		Container:      os.Getenv("PLUGIN_CONTAINER"),
		Image:          os.Getenv("PLUGIN_IMAGE"),
	}

	// check which deployment method to use based on the mode, default to rolling
	switch os.Getenv("PLUGIN_MODE") {
	case "blue-green":
		if err := checkBlueGreenVars(); err != nil {
			os.Exit(1)
		}
		if err := blueGreen(dc, maxDeployChecks); err != nil {
			os.Exit(1)
		}
	case "blue-green-cluster":
		// this is the same as rolling except that it deploys to the off color

		if err := checkBlueGreenClusterVars(); err != nil {
			os.Exit(1)
		}

		manager := newSecretsManagerClient(os.Getenv("PLUGIN_AWS_REGION"))
		//get the inactive env. Either service name (blue/green) can be used since it does a partial match.
		inactiveEnv, err := getGlobalInactiveEnvironment(manager, os.Getenv("DRONE_REPO_BRANCH"), os.Getenv("PLUGIN_SECRET_SERVICE"))

		if err != nil {
			log.Println(err)
			os.Exit(1)
		}

		//pick the image/service to deploy to based of configured live env
		image := os.Getenv("PLUGIN_BLUE_IMAGE")
		service := os.Getenv("PLUGIN_BLUE_SERVICE")

		if inactiveEnv == "green" {
			image = os.Getenv("PLUGIN_GREEN_IMAGE")
			service = os.Getenv("PLUGIN_GREEN_SERVICE")
		}

		count, err := deploy.GetServiceDesiredCount(context.Background(), dc.ECS, service, dc.Cluster)

		if err != nil {
			log.Printf("could not get desired count of service %s: %v\n", service, err)
			os.Exit(1)
		}

		if count != 0 {
			log.Printf("inactive environment for service %s has tasks running, this likely means we are attempting to deploy to the wrong env\n", service)
		}

		if err := rolling(dc.ECS, dc.Cluster, dc.Container, image, maxDeployChecks, service); err != nil {
			os.Exit(1)
		}
	default:
		if err := rolling(dc.ECS, dc.Cluster, dc.Container, dc.Image, maxDeployChecks, os.Getenv("PLUGIN_SERVICE")); err != nil {
			os.Exit(1)
		}
	}
}
