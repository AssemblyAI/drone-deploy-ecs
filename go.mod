module github.com/assemblyai/drone-deploy-ecs

go 1.16

require (
	github.com/aws/aws-sdk-go-v2 v1.11.0
	github.com/aws/aws-sdk-go-v2/config v1.8.2
	github.com/aws/aws-sdk-go-v2/service/applicationautoscaling v1.9.0
	github.com/aws/aws-sdk-go-v2/service/ecs v1.9.1
	github.com/aws/smithy-go v1.9.0
	github.com/pkg/errors v0.9.1 // indirect
	gotest.tools v2.2.0+incompatible
)
