module github.com/assemblyai/drone-deploy-ecs

go 1.16

require (
	github.com/aws/aws-sdk-go-v2 v1.19.1
	github.com/aws/aws-sdk-go-v2/config v1.8.2
	github.com/aws/aws-sdk-go-v2/credentials v1.13.27
	github.com/aws/aws-sdk-go-v2/service/applicationautoscaling v1.21.4
	github.com/aws/aws-sdk-go-v2/service/ecs v1.9.1
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.16.0
	github.com/aws/aws-sdk-go-v2/service/sts v1.19.3
	github.com/aws/smithy-go v1.13.5
	github.com/pkg/errors v0.9.1 // indirect
	gotest.tools v2.2.0+incompatible
)
