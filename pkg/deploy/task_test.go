package deploy

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/assemblyai/drone-deploy-ecs/pkg/types"
	"github.com/aws/aws-sdk-go-v2/aws"
	ecstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"gotest.tools/assert"
)

func TestRetrieveTaskDefinition1(t *testing.T) {
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

func Test_updateImage(t *testing.T) {
	type args struct {
		containers    []ecstypes.ContainerDefinition
		containerName string
		newImage      string
	}
	tests := []struct {
		name    string
		args    args
		want    []ecstypes.ContainerDefinition
		wantErr bool
	}{
		{
			name: "ensure-same-containers",
			args: args{
				containers: []ecstypes.ContainerDefinition{
					{
						Command: []string{"echo", "hello world"},
						Cpu:     100,
						Image:   aws.String("foo/sidecar:1"),
						Name:    aws.String("sidecar"),
					},
					{
						Command: []string{"echo", "hello world from app"},
						Cpu:     150,
						Name:    aws.String("app"),
						Image:   aws.String("foo/app:2"),
					},
				},
				containerName: "app",
				newImage:      "foo/app:3",
			},

			want: []ecstypes.ContainerDefinition{
				{
					Command: []string{"echo", "hello world"},
					Cpu:     100,
					Image:   aws.String("foo/sidecar:1"),
					Name:    aws.String("sidecar"),
				},
				{
					Command: []string{"echo", "hello world from app"},
					Cpu:     150,
					Name:    aws.String("app"),
					Image:   aws.String("foo/app:3"),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := updateImage(tt.args.containers, tt.args.containerName, tt.args.newImage)
			if (err != nil) != tt.wantErr {
				t.Errorf("updateImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("updateImage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateNewTaskDefinitionRevision(t *testing.T) {
	type args struct {
		ctx           context.Context
		c             types.ECSClient
		taskDefintion ecstypes.TaskDefinition
		containerName string
		newImage      string
	}
	tests := []struct {
		name    string
		args    args
		want    *ecstypes.TaskDefinition
		wantErr bool
	}{
		{
			name: "success",

			args: args{
				ctx: context.Background(),
				c: MockECSClient{
					TestingT:        t,
					DeploymentState: "COMPLETED",
				},
				containerName: "app",
				newImage:      "foo/app:3",
				taskDefintion: ecstypes.TaskDefinition{
					Compatibilities: []ecstypes.Compatibility{},
					ContainerDefinitions: []ecstypes.ContainerDefinition{
						{
							Command: []string{"echo", "hello world"},
							Cpu:     100,
							Image:   aws.String("foo/sidecar:1"),
							Name:    aws.String("sidecar"),
						},
						{
							Command: []string{"echo", "hello world from app"},
							Cpu:     150,
							Name:    aws.String("app"),
							Image:   aws.String("foo/app:2"),
						},
					},
					Cpu:                     aws.String("200"),
					EphemeralStorage:        &ecstypes.EphemeralStorage{},
					ExecutionRoleArn:        new(string),
					Family:                  new(string),
					InferenceAccelerators:   []ecstypes.InferenceAccelerator{},
					IpcMode:                 "",
					Memory:                  new(string),
					NetworkMode:             "",
					PidMode:                 "",
					PlacementConstraints:    []ecstypes.TaskDefinitionPlacementConstraint{},
					ProxyConfiguration:      &ecstypes.ProxyConfiguration{},
					RegisteredAt:            &time.Time{},
					RegisteredBy:            new(string),
					RequiresAttributes:      []ecstypes.Attribute{},
					RequiresCompatibilities: []ecstypes.Compatibility{},
					Revision:                0,
					Status:                  "COMPLETED",
					TaskDefinitionArn:       aws.String(testTDARN),
					TaskRoleArn:             new(string),
					Volumes:                 []ecstypes.Volume{},
				},
			},
			wantErr: false,
			want: &ecstypes.TaskDefinition{
				Compatibilities: []ecstypes.Compatibility{},
				ContainerDefinitions: []ecstypes.ContainerDefinition{
					{
						Command: []string{"echo", "hello world"},
						Cpu:     100,
						Image:   aws.String("foo/sidecar:1"),
						Name:    aws.String("sidecar"),
					},
					{
						Command: []string{"echo", "hello world from app"},
						Cpu:     150,
						Name:    aws.String("app"),
						Image:   aws.String("foo/app:3"),
					},
				},
				Cpu:                     aws.String("200"),
				EphemeralStorage:        &ecstypes.EphemeralStorage{},
				ExecutionRoleArn:        new(string),
				Family:                  new(string),
				InferenceAccelerators:   []ecstypes.InferenceAccelerator{},
				IpcMode:                 "",
				Memory:                  new(string),
				NetworkMode:             "",
				PidMode:                 "",
				PlacementConstraints:    []ecstypes.TaskDefinitionPlacementConstraint{},
				ProxyConfiguration:      &ecstypes.ProxyConfiguration{},
				RequiresCompatibilities: []ecstypes.Compatibility{},
				Status:                  "COMPLETED",
				TaskDefinitionArn:       aws.String(testTDARN),
				TaskRoleArn:             new(string),
				Volumes:                 []ecstypes.Volume{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateNewTaskDefinitionRevision(tt.args.ctx, tt.args.c, tt.args.taskDefintion, tt.args.containerName, tt.args.newImage)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateNewTaskDefinitionRevision() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateNewTaskDefinitionRevision() = %v, want %v", got, tt.want)
			}
		})
	}
}
