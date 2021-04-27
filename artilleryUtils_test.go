package main

import (
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func TestGetScenarioErrors(t *testing.T) {
	testFile := getTestFile(`{"timestamp":"2021-04-18T15:11:00.022Z","scenariosCreated":4,"scenariosCompleted":0,"requestsCompleted":8,"latency":{"min":24.5,"max":83.4,"median":44.9,"p95":83.4,"p99":83.4},"rps":{"count":8,"mean":1.01},"scenarioDuration":{"min":null,"max":null,"median":null,"p95":null,"p99":null},"scenarioCounts":{"0":4},"errors":{"Failed expectations for request http://carts.sockshop-dev.io/carts/1":4},"codes":{"200":8},"matches":0,"latencies":[83416101,25693521,65971341,25508039,64129390,24503857,67613131,24709839],"customStats":{},"counters":{},"concurrency":0,"pendingRequests":0,"scenariosAvoided":0}
{"timestamp":"2021-04-18T15:11:10.021Z","scenariosCreated":5,"scenariosCompleted":0,"requestsCompleted":10,"latency":{"min":24.5,"max":69.6,"median":46.4,"p95":69.6,"p99":69.6},"rps":{"count":10,"mean":1},"scenarioDuration":{"min":null,"max":null,"median":null,"p95":null,"p99":null},"scenarioCounts":{"0":5},"errors":{"Failed expectations for request http://carts.sockshop-dev.io/carts/1":5},"codes":{"200":10},"matches":0,"latencies":[64700220,28133657,69617153,27145386,67411077,24465594,67305722,25206591,66440882,24727293],"customStats":{},"counters":{},"concurrency":0,"pendingRequests":0,"scenariosAvoided":0}
{"timestamp":"2021-04-18T15:11:10.527Z","scenariosCreated":1,"scenariosCompleted":0,"requestsCompleted":2,"latency":{"min":24.4,"max":63.4,"median":43.9,"p95":63.4,"p99":63.4},"rps":{"count":2,"mean":4},"scenarioDuration":{"min":null,"max":null,"median":null,"p95":null,"p99":null},"scenarioCounts":{"0":1},"errors":{"Failed expectations for request http://carts.sockshop-dev.io/carts/1":1},"codes":{"200":2},"matches":0,"latencies":[63433160,24365387],"customStats":{},"counters":{},"concurrency":0,"pendingRequests":0,"scenariosAvoided":0}`)
	defer os.Remove(testFile.Name())

	t.Run("Extracts the errors from an artillery run with expectations plugin enabled", func(t *testing.T) {
		expectedErrors := map[string]float64{"Failed expectations for request http://carts.sockshop-dev.io/carts/1": 10}
		runErrors, _ := getScenarioErrors(testFile)

		if !reflect.DeepEqual(runErrors, expectedErrors) {
			t.Errorf("Got %v, wanted %v", runErrors, expectedErrors)
		}
	})
}

func getTestFile(content string) *os.File {
	testFile, _ := ioutil.TempFile("", "artillery-utils")

	testFile.Write([]byte(content))
	testFile.Seek(0, io.SeekStart)

	return testFile
}
