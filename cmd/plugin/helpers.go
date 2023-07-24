package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	pluginTypes "github.com/assemblyai/drone-deploy-ecs/pkg/types"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/applicationautoscaling"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"log"
	"os"
	"strings"
)

// checkEnvVars checks the vars needed for each mode
func checkEnvVars() error {
	requiredVars := []string{
		"PLUGIN_AWS_REGION",
		"PLUGIN_CLUSTER",
		"PLUGIN_CONTAINER",
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

func checkRollingVars() error {
	requiredVars := []string{
		"PLUGIN_SERVICE",
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

func newECSClient(region string, role_arn string) *ecs.Client {
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(region),
	)

	if err != nil {
		log.Fatalf("Failed to load SDK configuration, %v", err)
	}

	if role_arn != "" {
		stsClient := sts.NewFromConfig(cfg)
		provider := stscreds.NewAssumeRoleProvider(stsClient, role_arn)
		cfg.Credentials = aws.NewCredentialsCache(provider)
		cfg.Credentials.Retrieve(context.Background())
	}

	return ecs.NewFromConfig(cfg)
}

func newSecretsManagerClient(region string, role_arn string) *secretsmanager.Client {
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(region),
	)

	if err != nil {
		log.Fatalf("Failed to load SDK configuration, %v", err)
	}

	if role_arn != "" {
		stsClient := sts.NewFromConfig(cfg)
		provider := stscreds.NewAssumeRoleProvider(stsClient, role_arn)
		cfg.Credentials = aws.NewCredentialsCache(provider)
		cfg.Credentials.Retrieve(context.Background())
	}

	return secretsmanager.NewFromConfig(cfg)
}

func newAppAutoscalingClient(region string, role_arn string) *applicationautoscaling.Client {
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(region),
	)

	if err != nil {
		log.Fatalf("Failed to load SDK configuration, %v", err)
	}

	if role_arn != "" {
		stsClient := sts.NewFromConfig(cfg)
		provider := stscreds.NewAssumeRoleProvider(stsClient, role_arn)
		cfg.Credentials = aws.NewCredentialsCache(provider)
		cfg.Credentials.Retrieve(context.Background())
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
		return "", fmt.Errorf("failed to retrieve secret value (name: %s) for live environment %v", secretName, err)
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
