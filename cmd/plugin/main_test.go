package main

import (
	"os"
	"testing"

	"gotest.tools/assert"
)

func TestNewECSClient(t *testing.T) {
	newECSClient("us-east-2")
}

func TestCheckEnvVarsAllVarsSet(t *testing.T) {
	os.Setenv("PLUGIN_AWS_REGION", "us-east-2")
	os.Setenv("PLUGIN_SERVICE", "some-service")
	os.Setenv("PLUGIN_CLUSTER", "some-cluster")
	os.Setenv("PLUGIN_CONTAINER", "some-container-name")
	os.Setenv("PLUGIN_IMAGE", "some/image:with-tag")

	err := checkEnvVars()

	assert.Equal(t, nil, err)
}

func TestCheckEnvVarsMissing(t *testing.T) {
	os.Unsetenv("PLUGIN_IMAGE")

	err := checkEnvVars()

	assert.Error(t, err, "env var not set")

}
