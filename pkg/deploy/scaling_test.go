package deploy

import (
	"context"
	"testing"

	"github.com/assemblyai/drone-deploy-ecs/pkg/types"
)

func TestAppAutoscalingTargetExists(t *testing.T) {
	type args struct {
		ctx     context.Context
		c       types.AppAutoscalingClient
		cluster string
		service string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "target-does-not-exist",
			args: args{
				ctx:     context.Background(),
				c:       MockAppAutoscalingClient{WantError: false, TestingT: t, TargetExists: false},
				cluster: "test-cluster",
				service: "ci-service-green",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "target-exists",
			args: args{
				ctx:     context.Background(),
				c:       MockAppAutoscalingClient{WantError: false, TestingT: t, TargetExists: true},
				cluster: "test-cluster",
				service: "ci-service-green",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "error",
			args: args{
				ctx:     context.Background(),
				c:       MockAppAutoscalingClient{WantError: true, TestingT: t, TargetExists: true},
				cluster: "test-cluster",
				service: "ci-service-green",
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AppAutoscalingTargetExists(tt.args.ctx, tt.args.c, tt.args.cluster, tt.args.service)
			if (err != nil) != tt.wantErr {
				t.Errorf("AppAutoscalingTargetExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("AppAutoscalingTargetExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetServiceMinMaxCount(t *testing.T) {
	type args struct {
		ctx     context.Context
		c       types.AppAutoscalingClient
		cluster string
		service string
	}
	tests := []struct {
		name    string
		args    args
		want    int32
		want1   int32
		wantErr bool
	}{
		{
			name: "error",
			args: args{
				ctx:     context.Background(),
				c:       MockAppAutoscalingClient{WantError: true, TestingT: t, TargetExists: false},
				cluster: "test-cluster",
				service: "ci-service-green",
			},
			want:    0,
			want1:   0,
			wantErr: true,
		},
		{
			name: "correct-count",
			args: args{
				ctx:     context.Background(),
				c:       MockAppAutoscalingClient{WantError: false, TestingT: t, TargetExists: true},
				cluster: "test-cluster",
				service: "ci-service-green",
			},
			want:    20,
			want1:   4,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := GetServiceMinMaxCount(tt.args.ctx, tt.args.c, tt.args.cluster, tt.args.service)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetServiceMinMaxCount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetServiceMinMaxCount() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("GetServiceMinMaxCount() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_setAppAutoscalingCounts(t *testing.T) {
	type args struct {
		ctx      context.Context
		c        types.AppAutoscalingClient
		service  string
		cluster  string
		maxCount int32
		minCount int32
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				ctx:      context.Background(),
				c:        MockAppAutoscalingClient{WantError: false, TestingT: t},
				cluster:  "test-cluster",
				service:  "ci-service-green",
				maxCount: 50,
				minCount: 10,
			},
			wantErr: false,
		},
		{
			name: "error",
			args: args{
				ctx:      context.Background(),
				c:        MockAppAutoscalingClient{WantError: true, TestingT: t},
				cluster:  "test-cluster",
				service:  "ci-service-green",
				maxCount: 50,
				minCount: 10,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := setAppAutoscalingCounts(tt.args.ctx, tt.args.c, tt.args.service, tt.args.cluster, tt.args.maxCount, tt.args.minCount); (err != nil) != tt.wantErr {
				t.Errorf("setAppAutoscalingCounts() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
