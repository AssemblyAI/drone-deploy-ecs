package deploy

import "github.com/assemblyai/drone-deploy-ecs/pkg/types"

type DeployConfig struct {
	ECS            types.ECSClient
	AppAutoscaling types.AppAutoscalingClient
	Cluster        string
	Container      string
	Image          string
	// Logger
}
