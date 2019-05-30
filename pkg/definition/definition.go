package definition

import (
	"errors"

	"gopkg.in/yaml.v3"
)

// Method defines an HTTP method supported by the queries
type Method int

const (
	// GET is for HTTP method GET
	GET Method = iota
	// PUT is for HTTP method PUT
	PUT
	// POST is for HTTP method POST
	POST
	// DELETE is for HTTP method DELETE
	DELETE
	// PATCH is for HTTP method PATCH
	PATCH
	// OPTIONS is for HTTP method OPTIONS
	OPTIONS
	// HEAD is for HTTP method HEAD
	HEAD
)

// Test is the structure of a test. The test is the higher level object. A Test is compose of various Stages.
type Test struct {
	TestName               string  `yaml:"test_name"`
	ContinueOnStageFailure bool    `yaml:"continue_on_stage_failure"`
	Stages                 []Stage `yaml:"stages"`
	Swarm                  Swarm   `yaml:"swarm,omitempty"`
}

// Swarm is the structure defining the startup options
type Swarm struct {
	NumberOfRuns uint `yaml:"number_of_runs"`
	CreationRate uint `yaml:"creation_rate"`
}

// Stage is a logical partition inside a Test. A Stage is composed of various Actions.
type Stage struct {
	Name           string   `yaml:"stage_name"`
	MaximumRetries uint     `yaml:"max_retries,omitempty"`
	DelayBefore    uint     `yaml:"delay_before,omitempty"`
	DelayAfter     uint     `yaml:"delay_after,omitempty"`
	Actions        []Action `yaml:"actions"`
}

// Action is a single query/response to a REST service. Each Action is composed of one Query and One Response.
type Action struct {
	Name     string   `yaml:"action_name"`
	Query    Query    `yaml:"query"`
	Response Response `yaml:"response"`
}

// Query is the query executed against a REST service.
type Query struct {
	URL      string                 `yaml:"url"`
	Method   Method                 `yaml:"method"`
	Headers  map[string]string      `yaml:"headers,omitempty"`
	Params   map[string]string      `yaml:"params,omitempty"`
	BodyJSON map[string]interface{} `yaml:"body_json,omitempty"`
	BodyText string                 `yaml:"body_text,omitempty"`
	Timeout  uint                   `yaml:"timeout,omitempty"`
}

// Response is the operations executed for a response to a Query. A Response can have a Validation and / or a Capture.
type Response struct {
	Validation Validation `yaml:"validation,omitempty"`
	Capture    Capture    `yaml:"capture,omitempty"`
}

// Validation defines the validations that are run against a Response.
type Validation struct {
	StatusCodes []uint                 `yaml:"status_codes,omitempty"`
	Headers     map[string]string      `yaml:"headers,omitempty"`
	BodyJSON    map[string]interface{} `yaml:"body_json,omitempty"`
	BodyText    string                 `yaml:"body_text,omitempty"`
}

// Capture defines the capture that are done with the Response.
type Capture struct {
	Headers  map[string]string `yaml:"headers,omitempty"`
	BodyJSON map[string]string `yaml:"body_json,omitempty"`
	BodyText string            `yaml:"body_text,omitempty"`
}

var toString = map[Method]string{
	GET:     "GET",
	PUT:     "PUT",
	POST:    "POST",
	DELETE:  "DELETE",
	PATCH:   "PATCH",
	OPTIONS: "OPTIONS",
	HEAD:    "HEAD",
}

var toID = map[string]Method{
	"GET":     GET,
	"PUT":     PUT,
	"POST":    POST,
	"DELETE":  DELETE,
	"PATCH":   PATCH,
	"OPTIONS": OPTIONS,
	"HEAD":    HEAD,
}

// ToString returns the string representation of a Method
func (method Method) ToString() string {
	return toString[method]
}

// MarshalYAML marshals the enum as a quoted json string
func (method Method) MarshalYAML() (interface{}, error) {
	return toString[method], nil
}

// UnmarshalYAML unmarshals a quoted json string to the enum value
func (method *Method) UnmarshalYAML(node *yaml.Node) error {

	if node.Kind != yaml.ScalarNode {
		return errors.New("the method field is expected to be a string")
	}

	m, ok := toID[node.Value]
	if !ok {
		return errors.New("the method field is expected to be an HTTP method: GET, POST, PUT, DELETE, PATCH, OPTIONS or HEAD")
	}

	// Copy the value
	*method = m

	return nil
}
