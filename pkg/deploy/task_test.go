package deploy

import (
	"context"
	"testing"

	"gotest.tools/assert"
)

func TestRetrieveTaskDefinition(t *testing.T) {
	c := MockECSClient{}

	o, err := RetrieveTaskDefinition(
		context.TODO(),
		c,
		testTDARN,
	)

	assert.Equal(t, nil, err)
	assert.Equal(t, 2, len(o.ContainerDefinitions))
	assert.Equal(t, "sidecar", *o.ContainerDefinitions[1].Name)
}
