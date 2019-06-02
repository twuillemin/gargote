package runner

import (
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/twuillemin/gargote/pkg/definition"
)

var currentNumberOfRunningTests = 0
var maximumNumberOfRunningTests = 0

// RunTest executes a Test.
//
// Params:
//  - test: the Test to execute
//
// Return an error if the action fail, nil otherwise
func RunTest(test definition.Test) error {

	fmt.Printf("====================================================\n")
	fmt.Printf("=\n")
	fmt.Printf("=         Test: %s \n", test.TestName)
	fmt.Printf("=\n")
	fmt.Printf("====================================================\n")

	var wg sync.WaitGroup

	wg.Add(int(test.Swarm.NumberOfRuns))

	intervalBetweenStart := uint(1000 / test.Swarm.CreationRate)
	currentNumberOfRunningTests = 0

	ticker := time.NewTicker(time.Duration(intervalBetweenStart) * time.Millisecond)

	start := time.Now()

	for index := 0; index < int(test.Swarm.NumberOfRuns); index++ {

		<-ticker.C

		go func(t definition.Test, i int) {
			defer wg.Done()
			runSingleTest(t, i)
		}(test, index)
	}

	wg.Wait()

	fmt.Printf("All tests total duration: %v\n", time.Since(start))

	return nil
}

// GetCurrentNumberOfRunningTests returns the current number of test running in parallel. Should be zero when
// no test is running.
//
// Return the current number of test running in parallel
func GetCurrentNumberOfRunningTests() int {
	return currentNumberOfRunningTests
}

// GetMaximumNumberOfRunningTests returns the maximum number of test running in parallel. It is the maximum of
// GetCurrentNumberOfRunningTests.
//
// Return the maximum number of test running in parallel
func GetMaximumNumberOfRunningTests() int {
	return maximumNumberOfRunningTests
}

func runSingleTest(test definition.Test, testIndex int) {

	log.Infof("Test %v: starting ", testIndex)

	currentNumberOfRunningTests++
	if currentNumberOfRunningTests > maximumNumberOfRunningTests {
		maximumNumberOfRunningTests = currentNumberOfRunningTests
	}

	start := time.Now()

	for stageIndex, stage := range test.Stages {
		if err := RunStage(testIndex, stageIndex, stage); err != nil && !test.ContinueOnStageFailure {
			log.Infof("Test %v: ending prematurely due to error in stage", testIndex)
			break
		}
	}

	elapsed := time.Since(start)

	currentNumberOfRunningTests--

	log.Infof("Test %v: total duration: %v", testIndex, elapsed)
}
