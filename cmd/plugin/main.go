package main

import (
	"log"
	"os"
	"strconv"
)

const (
	defaultMaxChecksUntilFailed = 60 // 10 second between checks + 60 checks = 600 seconds = 10 minutes
)

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
