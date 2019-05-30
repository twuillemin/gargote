package runner

import (
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/twuillemin/gargote/pkg/definition"
)

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

	ticker := time.NewTicker(time.Duration(intervalBetweenStart) * time.Millisecond)

	start := time.Now()

	for index := uint(0); index < test.Swarm.NumberOfRuns; index++ {

		<-ticker.C

		go func(t definition.Test, i uint) {
			defer wg.Done()
			runSingleTest(t, i)
		}(test, index)
	}

	wg.Wait()

	fmt.Printf("All tests total duration: %v\n", time.Since(start))

	return nil
}

func runSingleTest(test definition.Test, testIndex uint) {

	log.Infof("Test %v: starting ", testIndex)

	start := time.Now()

	for stageIndex, stage := range test.Stages {
		if err := RunStage(testIndex, uint(stageIndex), stage); err != nil && !test.ContinueOnStageFailure {
			log.Warnf("Test %v: ending prematurely due to error in stage", testIndex)
			break
		}
	}

	elapsed := time.Since(start)
	log.Infof("Test %v: total duration: %v", testIndex, elapsed)
}
