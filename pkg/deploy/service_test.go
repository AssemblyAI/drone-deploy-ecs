package deploy

import (
	"context"
	"testing"

	"gotest.tools/assert"
)

func TestGetServiceRunningTaskDefinition(t *testing.T) {
	c := MockECSClient{
		DeploymentState: "COMPLETE",
		TestingT:        t,
	}

	o, err := GetServiceRunningTaskDefinition(
		context.TODO(),
		c,
		"test-service",
		"test-cluster",
	)

	assert.Equal(t, nil, err)
	assert.Equal(t, testTDARN, o)
}
