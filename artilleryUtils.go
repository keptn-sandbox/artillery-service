package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	keptn "github.com/keptn/go-utils/pkg/lib"
)

func getScenarioErrors(file *os.File) (map[string]float64, error) {
	errors := make(map[string]float64)
	decoder := json.NewDecoder(file)

	for {
		var data map[string]interface{}

		if err := decoder.Decode(&data); err == io.EOF {
			break
		} else if err != nil {
			fmt.Println(fmt.Errorf("Error: failed to unmarshal json %s", err.Error()))

			return errors, err
		}

		for key, value := range data["errors"].(map[string]interface{}) {
			if counter, exists := errors[key]; exists {
				errors[key] = value.(float64) + counter
			} else {
				errors[key] = value.(float64)
			}
		}
	}

	return errors, nil
}

func runArtillery(resource string, serviceUrl string, outputDestination string) error {
	fmt.Println(fmt.Sprintf("Artillery output: %s", outputDestination))

	args := []string{
		"run",
		"-t",
		serviceUrl,
		"--overrides",
		fmt.Sprintf("{\"config\": { \"plugins\": {\"save-stats\": { \"destination\": \"%s\" }}}}", outputDestination),
		resource,
	}

	// artillery run -t HOST SCENARIO_FILE
	_, err := keptn.ExecuteCommand("artillery", args)

	return err
}
