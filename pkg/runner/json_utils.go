package runner

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// checkJSONValue checks that expected is present in the given JSON tree and that the value in the JSON tree matches
// the expected one. If the value is absent or does not match the expected one, an error is returned
//
// Params:
//  - json: a JSON tree
//  - key: the key of the value, dot separated. For example "user.name.first"
//  - expected: the expected value. Note that expected string values are considered as regexp
//
// Return nil, if the value is present in the json and valid, an error otherwise
func checkJSONValue(json interface{}, key string, expected interface{}) error {

	// Split the key in parts
	rawValue, err := getJSONValue(json, key)

	if err != nil {
		return fmt.Errorf("unable to check the retrieve the key '%s' from JSON body", key)
	}

	// Otherwise test the value
	switch readValueType := rawValue.(type) {

	// Checks when an int was read
	case int:
		switch expectedValueType := expected.(type) {
		case int:
			if readValueType == expectedValueType {
				return fmt.Errorf("unable to check the JSON attribute '%s', expected '%v'(int) and got '%v'(int)", key, expectedValueType, readValueType)
			}
		case float64:
			if float64(readValueType) == expectedValueType {
				return fmt.Errorf("unable to check the JSON attribute '%s', expected '%v'(float64) and got '%v'(int)", key, expectedValueType, readValueType)
			}

		case string:
			readValueString := []byte(strconv.Itoa(readValueType))

			matched, err := regexp.Match(expectedValueType, readValueString)

			if err != nil {
				return fmt.Errorf("value should be checked against the RegExp '%v' but the RegExp is probably malformed", expectedValueType)
			}

			if !matched {
				return fmt.Errorf("unable to check the JSON attribute '%s', expected '%v'(string) and got '%v'(int)", key, expectedValueType, readValueType)
			}

		default:
			return fmt.Errorf("unable to check the JSON attribute '%s', expected '%v'(%s) and got '%v'(int)", key, expectedValueType, reflect.TypeOf(expectedValueType).String(), readValueType)
		}

	// Checks when a float64 was read
	case float64:

		switch expectedValueType := expected.(type) {
		case int:
			if readValueType == float64(expectedValueType) {
				return fmt.Errorf("unable to check the JSON attribute '%s', expected '%v'(int) and got '%v'(float64)", key, expectedValueType, readValueType)
			}
		case float64:
			if readValueType == expectedValueType {
				return fmt.Errorf("unable to check the JSON attribute '%s', expected '%v'(float64) and got '%v'(float64)", key, expectedValueType, readValueType)
			}

		case string:
			readValueString := []byte(fmt.Sprintf("%f", readValueType))

			matched, err := regexp.Match(expectedValueType, readValueString)

			if err != nil {
				return fmt.Errorf("value should be checked against the RegExp '%v' but the RegExp is probably malformed", expectedValueType)
			}

			if !matched {
				return fmt.Errorf("unable to check the JSON attribute '%s', expected '%v'(string) and got '%v'(float64)", key, expectedValueType, readValueType)
			}

		default:
			return fmt.Errorf("unable to check the JSON attribute '%s', expected '%v'(%s) and got '%v'(float64)", key, expectedValueType, reflect.TypeOf(expectedValueType).String(), readValueType)
		}

	// Checks when a string was read
	case string:
		switch expectedValueType := expected.(type) {
		case string:

			matched, err := regexp.Match(expectedValueType, []byte(readValueType))

			if err != nil {
				return fmt.Errorf("value should be checked against the RegExp '%v' but the RegExp is probably malformed", expectedValueType)
			}

			if !matched {
				return fmt.Errorf("unable to check the JSON attribute '%s', expected '%v'(string) and got '%v'(string)", key, expectedValueType, readValueType)
			}

		default:
			return fmt.Errorf("unable to check the JSON attribute '%s', expected '%v'(%s) and got '%v'(string)", key, expectedValueType, reflect.TypeOf(expectedValueType).String(), readValueType)
		}

	// Checks when other things were read
	default:
		return fmt.Errorf("unable to check the JSON attribute '%s', expected '%v'(%s) and got '%v'(%s)", key, expected, reflect.TypeOf(expected).String(), readValueType, reflect.TypeOf(rawValue).String())
	}

	// Wow, no failure
	return nil
}

// getJSONValue returns a specific value in a JSON tree. If the value is absent,  an error is returned. Note that
// the returned value, may itself be a node, with sub-nodes, etc.
//
// Params:
//  - json: a JSON tree
//  - key: the key of the value, dot separated. For example "user.name.first"
//
// Return the value if found, an error otherwise
func getJSONValue(json interface{}, key string) (interface{}, error) {

	// Split the key in parts
	path := strings.Split(key, ".")

	return getJSONPath(json, path)
}

// getJSONPath returns a specific value in a JSON tree. If the value is absent,  an error is returned. Note that
// the returned value, may itself be a node, with sub-nodes, etc.
//
// Params:
//  - json: a JSON tree
//  - path: the components of the path in an array. For example ["user", "name", "first"]
//
// Return the value if found, an error otherwise
func getJSONPath(json interface{}, path []string) (interface{}, error) {

	// The node to test must be a map
	targetTree, ok := json.(map[string]interface{})
	if !ok {
		return nil, errors.New("response's JSON does not match requested format")
	}

	targetNode, ok := targetTree[path[0]]
	if !ok {
		return nil, errors.New("response's JSON does not match requested path")
	}

	// Go down in the path if needed
	if len(path) > 1 {

		return getJSONPath(targetNode, path[1:])
	}

	return targetNode, nil
}
