# drone-deploy-ecs

## Overview

`drone-deploy-ecs` is an opinionated Drone plugin for updating a single container within an ECS Task.

This plugin has support for two deployment modes: rolling and blue / green.

During a rolling deployment, the plugin retrieves the active Task Definition for a specified ECS Service, creates a new revision of the Task Definition with an updated image for a specified container, updates the Service to use the new Task Definition, and waits for the deployment to complete.

A blue / green deployment is similar to a rolling deployment. The key difference is that once the number of running green tasks matches the number of desired green tasks, the blue service is scaled down. It's important to note that this plugin _only_ uses desired vs running to determine deployment health. It will not check the health of a target in a target group, for example

[ECR Link](https://gallery.ecr.aws/assemblyai/drone-deploy-ecs)

## Important Notes

This plugin cannot update multiple containers within the same Task Definition simultaneously. It will only update the image for a single container within a Task Defintion

The ECS Service must use the `ECS` deployment controller.


## Requirements

The ECS Service being deployed to must use the `ECS` deployment controller.

### IAM

- `iam:PassRole` on any Task Role (container) or Task Execution Role (ECS Agent) defined in any Task Definition that this tool will modify
  - You might consider using tag-based access control if there are a lot of roles Drone must be able to pass
- `ecs:DescribeTaskDefinition` on any task definitions this tool will modify
- `ecs:DescribeServices` on any services this tool will modify
- `ecs:UpdateService` on any services this tool will modify
- `ecs:ListTasks` on `*`
- `ecs:DescribeTasks` on `*`
- `ecs:RegisterTaskDefinition` on `*`
- `application-autoscaling:DescribeScalableTargets` on `*`
- `application-autoscaling:RegisterScalableTarget` on `*` if you plan on using a blue/green deployment

## Example usage

### Rolling Deployment

```yaml
---
kind: pipeline
name: deploy

steps:
- name: deploy
  image: public.ecr.aws/assemblyai/drone-deploy-ecs
  settings:
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

You can deploy to multiple services in the same step by declaring a comma-separated list of `settings.service`. This is beneficial if you have multiple services using the same task definition.

For example:

```yml
steps:
- name: deploy
  image: public.ecr.aws/assemblyai/drone-deploy-ecs
  settings:
    mode: rolling
    aws_region: us-east-2
    service: webapp,webapp-spot
    cluster: dev-ecs-cluster
    container: nginx
    image: myorg/nginx-${DRONE_COMMIT_SHA}
    max_deploy_checks: 10
```

#### Disabling rollbacks

You can disable rollbacks by setting the `disable_rollbacks` to any string. Simply omit it to enable rollbacks. You may want to disable rollbacks if you have the ECS Circuit Breaker enabled for your service.

### Blue / Green

Blue / Green deployments will work with services that use Application Autoscaling and those that do not.

One service must have a desired count of 0, the other must have a desired count > 0.

It does not matter which service is set for `blue_service` or `green_service`. The plugin will use the service with a desired count of 0 as the green service. This is simply a way to define which two services the plugin should modify. 

Once the number of running containers equals the number of desired containers for the green service, the plugin will begin scaling down the blue service by  `scale_down_percent`. 

Blue / Green deployments do not support disabling rollbacks.

```yml
---
kind: pipeline
name: deploy

steps:
- name: deploy
  image: public.ecr.aws/assemblyai/drone-deploy-ecs
  settings:
    mode: blue-green
    aws_region: us-east-2
    # The name of the green ECS service
    green_service: webapp-green
    # The name of the blue ECS service
    blue_service: webapp-blue
    # The name of the ECS cluster that the service is in
    cluster: dev-ecs-cluster
    # The name of the container to update
    container: nginx
    # The image to deploy
    image: myorg/nginx-${DRONE_COMMIT_SHA}
    # How many times to check rollout status before failing
    max_deploy_checks: 10
    # Percent of instances to scale down blue service by
    scale_down_percent: 50
    # Seconds to wait between scale down events
    scale_down_interval: 600
    # Number of seconds between scaling up green service and scaling down blue
    # This is useful if your application takes some time to become healthy
    scale_down_wait_period: 10
    # Number of times running count must equal desired count in order to mark green deployment as a success
    checks_to_pass: 2
```


## TODO

- Code cleanup
- Better `settings` documentation
- Tests
- Better, more consistent logging
- Update `pkg/deploy` functions to use `deploy.DeployConfig`