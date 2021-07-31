package main

import (
	"context"
	"errors"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
)

func checkEnvVars() error {
	requiredVars := []string{
		"PLUGIN_AWS_REGION",
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
		"PLUGIN_BLUE_SERVICE",
		"PLUGIN_GREEN_SERVICE",
		"PLUGIN_SCALE_DOWN_PERCENT",
		"PLUGIN_SCALE_DOWN_INTERVAL",
		"PLUGIN_SCALE_DOWN_WAIT_PERIOD",
		"PLUGIN_CHECKS_TO_PASS",
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

func newAppAutoscalingClient(region string) *applicationautoscaling.Client {
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(region),
	)

	if err != nil {
		log.Fatalf("Failed to load SDK configuration, %v", err)
	}

	return applicationautoscaling.NewFromConfig(cfg)
}

func getServiceNames(s string) []string {

	return strings.Split(s, ",")
}
