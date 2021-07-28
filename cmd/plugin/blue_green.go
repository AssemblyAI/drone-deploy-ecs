package main

import (
	"context"
	"errors"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/assemblyai/drone-deploy-ecs/pkg/deploy"
	"github.com/assemblyai/drone-deploy-ecs/pkg/types"
)

var (
	initialDesiredCount int
)

// Returns blue service, green service, error
func determineBlueGreen(e types.ECSClient, blueService string, greenService string, cluster string) (string, string, error) {

	blueCount, err := deploy.GetServiceDesiredCount(context.Background(), e, blueService, cluster)

	if err != nil {
		log.Println("Error retrieving desired count for blue service", err.Error())
		return "", "", errors.New("deploy failed")
	}

	greenCount, err := deploy.GetServiceDesiredCount(context.Background(), e, greenService, cluster)

	if err != nil {
		log.Println("Error retrieving desired count for blue service", err.Error())
		return "", "", errors.New("deploy failed")
	}

	log.Printf("Service '%s' has a desired count of '%d'\n", blueService, blueCount)
	log.Printf("Service '%s' has a desired count of '%d'\n", greenService, greenCount)

	if blueCount == 0 {
		return greenService, blueService, nil
	}

	if greenCount == 0 {
		return blueService, greenService, nil
	}

	log.Println("Unable to determine which service is blue and which is green")
	log.Printf("Service '%s' has %d desired replicas while service '%s' has %d desired replicas. One of these should be 0\n", blueService, blueCount, greenService, greenCount)
	return "", "", errors.New("reconcile error")
}

func blueGreen(dc deploy.DeployConfig, maxDeployChecks int) error {
	log.Println("Beginning blue green deployment")

	blueServiceName := os.Getenv("PLUGIN_BLUE_SERVICE")
	greenServiceName := os.Getenv("PLUGIN_GREEN_SERVICE")

	determinedBlueService, determinedGreenService, err := determineBlueGreen(dc.ECS, blueServiceName, greenServiceName, dc.Cluster)

	if err != nil {
		return err
	}

	log.Printf("Determined service '%s' is blue and '%s' is green\n", determinedBlueService, determinedGreenService)

	td, err := deploy.GetServiceRunningTaskDefinition(context.TODO(), dc.ECS, determinedBlueService, dc.Cluster)

	if err != nil {
		log.Println("Failing because of an error determining the currently in-use task definition")
		return err
	}

	currTD, err := deploy.RetrieveTaskDefinition(context.TODO(), dc.ECS, td)

	if err != nil {
		log.Println("Failing because of an error retrieving the currently in-use task definition")
		return err
	}

	newTD, err := deploy.CreateNewTaskDefinitionRevision(context.TODO(), dc.ECS, currTD, dc.Container, dc.Image)

	if err != nil {
		log.Println("Failing because of an error retrieving the creating a new task definition revision")
		return err
	}

	log.Println("Created new task definition revision", newTD.Revision)

	currBlueDesiredCount, err := deploy.GetServiceDesiredCount(context.Background(), dc.ECS, determinedBlueService, dc.Cluster)

	if err != nil {
		log.Println("Failing because of an error determining desired count for blue service", err.Error())
		return err
	}

	// There is no deployment ID so discard it
	_, err = deploy.UpdateServiceTaskDefinitionVersion(context.TODO(), dc.ECS, determinedGreenService, dc.Cluster, *newTD.TaskDefinitionArn)

	if err != nil {
		log.Println("Error updating task definition for service", err.Error())
		return errors.New("deploy failed")
	}

	serviceUsesAppAutoscaling, err := deploy.AppAutoscalingTargetExists(context.Background(), dc.AppAutoscaling, dc.Cluster, determinedBlueService)

	if err != nil {
		log.Println("Error determining if service uses application autoscaling", err.Error())
		return err
	}

	var serviceMaxCount int32
	var serviceMinCount int32

	if serviceUsesAppAutoscaling {
		log.Printf("Service '%s' uses application autoscaling. Will modify autoscaling max count", determinedGreenService)
		serviceMaxCount, serviceMinCount, err = deploy.GetServiceMinMaxCount(context.Background(), dc.AppAutoscaling, dc.Cluster, determinedBlueService)

		if err != nil {
			log.Println("Error determining service max count", err.Error())
			return err
		}
	} else {
		serviceMaxCount = -1
		serviceMinCount = 0
	}

	// Scale up green service to the same count as blue
	dc.ScaleUp(currBlueDesiredCount, serviceMinCount, serviceMaxCount, determinedGreenService)

	log.Println("Pausing for 45 seconds while ECS schedules", currBlueDesiredCount, "containers")
	initialDesiredCount = int(currBlueDesiredCount)
	time.Sleep(45 * time.Second)

	// Start polling deployment
	greenScaleupFinished := false
	deployCounter := 0
	successCounter := 0

	successCountThreshold, _ := strconv.Atoi(os.Getenv("PLUGIN_CHECKS_TO_PASS"))

	greenScaleupFinished, err = dc.GreenScaleUpFinished(context.Background(), determinedGreenService)

	if err != nil {
		log.Println("Error checking if green service has finished scaling", err.Error())
		return err
	}

	for {
		// Ensure we haven't surpassed the max check limit
		if deployCounter > maxDeployChecks {
			log.Println("Max deploy checks surpassed. Scaling green down and marking deployment a failure")
			dc.ScaleDown(0, 0, 0, determinedGreenService, serviceUsesAppAutoscaling)
			return errors.New("deploy failed")
		}

		if !greenScaleupFinished {
			// In this case, the service is not done scaling up
			// Increment counter first
			deployCounter++

			// Check if scale up has finished
			greenScaleupFinished, err = dc.GreenScaleUpFinished(context.Background(), determinedGreenService)

			if err != nil {
				log.Println("Error checking if green has finished scaling up", err.Error())
				return err
			}
			// Wait for 10 seconds
			log.Println("Waiting 10 seconds for green service to scale up")
			time.Sleep(10 * time.Second)
			// Reset successCounter, successful checks must be consecutive
			successCounter = 0

		} else {
			// In this case, running == desired
			// Now we need to make sure the healthy check threshold has been reached
			if successCounter < successCountThreshold {
				// We simply need to increment the counter here
				// because we already know that running == desired
				log.Println("Successful checks:", successCounter)
				// Wait for 10 seconds
				log.Println("Waiting 10 seconds before incrementing healthy checks")
				time.Sleep(10 * time.Second)
				successCounter++
			} else {
				// Again, running == desired
				// _and_ successCounter >= successCountThreshold
				log.Println("Green deployment has reached healthy check threshold")
				break
			}
		}

	}

	log.Printf("Green service '%s' finished scaling up! Scaling down blue service '%s'\n", determinedGreenService, determinedBlueService)

	log.Printf("Waiting %s seconds before scaling down blue", os.Getenv("PLUGIN_SCALE_DOWN_WAIT_PERIOD"))

	scaleDownPause, _ := strconv.Atoi(os.Getenv("PLUGIN_SCALE_DOWN_WAIT_PERIOD"))

	time.Sleep(time.Duration(scaleDownPause) * time.Second)

	err = scaleDownInPercentages(
		dc,
		determinedBlueService,
		serviceUsesAppAutoscaling,
		os.Getenv("PLUGIN_SCALE_DOWN_PERCENT"),
		os.Getenv("PLUGIN_SCALE_DOWN_INTERVAL"),
		int(currBlueDesiredCount),
	)

	return err
}

func scaleDownInPercentages(dc deploy.DeployConfig, service string, serviceUsesAppAutoscaling bool, scalePercent string, scaleDownInterval string, desiredCount int) error {
	scalePercentString, err := strconv.Atoi(scalePercent)

	if err != nil {
		log.Println("Error converting scale down percent", scalePercent, "to integer. Failing.")
		return err
	}

	scaleDownWait, err := strconv.Atoi(scaleDownInterval)

	if err != nil {
		log.Println("Error converting scale down interval", scaleDownInterval, "to integer. Failing.")
		return err
	}

	scalePercentAsFloat := float64(scalePercentString)

	// Convert int to decimal
	percent := float64(scalePercentAsFloat) / float64(100)

	log.Printf("Scaling down by %d percent\n", int(percent*100))

	var scaleDownBy int
	var newDesiredCount int32
	var lastScaleDownEvent bool

	scaleDownNumber := float64(initialDesiredCount) * percent

	if scaleDownNumber < 0 {
		scaleDownBy = 1
	} else if scaleDownNumber > 0 && scaleDownNumber < 1 {
		// Handle a bug where, if running count is less than 10
		// the desired count would be set to 0
		log.Println("Number of containers to remove is a decimal between 0 and 1. Removing one container")
		scaleDownBy = 1
	} else {
		// int() will give us a round number
		scaleDownBy = int(scaleDownNumber)
	}

	calculatedDesiredCount := desiredCount - scaleDownBy

	if calculatedDesiredCount <= 0 {
		newDesiredCount = 0
		lastScaleDownEvent = true
	} else {
		newDesiredCount = int32(calculatedDesiredCount)
		lastScaleDownEvent = false
	}

	err = dc.ScaleDown(newDesiredCount, 0, newDesiredCount, service, serviceUsesAppAutoscaling)

	if err != nil {
		log.Println("Error scaling down service", err.Error())
		return err
	}

	status, err := dc.GreenScaleUpFinished(context.Background(), service)

	if err != nil {
		log.Println("Error checking scale down status", err.Error())
		return err
	}

	for !status {
		status, err = dc.GreenScaleUpFinished(context.Background(), service)

		if err != nil {
			log.Println("Error checking scale down status", err.Error())
			return err
		}
		log.Println("Waiting 15 seconds for blue service to finish scaling down")
		time.Sleep(15 * time.Second)
	}

	log.Println("Finished scaling blue service down to", newDesiredCount)

	if lastScaleDownEvent {
		log.Println("Scale down complete")
		return nil
	} else {
		log.Println("Waiting", scaleDownWait, "seconds before scaling down again")
		time.Sleep(time.Duration(scaleDownWait) * time.Second)

		scaleDownInPercentages(dc, service, serviceUsesAppAutoscaling, scalePercent, scaleDownInterval, int(newDesiredCount))
	}

	return nil
}
