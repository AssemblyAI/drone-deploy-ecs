package main

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
)

func checkEnvVars() error {
	requiredVars := []string{
		"PLUGIN_AWS_REGION",
		"PLUGIN_SERVICE",
		"PLUGIN_CLUSTER",
		"PLUGIN_CONTAINER",
		"PLUGIN_IMAGE",
		"PLUGIN_MODE",
	}

	for _, v := range requiredVars {
		if os.Getenv(v) == "" {
			log.Printf("Required environment variable '%s' is missing\n", v)
			return errors.New("env var not set")
		}
	}

	return nil
}

func parseRollingVars() error {
	requiredVars := []string{
		"PLUGIN_SERVICE",
	}

	for _, v := range requiredVars {
		if os.Getenv(v) == "" {
			log.Printf("Required environment variable '%s' is missing\n", v)
			return errors.New("env var not set")
		}
	}

	return nil
}

func checkBlueGreenVars() error {
	requiredVars := []string{
		"PLUGIN_BLUE_SERVICE_NAME",
		"PLUGIN_GREEN_SERVICE_NAME",
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

func newAppAutoscalingTarget(region string) *applicationautoscaling.Client {
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(region),
	)

	if err != nil {
		log.Fatalf("Failed to load SDK configuration, %v", err)
	}

	return applicationautoscaling.NewFromConfig(cfg)
}
