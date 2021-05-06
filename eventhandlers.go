package main

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2" // make sure to use v2 cloudevents here
	keptn "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

// Artillery configuration file path
const (
	// ArtilleryConfFilename defines the path to the artillery.conf.yaml
	ArtilleryConfFilename = "scenarios/artillery.conf.yaml"
	// DefaultArtilleryFilename defines the path to the default load.yaml
	DefaultArtilleryFilename = "scenarios/load.yaml"
)

// ArtilleryConf Configuration file type
type ArtilleryConf struct {
	SpecVersion string      `json:"spec_version" yaml:"spec_version"`
	Workloads   []*Workload `json:"workloads" yaml:"workloads"`
}

// Workload of Keptn stage
type Workload struct {
	TestStrategy string `json:"teststrategy" yaml:"teststrategy"`
	Script       string `json:"script" yaml:"script"`
}

// Loads artillery.conf.yaml for the current service
func getArtilleryConf(myKeptn *keptnv2.Keptn, project string, stage string, service string) (*ArtilleryConf, error) {
	var err error

	log.Printf("Loading %s for %s.%s.%s", ArtilleryConfFilename, project, stage, service)

	keptnResourceContent, err := myKeptn.GetKeptnResource(ArtilleryConfFilename)

	if err != nil {
		logMessage := fmt.Sprintf("error when trying to load %s file for service %s on stage %s or project-level %s: %s", ArtilleryConfFilename, service, stage, project, err.Error())
		return nil, errors.New(logMessage)
	}
	if len(keptnResourceContent) == 0 {
		// if no artillery.conf.yaml file is available, this is not an error, as the service will proceed with the default workload
		log.Printf("no %s found", ArtilleryConfFilename)
		return nil, nil
	}

	var artilleryConf *ArtilleryConf
	artilleryConf, err = parseArtilleryConf([]byte(keptnResourceContent))
	if err != nil {
		logMessage := fmt.Sprintf("Couldn't parse %s file found for service %s in stage %s in project %s. Error: %s", ArtilleryConfFilename, service, stage, project, err.Error())
		return nil, errors.New(logMessage)
	}

	log.Printf("Successfully loaded artillery.conf.yaml with %d workloads", len(artilleryConf.Workloads))

	return artilleryConf, nil
}

// parses content and maps it to the ArtilleryConf struct
func parseArtilleryConf(input []byte) (*ArtilleryConf, error) {
	artilleryconf := &ArtilleryConf{}
	err := yaml.Unmarshal(input, &artilleryconf)
	if err != nil {
		return nil, err
	}

	return artilleryconf, nil
}

// GenericLogKeptnCloudEventHandler is a generic handler for Keptn Cloud Events that logs the CloudEvent
func GenericLogKeptnCloudEventHandler(myKeptn *keptnv2.Keptn, incomingEvent cloudevents.Event, data interface{}) error {
	log.Printf("Handling %s Event: %s", incomingEvent.Type(), incomingEvent.Context.GetID())
	log.Printf("CloudEvent %T: %v", data, data)

	return nil
}

//
// returns the service URL that is either passed via the DeploymentURI* parameters or constructs one based on keptn naming structure
//
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

	// Cut away folders from the path (if there are any)
	path := strings.Split(resourceName, "/")

	targetFileName := fmt.Sprintf("%s/%s", tempDir, path[len(path)-1])

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

	_, err := myKeptn.SendTaskStartedEvent(&keptnv2.EventData{}, ServiceName)

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

	var artilleryconf *ArtilleryConf
	artilleryconf, err = getArtilleryConf(myKeptn, myKeptn.Event.GetProject(), myKeptn.Event.GetStage(), myKeptn.Event.GetService())

	if err != nil {
		log.Println(err)
	}

	var artilleryFilename = ""

	if artilleryconf != nil {
		for _, workload := range artilleryconf.Workloads {
			if workload.TestStrategy == data.Test.TestStrategy {
				if workload.Script != "" {
					artilleryFilename = workload.Script
				} else {
					artilleryFilename = ""
				}
			}
		}
	} else {
		artilleryFilename = DefaultArtilleryFilename
		fmt.Println("No artillery.conf.yaml file provided. Continuing with default settings!")
	}

	fmt.Printf("TestStrategy=%s -> testFile=%s, serviceUrl=%s\n", data.Test.TestStrategy, artilleryFilename, serviceURL.String())

	// create a tempdir
	tempDir, err := ioutil.TempDir("", "artillery")
	if err != nil {
		log.Fatal(err)
	}
	//defer os.RemoveAll(tempDir)

	var artilleryResourceFilenameLocal = ""
	if artilleryFilename != "" {
		artilleryResourceFilenameLocal, err = getKeptnResource(myKeptn, artilleryFilename, tempDir)

		// FYI you do not need to "fail" if sli.yaml is missing, you can also assume smart defaults like we do
		// in keptn-contrib/dynatrace-service and keptn-contrib/prometheus-service
		if err != nil {
			// failed to fetch sli config file
			errMsg := fmt.Sprintf("Failed to fetch artillery file %s from config repo: %s", artilleryFilename, err.Error())
			log.Println(errMsg)
			// send a get-sli.finished event with status=error and result=failed back to Keptn

			_, err = myKeptn.SendTaskFinishedEvent(&keptnv2.EventData{
				Status:  keptnv2.StatusErrored,
				Result:  keptnv2.ResultFailed,
				Message: errMsg,
			}, ServiceName)

			return err
		}

		log.Println("Successfully fetched artillery test file")
	}

	// CAPTURE START TIME
	startTime := time.Now()

	outputDestination, _ := ioutil.TempFile("", "stats")
	defer os.Remove(outputDestination.Name())

	var endTime time.Time

	if artilleryResourceFilenameLocal == "" {
		log.Println("No test file provided for stage -> Skipping tests")
	} else {
		err = runArtillery(artilleryResourceFilenameLocal, serviceURL.String(), outputDestination.Name())

		endTime = time.Now()

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

		artilleryRunErrors, _ := getScenarioErrors(outputDestination)

		if len(artilleryRunErrors) != 0 {
			myKeptn.SendTaskFinishedEvent(&keptnv2.TestFinishedEventData{
				Test: keptnv2.TestFinishedDetails{
					Start: startTime.Format(time.RFC3339),
					End:   endTime.Format(time.RFC3339),
				},
				EventData: keptnv2.EventData{
					Result:  keptnv2.ResultFailed,
					Status:  keptnv2.StatusSucceeded,
					Message: fmt.Sprintf("Artillery test [%s] failed: %v", artilleryFilename, artilleryRunErrors),
				},
			}, ServiceName)

			return nil
		}
	}

	finishedEvent := &keptnv2.TestFinishedEventData{
		Test: keptnv2.TestFinishedDetails{
			Start: startTime.Format(time.RFC3339),
			End:   endTime.Format(time.RFC3339),
		},
		EventData: keptnv2.EventData{
			Result:  keptnv2.ResultPass,
			Status:  keptnv2.StatusSucceeded,
			Message: fmt.Sprintf("Artillery test [%s] finished successfully", artilleryFilename),
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
