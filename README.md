# drone-deploy-ecs

## Overview

`drone-deploy-ecs` is an opinionated Drone plugin for updating a single container within an ECS Task.

During deployment, the plugin retrieves the active Task Definition for a specified ECS Service, creates a new revision of the Task Definition with an updated image for a specified container, updates the Service to use the new Task Definition, and waits for the deployment to complete.

[ECR Link](https://gallery.ecr.aws/assemblyai/drone-deploy-ecs)

## Important Notes

This plugin cannot update different containers within the same Task Definition simultaneously. It will only update the image for a single container within a Task Defintion

The ECS Service must use the `ECS` deployment controller.

~~This plugin will not rollback for you. For rollbacks, use a [deployment circuit breaker](https://aws.amazon.com/blogs/containers/announcing-amazon-ecs-deployment-circuit-breaker/).~~

## Requirements

The ECS Service being deployed to must use the `ECS` deployment controller.

### IAM

- `iam:PassRole` on any Task Role (container) or Task Execution Role (ECS Agent) defined in any Task Definition that this tool will modify
  - You might consider using tag-based access control if there are a lot of roles Drone must be able to pass
- `ecs:DescribeTaskDefinition` on any task definitions this tool will modify
- `ecs:DescribeServices` on any services this tool will modify
- `ecs:UpdateService` on any services this tool will modify 
- `ecs:RegisterTaskDefinition` on `*`


## Example usage

```yaml
---
kind: pipeline
name: deploy

steps:
- name: deploy
  image: public.ecr.aws/assemblyai/drone-deploy-ecs
  settings:
    # Can either be rolling or blue-green
    mode: rolling
    aws_region: us-east-2
    # The name of the ECS service
    service: webapp
    # The name of the ECS cluster that the service is in
    cluster: dev-ecs-cluster
    # The name of the container to update
    container: nginx
    # The image to deploy
    image: myorg/nginx-${DRONE_COMMIT_SHA}
    max_deploy_checks: 10
```

## Blue / Green

The ECS service must have an associated Application Autoscaling Target

The deployment type must be ECS
