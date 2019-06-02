package runner

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/twuillemin/gargote/pkg/definition"
)

// RunAction executes a single Action.
//
// Params:
//  - testIndex: the test number
//  - stageIndex: the stage number
//  - actionIndex: The action number
//  - action: the Action to execute
//  - variables: the existing variables. Note that the map is updated is the action has Capture elements.
//
// Return an error if the action fail, nil otherwise
func RunAction(testIndex int, stageIndex int, actionIndex int, action definition.Action, variables map[string]interface{}) error {

	stageTitle := fmt.Sprintf("Action %v-%v-%v:", testIndex, stageIndex, actionIndex)

	log.Infof("%s starting %s", stageTitle, action.Name)

	// Use a default timeout of 1 minute if nothing specified
	timeout := 1 * time.Minute
	if action.Query.Timeout > 0 {
		timeout = time.Duration(action.Query.Timeout) * time.Millisecond
	}

	// Create the HTTP client with the requested parameters
	client := &http.Client{
		Timeout: timeout,
	}

	// Prepare the query
	req, err := prepareQuery(variables, action.Query)
	if err != nil {
		log.Warnf("%s ---> Error while preparing the query %v", stageTitle, err)
		return err
	}

	// Make the actual query
	resp, err := client.Do(req)
	if err != nil {
		log.Warnf("%s ---> Error while sending query %v", stageTitle, err)
		return err
	}

	if resp == nil {
		log.Warnf("%s ---> No response Received", stageTitle)
		return errors.New("no response received to query")
	}

	// Read the body as it is used by check and save
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Warnf("%s ---> Unable to read the body", stageTitle)
		return err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Errorf("%s unable to close the body of the response due to %v", stageTitle, err)
		}
	}()

	// Check the response
	if err = checkResponse(resp, body, action.Response.Validation); err != nil {
		log.Warnf("%s ---> Error while checking the response: %v", stageTitle, err)
		return err
	}

	// Capture the response
	if err = saveResponse(resp, body, action.Response.Capture, variables); err != nil {
		log.Warnf("%s ---> Error while capturing the response: %v", stageTitle, err)
		return err
	}

	log.Infof("%s ---> OK", stageTitle)
	return nil
}

func prepareQuery(variables map[string]interface{}, query definition.Query) (*http.Request, error) {
	// Prepare the bodyToSend the bodyToSend
	bodyToSend := []byte("")
	if len(query.BodyText) > 0 {
		//
		// If the bodyToSend is full text
		//
		stringToSend, err := formatString(query.BodyText, variables)
		if err != nil {
			return nil, fmt.Errorf("the definition of the text body is not usable due to %v", err)
		}

		bodyToSend = []byte(stringToSend)

	} else if len(query.BodyJSON) > 0 {

		jsonToSend, err := prepareJSONBody(variables, query.BodyJSON)
		if err != nil {
			return nil, fmt.Errorf("the definition of the json body is not usable due to %v", err)
		}

		jsonBody, err := json.Marshal(jsonToSend)
		if err != nil {
			return nil, fmt.Errorf("the definition of the json body is not convertible to json due to %v", err)
		}
		bodyToSend = jsonBody
	}

	url, err := formatString(query.URL, variables)
	if err != nil {
		return nil, fmt.Errorf("the definition of the URL is not usable due to %v", err)
	}

	// Make the base request
	req, err := http.NewRequest(
		query.Method.ToString(),
		url,
		bytes.NewBuffer(bodyToSend))
	if err != nil {
		return nil, err
	}

	// Add the headers of the query (if any)
	for k, v := range query.Headers {
		formatted, err := formatString(v, variables)
		if err != nil {
			return nil, fmt.Errorf("the definition of the headers is not usable due to %v", err)
		}

		req.Header.Add(k, formatted)
	}

	// Add the parameters of the query (if any)
	for k, v := range query.Params {

		formatted, err := formatString(v, variables)
		if err != nil {
			return nil, fmt.Errorf("the definition of the parameters is not usable due to %v", err)
		}

		req.URL.Query().Add(k, formatted)
	}

	return req, nil
}

func prepareJSONBody(variables map[string]interface{}, source interface{}) (interface{}, error) {

	// Depending on the type
	switch sourceType := source.(type) {

	case int:
		return source, nil

	case float64:
		return source, nil

	case []interface{}:

		result := make([]interface{}, len(sourceType))
		for i, obj := range sourceType {
			converted, err := prepareJSONBody(variables, obj)
			if err != nil {
				return nil, err
			}
			result[i] = converted
		}

		return result, nil

	case map[string]interface{}:

		result := make(map[string]interface{}, len(sourceType))
		for key, val := range sourceType {
			converted, err := prepareJSONBody(variables, val)
			if err != nil {
				return nil, err
			}

			result[key] = converted
		}

		return result, nil

	case string:

		// If the string does not have marker for template, use it directly
		if !strings.Contains(sourceType, "{{") {
			return sourceType, nil
		}

		// If the string is just a single marker, do it quickly
		r := regexp.MustCompile(`^\s*{{\s*\.(\w+)\s*}}\s*$`)
		matches := r.FindStringSubmatch(sourceType)
		if len(matches) == 2 {
			return variables[matches[1]], nil
		}

		// Otherwise something more complicated is needed
		return formatString(sourceType, variables)

	default:
		return nil, fmt.Errorf("the definition of the body is using not supported data format (%v)", reflect.TypeOf(sourceType).String())
	}
}

func checkResponse(resp *http.Response, body []byte, check definition.Validation) error {

	// If status code are defined, check if they are correct
	if len(check.StatusCodes) > 0 {

		codeReceivedAccepted := false
		for i := 0; i < len(check.StatusCodes) && !codeReceivedAccepted; i++ {
			if uint(resp.StatusCode) == check.StatusCodes[i] {
				codeReceivedAccepted = true
			}
		}

		if !codeReceivedAccepted {
			return fmt.Errorf("received Bad Status %d (Expected: %v)", resp.StatusCode, check.StatusCodes)
		}
	}

	// Check Headers
	if len(check.Headers) > 0 {

		for headerName, expectedValue := range check.Headers {

			responseHeader := resp.Header.Get(headerName)
			if len(responseHeader) == 0 || responseHeader != expectedValue {
				return fmt.Errorf("expected '%s' for response header '%s', but was not received", expectedValue, headerName)
			}
		}
	}

	if len(check.BodyJSON) > 0 || len(check.BodyText) > 0 {

		if len(body) == 0 {
			return fmt.Errorf("body should be checked, but can not be read from response")
		}

		// Check the body against a Regex
		if len(check.BodyText) > 0 {

			matched, err := regexp.Match(check.BodyText, body)

			if err != nil {
				return fmt.Errorf("body text should be checked against the RegExp '%v' but the RegExp is probably malformed", check.BodyText)
			}

			if !matched {
				return fmt.Errorf("body of the query is not matching the expected RegExp")
			}
		}

		if len(check.BodyJSON) > 0 {
			var data interface{}
			if err := json.Unmarshal(body, &data); err != nil {
				return fmt.Errorf("body text should be checked against JSON, but the body can not be converted to JSON")
			}

			// Check the body against a Regex
			for jsonKey, jsonValue := range check.BodyJSON {
				if err := checkJSONValue(data, jsonKey, jsonValue); err != nil {
					return fmt.Errorf("an expected value is the JSON response can not be found : %v", err)
				}
			}
		}
	}

	return nil
}

func saveResponse(resp *http.Response, body []byte, save definition.Capture, variables map[string]interface{}) error {

	// Capture the header
	for headerName, variableName := range save.Headers {
		headerValue := resp.Header.Get(headerName)
		variables[variableName] = headerValue
	}

	// Capture the body as a string
	if len(save.BodyText) > 0 {
		variables[save.BodyText] = string(body)
	}

	// Capture the json part
	if len(save.BodyJSON) > 0 {

		var jsonBody interface{}

		if err := json.Unmarshal(body, &jsonBody); err != nil {
			return fmt.Errorf("expected a response with a JSON body, but was not readable due to %v", err)
		}

		for jsonKey, variableName := range save.BodyJSON {

			jsonValue, err := getJSONValue(jsonBody, jsonKey)
			if err != nil {
				return fmt.Errorf("expected a response with a JSON having a value for the key '%s'", jsonKey)
			}
			variables[variableName] = jsonValue
		}
	}

	return nil
}
