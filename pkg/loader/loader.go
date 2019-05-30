package loader

import (
	"github.com/twuillemin/gargote/pkg/definition"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

// LoadFromFile loads a Test from a file. The test is checked and if needed some sane default values
// are set
//
// Params:
//  - fileName: the name of the file to load
//
// Return a Test object and an error if the loading fail
func LoadFromFile(fileName string) (*definition.Test, error) {

	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	var test definition.Test
	err = yaml.Unmarshal(data, &test)

	if err != nil {
		return nil, err
	}

	return validateAndFix(&test)
}

func validateAndFix(test *definition.Test) (*definition.Test, error) {

	if test.Swarm.CreationRate == 0 {
		test.Swarm.CreationRate = 1
	}

	if test.Swarm.NumberOfRuns == 0 {
		test.Swarm.NumberOfRuns = 1
	}

	return test, nil
}
