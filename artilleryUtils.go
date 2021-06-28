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

func runArtillery(resource string, serviceURL string, outputDestination string) (string, error) {
	args := []string{
		"run",
		"-t",
		serviceURL,
		"--overrides",
		fmt.Sprintf("{\"config\": { \"plugins\": {\"save-stats\": { \"destination\": \"%s\" }}}}", outputDestination),
		resource,
	}

	return keptn.ExecuteCommand("artillery", args)
}
