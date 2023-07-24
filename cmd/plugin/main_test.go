package main

import (
	"os"
	"testing"

	"github.com/assemblyai/drone-deploy-ecs/pkg/deploy"
	"github.com/assemblyai/drone-deploy-ecs/pkg/types"
	"gotest.tools/assert"
)

func TestNewECSClient(t *testing.T) {
	newECSClient("us-east-2", "arn:aws:iam::123456789012:role/some-role")
}

func TestCheckEnvVarsAllVarsSet(t *testing.T) {
	os.Setenv("PLUGIN_AWS_REGION", "us-east-2")
	os.Setenv("PLUGIN_SERVICE", "some-service")
	os.Setenv("PLUGIN_CLUSTER", "some-cluster")
	os.Setenv("PLUGIN_CONTAINER", "some-container-name")
	os.Setenv("PLUGIN_IMAGE", "some/image:with-tag")
	os.Setenv("PLUGIN_MODE", "rolling")

	err := checkEnvVars()

	assert.Equal(t, nil, err)
}

func TestCheckEnvVarsMissing(t *testing.T) {
	os.Unsetenv("PLUGIN_CLUSTER")

	err := checkEnvVars()

	assert.Error(t, err, "env var not set")

}

func Test_release(t *testing.T) {
	type args struct {
		e                 types.ECSClient
		service           string
		cluster           string
		maxDeployChecks   int
		taskDefinitionARN string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "test-failure",
			args: args{
				// TODO create a mock pkg
				e: deploy.MockECSClient{
					DeploymentState: "FAILED",
					TestingT:        t,
				},
				service:           "test-service",
				cluster:           "test-cluster",
				maxDeployChecks:   3,
				taskDefinitionARN: "arn:aws:ecs:us-west-2:123456789012:task-definition/amazon-ecs-sample:1",
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "test-success",
			args: args{
				// TODO create a mock pkg
				e: deploy.MockECSClient{
					DeploymentState: "COMPLETED",
					TestingT:        t,
				},
				service:           "test-service",
				cluster:           "test-cluster",
				maxDeployChecks:   3,
				taskDefinitionARN: "arn:aws:ecs:us-west-2:123456789012:task-definition/amazon-ecs-sample:1",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "test-timeout",
			args: args{
				// TODO create a mock pkg
				e: deploy.MockECSClient{
					DeploymentState: "IN_PROGRESS",
					TestingT:        t,
				},
				service:           "test-service",
				cluster:           "test-cluster",
				maxDeployChecks:   1,
				taskDefinitionARN: "arn:aws:ecs:us-west-2:123456789012:task-definition/amazon-ecs-sample:1",
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := release(tt.args.e, tt.args.service, tt.args.cluster, tt.args.maxDeployChecks, tt.args.taskDefinitionARN)
			if (err != nil) != tt.wantErr {
				t.Errorf("release() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("release() = %v, want %v", got, tt.want)
			}
		})
	}
}
