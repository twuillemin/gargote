package runner

import (
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
func RunStage(testIndex uint, stageIndex uint, stage definition.Stage) error {

	log.Infof("Stage %v-%v: starting ", testIndex, stageIndex)

	start := time.Now()

	var err error
	success := false
	maxTries := 1 + int(stage.MaximumRetries)
	tryNumber := 0

	variables := make(map[string]interface{})

	for ; tryNumber < maxTries && !success; tryNumber++ {

		if stage.DelayBefore > 0 {
			time.Sleep(time.Duration(stage.DelayBefore) * time.Millisecond)
		}

		for actionIndex, action := range stage.Actions {
			err = RunAction(testIndex, stageIndex, uint(actionIndex), action, variables)
			if err != nil {
				log.Warnf("End of try %v of stage", tryNumber)
				break
			}
		}

		// Even if an error was raised, wait as other tests may rely on the wait timing
		if stage.DelayAfter > 0 {
			time.Sleep(time.Duration(stage.DelayAfter) * time.Millisecond)
		}

		// If no errors raised, prepare to leave the loop
		if err == nil {
			success = true
		}
	}

	elapsed := time.Since(start)

	log.Infof("Stage %v-%v: Finished - total duration: %v, %v try(ies) (including %v ms of delay)", testIndex, stageIndex, elapsed, tryNumber, (stage.DelayBefore+stage.DelayAfter)*uint(tryNumber))

	return err
}
