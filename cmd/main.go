package main

import (
	"fmt"
	"github.com/montanaflynn/stats"
	"github.com/twuillemin/gargote/pkg/db"
	"github.com/twuillemin/gargote/pkg/definition"
	"github.com/twuillemin/gargote/pkg/loader"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/twuillemin/gargote/pkg/runner"
)

func main() {

	if len(os.Args) != 2 {
		log.Fatal("gargote need the script file name as argument")
	}

	fileName := os.Args[1]

	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: false,
	})

	log.SetLevel(log.WarnLevel)

	log.Info("Starting...")

	// Load test scenario
	test, err := loader.LoadFromFile(fileName)
	if err != nil {
		log.Fatal(err)
	}

	// Create the database
	err = db.CreateDatabase()
	if err != nil {
		log.Fatal(err)
	}

	// Display load during test
	quitDisplayLoadChannel := make(chan struct{})
	go displayLoad(quitDisplayLoadChannel)

	// Run the tests
	err = runner.RunTest(*test)
	if err != nil {
		log.Errorf("Tests finished with error: %v", err)
		return
	}

	log.Info("Finished successfully...\n")

	quitDisplayLoadChannel <- struct{}{}

	displayResults(*test)
}

func displayLoad(quitChannel chan struct{}) {

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fmt.Printf("%v / %v\n", runner.GetCurrentNumberOfRunningTests(), runner.GetMaximumNumberOfRunningTests())
			//fmt.Printf("%v\n", runner.GetCurrentNumberOfRunningTests())
		case <-quitChannel:
			return
		}
	}
}

func displayResults(test definition.Test) {

	results, err := db.GetAllRequests()
	if err != nil {
		log.Errorf("error while retrieving results: %v", err)
		return
	}

	for requestID, requestResult := range results {
		url := test.Stages[requestID.StageIndex].Actions[requestID.ActionIndex].Query.URL
		fmt.Printf("URL: %v, Fail: %v, Success: %v, Stats:[%v]\n", url, requestResult.NbFailure, len(requestResult.SuccessNanoTimes), getStatistics(requestResult.SuccessNanoTimes))
	}
}

func getStatistics(rawData []int) string {

	// Get the data in milliseconds (drop nano)
	data := make([]float64, len(rawData))
	for i, d := range rawData {
		data[i] = float64(int(d/1000)) / 1000.0
	}

	min, _ := stats.Min(data)
	max, _ := stats.Max(data)
	mean, _ := stats.Mean(data)
	median, _ := stats.Median(data)
	variance, _ := stats.StandardDeviation(data)

	return fmt.Sprintf(
		"min: %v, max: %v, mean: %v, median: %v, variance: %v",
		min,
		max,
		mean,
		median,
		variance)
}
