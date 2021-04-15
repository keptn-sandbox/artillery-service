package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2" // make sure to use v2 cloudevents here
	keptn "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

/**
* Here are all the handler functions for the individual event
* See https://github.com/keptn/spec/blob/0.8.0-alpha/cloudevents.md for details on the payload
**/

// GenericLogKeptnCloudEventHandler is a generic handler for Keptn Cloud Events that logs the CloudEvent
func GenericLogKeptnCloudEventHandler(myKeptn *keptnv2.Keptn, incomingEvent cloudevents.Event, data interface{}) error {
	log.Printf("Handling %s Event: %s", incomingEvent.Type(), incomingEvent.Context.GetID())
	log.Printf("CloudEvent %T: %v", data, data)

	return nil
}

func getServiceURL(data *keptnv2.TestTriggeredEventData) (*url.URL, error) {
	if len(data.Deployment.DeploymentURIsPublic) > 0 && data.Deployment.DeploymentURIsPublic[0] != "" {
		return url.Parse(data.Deployment.DeploymentURIsPublic[0])
	} else if len(data.Deployment.DeploymentURIsLocal) > 0 && data.Deployment.DeploymentURIsLocal[0] != "" {
		return url.Parse(data.Deployment.DeploymentURIsLocal[0])
	}

	return nil, errors.New("no deployment URI included in event")
}

// getKeptnResource fetches a resource from Keptn config repo and stores it in a temp directory
func getKeptnResource(myKeptn *keptnv2.Keptn, resourceName string, tempDir string) (string, error) {
	requestedResourceContent, err := myKeptn.GetKeptnResource(resourceName)

	if err != nil {
		fmt.Printf("Failed to fetch file: %s\n", err.Error())
		return "", err
	}

	targetFileName := fmt.Sprintf("%s/%s", tempDir, "artillery-scenario.yaml")

	resourceFile, err := os.Create(targetFileName)
	defer resourceFile.Close()

	_, err = resourceFile.Write([]byte(requestedResourceContent))

	if err != nil {
		fmt.Printf("Failed to create tempfile: %s\n", err.Error())
		return "", err
	}

	return targetFileName, nil
}

// OldHandleConfigureMonitoringEvent handles old configure-monitoring events
// TODO: add in your handler code
func OldHandleConfigureMonitoringEvent(myKeptn *keptnv2.Keptn, incomingEvent cloudevents.Event, data *keptn.ConfigureMonitoringEventData) error {
	log.Printf("Handling old configure-monitoring Event: %s", incomingEvent.Context.GetID())

	return nil
}

// HandleTestTriggeredEvent handles test.triggered events
func HandleTestTriggeredEvent(myKeptn *keptnv2.Keptn, incomingEvent cloudevents.Event, data *keptnv2.TestTriggeredEventData) error {
	log.Printf("Handling test.triggered Event: %s", incomingEvent.Context.GetID())

	_, err := myKeptn.SendTaskFinishedEvent(&keptnv2.EventData{}, ServiceName)

	if err != nil {
		errMsg := fmt.Sprintf("Failed to send task started CloudEvent (%s), aborting...", err.Error())
		log.Println(errMsg)
		return err
	}

	serviceURL, err := getServiceURL(data)

	if err != nil {
		// report error
		log.Print(err)
		// send out a test.finished failed CloudEvent
		_, err = myKeptn.SendTaskFinishedEvent(&keptnv2.EventData{
			Status:  keptnv2.StatusErrored,
			Result:  keptnv2.ResultFailed,
			Message: err.Error(),
		}, ServiceName)
	}

	var artilleryScenario string

	if data.Test.TestStrategy == "performance" {
		artilleryScenario = "scenarios/load.yaml"
	} else if data.Test.TestStrategy == "functional" {
		artilleryScenario = "scenarios/basic.yaml"
	} else {
		artilleryScenario = "scenarios/health.yaml"
	}

	fmt.Printf("TestStrategy=%s -> testFile=%s, serviceUrl=%s\n", data.Test.TestStrategy, artilleryScenario, serviceURL.String())

	// create a tempdir
	tempDir, err := ioutil.TempDir("", "artillery")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	artilleryScenarioResourceLocal, err := getKeptnResource(myKeptn, artilleryScenario, tempDir)

	if err != nil {
		// failed to fetch sli config file
		errMsg := fmt.Sprintf("Failed to fetch artillery scenario %s from config repo: %s", artilleryScenario, err.Error())
		log.Println(errMsg)
		// send a get-sli.finished event with status=error and result=failed back to Keptn

		_, err = myKeptn.SendTaskFinishedEvent(&keptnv2.EventData{
			Status:  keptnv2.StatusErrored,
			Result:  keptnv2.ResultFailed,
			Message: errMsg,
		}, ServiceName)

		return err
	}

	// CAPTURE START TIME
	startTime := time.Now()

	// artillery run -t HOST SCENARIO_FILE
	str, err := keptn.ExecuteCommand("artillery", []string{
		"run",
		"-t",
		serviceURL.String(),
		artilleryScenarioResourceLocal})

	log.Print(str)

	if err != nil {
		// report error
		log.Print(err)
		// send out a test.finished failed CloudEvent
		_, err = myKeptn.SendTaskFinishedEvent(&keptnv2.EventData{
			Status:  keptnv2.StatusErrored,
			Result:  keptnv2.ResultFailed,
			Message: err.Error(),
		}, ServiceName)

		return err
	}

	endTime := time.Now()

	finishedEvent := &keptnv2.TestFinishedEventData{
		Test: keptnv2.TestFinishedDetails{
			Start: startTime.Format(time.RFC3339),
			End:   endTime.Format(time.RFC3339),
		},
		EventData: keptnv2.EventData{
			Result:  keptnv2.ResultPass,
			Status:  keptnv2.StatusSucceeded,
			Message: fmt.Sprintf("Artillery test [%s] finished successfully", artilleryScenario),
		},
	}

	// Finally: send out a test.finished CloudEvent
	_, err = myKeptn.SendTaskFinishedEvent(finishedEvent, ServiceName)

	if err != nil {
		errMsg := fmt.Sprintf("Failed to send task finished CloudEvent (%s), aborting...", err.Error())
		log.Println(errMsg)
		return err
	}

	return nil
}
