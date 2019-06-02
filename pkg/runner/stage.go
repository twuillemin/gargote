package runner

import (
	"github.com/twuillemin/gargote/pkg/db"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/twuillemin/gargote/pkg/definition"
)

// RunStage executes a single Stage.
//
// Params:
//  - testIndex: the test number
//  - stageIndex: the stage number
//  - stage: the Stage to execute
//
// Return an error if the action fail, nil otherwise
func RunStage(testIndex int, stageIndex int, stage definition.Stage) error {

	log.Infof("Stage %v-%v: starting ", testIndex, stageIndex)

	start := time.Now()

	var err error
	success := false
	maxTries := 1 + int(stage.MaximumRetries)
	tryNumber := 0

	// Run the stages n-times until success
	for ; tryNumber < maxTries && !success; tryNumber++ {

		err = runStageOneTime(testIndex, stageIndex, stage)

		// If no errors raised, prepare to leave the loop
		if err == nil {
			success = true
		} else {
			log.Warnf("Stage %v-%v: try %v terminated in error", testIndex, stageIndex, tryNumber)
		}

	}

	elapsed := time.Since(start)

	log.Infof("Stage %v-%v: Finished - total duration: %v, %v try(ies) (including %v ms of delay)", testIndex, stageIndex, elapsed, tryNumber, (stage.DelayBefore+stage.DelayAfter)*uint(tryNumber))

	return err
}

func runStageOneTime(testIndex int, stageIndex int, stage definition.Stage) error {

	// variables will store the stage variables
	variables := make(map[string]interface{})

	// results will store the result for each action
	results := make([]*db.ActionEntry, 0, len(stage.Actions))

	if stage.DelayBefore > 0 {
		time.Sleep(time.Duration(stage.DelayBefore) * time.Millisecond)
	}

	// Define err out of the loop so that it can be returned
	var err error

	// For each action of the stage
	for actionIndex, action := range stage.Actions {

		startTime := time.Now()

		// Execute the action
		err = RunAction(testIndex, stageIndex, actionIndex, action, variables)

		// If no error
		if err == nil {

			// Keep the result
			results = append(results, &db.ActionEntry{
				TestIndex:    testIndex,
				StageIndex:   stageIndex,
				ActionIndex:  actionIndex,
				TimeNano:     startTime.Nanosecond(),
				DurationNano: int(time.Since(startTime).Nanoseconds()),
				Success:      true,
			})

		} else {

			// Keep the result
			results = append(results, &db.ActionEntry{
				TestIndex:    testIndex,
				StageIndex:   stageIndex,
				ActionIndex:  actionIndex,
				TimeNano:     startTime.Nanosecond(),
				DurationNano: 0,
				Success:      false,
			})

			// Stop the loop
			break
		}
	}

	// Even if an error was raised, wait as other tests may rely on the wait timing
	if stage.DelayAfter > 0 {
		time.Sleep(time.Duration(stage.DelayAfter) * time.Millisecond)
	}

	if saveErr := db.Insert(results); saveErr != nil {
		log.Errorf("unable to save stage results due to %v", saveErr)
	}

	// Return the last error found while executing the actions
	return err
}
