package main

import (
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/twuillemin/gargote/pkg/definition"
	"github.com/twuillemin/gargote/pkg/runner"
	"gopkg.in/yaml.v3"
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

	log.Info("Starting...")

	//generateTemp()

	test, err := readFile(fileName)
	if err != nil {
		log.Fatal(err)
	}

	err = runner.RunTest(*test)
	if err != nil {
		log.Errorf("Tests finished with error: %v", err)
		return
	}

	log.Info("Finished successfully...\n")
}

func readFile(fileName string) (*definition.Test, error) {

	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	var test definition.Test
	err = yaml.Unmarshal(data, &test)

	if err != nil {
		return nil, err
	}

	return &test, nil
}

// For now keep the following function just in case
/*
func generateExample() []byte {

	t := definition.Test{
		TestName: "The typicode API",
		Stages: []definition.Stage{
			{
				Name:           "Basic API usage",
				MaximumRetries: 0,
				DelayBefore:    100,
				DelayAfter:     100,
				Actions: []definition.Action{
					{
						Name: "Get a to-do",
						Query: definition.Query{
							URL:    "https://jsonplaceholder.typicode.com/todos/1",
							Method: definition.GET,
							Headers: map[string]string{
								"Accept": "application/json",
							},
						},
						Response: definition.Response{
							Validation: definition.Validation{
								StatusCodes: []uint{200},
								Headers: map[string]string{
									"Connection": "keep-alive",
								},
							},
							Capture: definition.Capture{
								BodyJSON: map[string]string{
									"userId": "the_user_id",
								},
							},
						},
					},
					{
						Name: "Get a user",
						Query: definition.Query{
							URL:    "https://jsonplaceholder.typicode.com/users/{{ .the_user_id }}",
							Method: definition.GET,
							Headers: map[string]string{
								"Accept": "application/json",
							},
						},
						Response: definition.Response{
							Validation: definition.Validation{
								StatusCodes: []uint{200},
								BodyJSON: map[string]interface{}{
									"company.name": "Romaguera-Crona",
								},
							},
						},
					},
				},
			},
		},
	}

	y, err := yaml.Marshal(t)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(y))

	return y
}
*/
