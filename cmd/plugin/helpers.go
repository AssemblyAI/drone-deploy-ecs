package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	pluginTypes "github.com/assemblyai/drone-deploy-ecs/pkg/types"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"log"
	"os"
	"strings"
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

// checkBlueGreenClusterVars validates that the settings in the drone file are correct
func checkBlueGreenClusterVars() error {
	requiredVars := []string{
		"PLUGIN_BLUE_SERVICE",
		"PLUGIN_GREEN_SERVICE",
		"PLUGIN_BLUE_IMAGE",
		"PLUGIN_GREEN_IMAGE",
		"PLUGIN_SECRET_SERVICE", // this is the service tag in terraform for the secret with the color
	}

	hasError := false

	for _, v := range requiredVars {
		if os.Getenv(v) == "" {
			log.Printf("Required environment variable '%s' is missing\n", v)
			hasError = true
		}
	}

	if hasError {
		return errors.New("env var not set")
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

func newSecretsManagerClient(region string) *secretsmanager.Client {
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(region),
	)

	if err != nil {
		log.Fatalf("Failed to load SDK configuration, %v", err)
	}

	return secretsmanager.NewFromConfig(cfg)
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

// getGlobalInactiveEnvironment finds the appropriate global secret store that holds the current live color
func getGlobalInactiveEnvironment(manager pluginTypes.SecretmanagerClient, branch string, serviceSuffix string) (string, error) {
	environment := branch

	if branch == "main" {
		environment = "production"
	}

	secretName := fmt.Sprintf("%s-%s", environment, serviceSuffix)

	getParams := &secretsmanager.GetSecretValueInput{
		SecretId: &secretName,
	}

	getOut, err := manager.GetSecretValue(context.Background(), getParams)

	if err != nil {
		return "", fmt.Errorf("failed to retrieve secret value for live environment %v", err)
	}

	if getOut.SecretString == nil {
		return "", errors.New("no secret found")
	}

	secretString := *getOut.SecretString
	jsonMap := make(map[string]string)

	err = json.Unmarshal([]byte(secretString), &jsonMap)

	if err != nil {
		return "", fmt.Errorf("could not decode json secret %v", err)
	}

	inactiveEnv := "blue"

	if jsonMap["CURRENT_LIVE_ENVIRONMENT"] == "blue" {
		inactiveEnv = "green"
	}

	return inactiveEnv, nil
}
