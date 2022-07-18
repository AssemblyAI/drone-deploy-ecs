module github.com/assemblyai/drone-deploy-ecs

go 1.16

require (
	github.com/aws/aws-sdk-go-v2 v1.16.7
	github.com/aws/aws-sdk-go-v2/config v1.15.14
	github.com/aws/aws-sdk-go-v2/service/applicationautoscaling v1.6.1
	github.com/aws/aws-sdk-go-v2/service/ecs v1.9.1
	github.com/aws/smithy-go v1.12.0
	github.com/pkg/errors v0.9.1 // indirect
	gotest.tools v2.2.0+incompatible
)
