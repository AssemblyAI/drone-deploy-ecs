package deploy

import (
	"context"
	"testing"

	"github.com/assemblyai/drone-deploy-ecs/pkg/types"
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

func TestCheckDeploymentStatus(t *testing.T) {
	type args struct {
		ctx          context.Context
		c            types.ECSClient
		service      string
		cluster      string
		deploymentID string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "test-in-progress",
			args: args{
				ctx: context.TODO(),
				c: MockECSClient{
					TestingT:        t,
					DeploymentState: "IN_PROGRESS",
				},
				service:      "test-service",
				cluster:      "test-cluster",
				deploymentID: "test-deployment-id",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "test-completed",
			args: args{
				ctx: context.TODO(),
				c: MockECSClient{
					TestingT:        t,
					DeploymentState: "COMPLETED",
				},
				service:      "test-service",
				cluster:      "test-cluster",
				deploymentID: "test-deployment-id",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "test-failed",
			args: args{
				ctx: context.TODO(),
				c: MockECSClient{
					TestingT:        t,
					DeploymentState: "FAILED",
				},
				service:      "test-service",
				cluster:      "test-cluster",
				deploymentID: "test-deployment-id",
			},
			want:    true,
			wantErr: true,
		},
		{
			name: "test-failed-tasks",
			args: args{
				ctx: context.TODO(),
				c: MockECSClient{
					TestingT:        t,
					DeploymentState: "IN_PROGRESS",
					FailedTasks:     2,
				},
				service:      "test-service",
				cluster:      "test-cluster",
				deploymentID: "test-deployment-id",
			},
			want:    true,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CheckDeploymentStatus(tt.args.ctx, tt.args.c, tt.args.service, tt.args.cluster, tt.args.deploymentID)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckDeploymentStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CheckDeploymentStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdateServiceTaskDefinitionVersion(t *testing.T) {
	type args struct {
		ctx              context.Context
		c                types.ECSClient
		service          string
		cluster          string
		taskDefinitonARN string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx:              context.Background(),
				c:                MockECSClient{TestingT: t, WantError: false},
				service:          "test-service",
				cluster:          "test-cluster",
				taskDefinitonARN: testTDARN,
			},
			want:    "test-deployment",
			wantErr: false,
		},
		{
			name: "failure",
			args: args{
				ctx:              context.Background(),
				c:                MockECSClient{TestingT: t, WantError: true},
				service:          "test-service",
				cluster:          "test-cluster",
				taskDefinitonARN: testTDARN,
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UpdateServiceTaskDefinitionVersion(tt.args.ctx, tt.args.c, tt.args.service, tt.args.cluster, tt.args.taskDefinitonARN)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateServiceTaskDefinitionVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("UpdateServiceTaskDefinitionVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_setECSServiceDesiredCount(t *testing.T) {
	type args struct {
		c            types.ECSClient
		service      string
		cluster      string
		desiredCount int32
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				c:       MockECSClient{TestingT: t, WantError: false},
				service: "test-service",
				cluster: "test-cluster",
			},
			wantErr: false,
		},
		{
			name: "failure",
			args: args{
				c:       MockECSClient{TestingT: t, WantError: true},
				service: "test-service",
				cluster: "test-cluster",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := setECSServiceDesiredCount(tt.args.c, tt.args.service, tt.args.cluster, tt.args.desiredCount); (err != nil) != tt.wantErr {
				t.Errorf("setECSServiceDesiredCount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetServiceDesiredCount(t *testing.T) {
	type args struct {
		ctx     context.Context
		c       types.ECSClient
		service string
		cluster string
	}
	tests := []struct {
		name    string
		args    args
		want    int32
		wantErr bool
	}{
		{
			name: "desired-count-is-2",
			args: args{
				ctx:     context.Background(),
				c:       MockECSClient{WantError: false, TestingT: t},
				cluster: "test-cluster",
				service: "test-service",
			},
			want:    2,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetServiceDesiredCount(tt.args.ctx, tt.args.c, tt.args.service, tt.args.cluster)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetServiceDesiredCount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetServiceDesiredCount() = %v, want %v", got, tt.want)
			}
		})
	}
}
