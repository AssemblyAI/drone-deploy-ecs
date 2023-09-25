module github.com/assemblyai/drone-deploy-ecs

go 1.16

require (
	github.com/aws/aws-sdk-go-v2 v1.21.0
	github.com/aws/aws-sdk-go-v2/config v1.18.42
	github.com/aws/aws-sdk-go-v2/credentials v1.13.40
	github.com/aws/aws-sdk-go-v2/service/applicationautoscaling v1.6.1
	github.com/aws/aws-sdk-go-v2/service/ecs v1.9.1
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.16.0
	github.com/aws/aws-sdk-go-v2/service/sts v1.22.0
	github.com/aws/smithy-go v1.14.2
	github.com/pkg/errors v0.9.1 // indirect
	gotest.tools v2.2.0+incompatible
)
