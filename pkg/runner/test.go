package runner

import (
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

	log.Info("====================================================")
	log.Info("=")
	log.Infof("=           Start test: %s ", test.TestName)
	log.Info("=")
	log.Info("====================================================")

	start := time.Now()

	var err error

	for _, stage := range test.Stages {
		err = RunStage(stage)
		if err != nil {
			log.Warn("End of test due to error in stage")
			break
		}
	}

	elapsed := time.Since(start)
	log.Infof("Test total duration: %v", elapsed)

	return err
}
