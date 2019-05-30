# Introduction

Gargote is a very simple and straightforward REST API tester. In short, it takes a list of REST queries from a file and 
execute. Although this may seem limited, Gargote offers some nice features:
 
 * Simple configuration file
 * Validation of the result of a query using RegExp
 * Capture of the result of a query
 * Usage of captured values to inject in next query
 * Swarm mode (load testing)
 
The capture/inject feature allows to build easily complex scenarios, for example getting an authentication token in the
body of a response and injecting it in the header of the next query.

# Usage
```bash
./gargote [configuration file]
```

# History and status

Currently, a some features and options are still missing, and some bugs are probably remaining. However, Gargote is 
usable and the information describe in the documentation "should" just work.

 * v0.0.2: 
   * Add the swarm mode
   * Add option continue on stage failure
 * v0.0.1: Initial version
 
Main coming features:

 * Keeping track of the result of each query (time, success, etc)
 * Exporting results and statistics
 * Some kind of UI ?

# Configuration file
The configuration file is a yaml file. Apart from being more readable than JSON, it also allows to comment the tests.

The structure is the following:
 * Test: the head object. Each test can be composed of various stages.
 * Stage: a logical division of the tests. Each stage can be composed of various actions.
 * Action: a REST request and its response. The response can have validation and capture information. 
 
## The test
 
| Attribute name | Type | Description |
| --- | --- | --- |
| test_name | string | The name of the test _(Mandatory)_|
| continue_on_stage_failure | bool | if true, in case of a stage failing, the test will follow up at the next stage (Default: false) |
| stages | List of Stage | The stages |
| swarm | An object Swarm | The configuration of the swarm |

### The swarm

| Attribute name | Type | Description |
| --- | --- | --- |
| number_of_runs | uint | The number of times the test will be run (Default: 1) |
| creation_rate | uint | The number of times that a new test is started by second (Default: 1)|

The swarm parameter allows to execute multiple times the same test, generating load on the server.

## The stage

| Attribute name | Type | Description |
| --- | --- | --- |
| stage_name | string | The name of the stage _(Mandatory)_ |
| max_retries | uint | The number of times the stage is re-tried if the execution fails (Default: 0) |
| delay_before | uint | A delay (in milliseconds) to wait before starting (Default: 0)  |
| delay_after | uint | A delay (in milliseconds) to wait after (Default: 0)  |
| actions | List of Action | The actions |

During its execution, the stage manage a list of variables. The variables are dynamically created by the capture and 
can then be injected in the queries. The variable do not need to be typed and can even be full object if needed.

## The action

| Attribute name | Type | Description |
| --- | --- | --- |
| action_name | string | The name of the action _(Mandatory)_ |
| query | An object Query | The REST query to be executed |
| response | An object Response | The response to the query to be checked |

### The query

| Attribute name | Type | Description |
| --- | --- | --- |
| url | string | The URL _(Mandatory)_ |
| method | string | The REST method: GET, PUT, POST, DELETE, PATCH, OPTIONS, HEAD _(Mandatory)_|
| headers | map[string]string | A list of values to set in the header |
| params | map[string]string | A list of params to add to the query |
| timeout | uint | The timeout of the query (default 1 minute) |
| body_text | string | The plain text body of the query |
| body_json | map[string]interface{} | A object to be used as the body once converted to JSON |

All the parameters can be injected with stage variables. The injection is simply using Go templates. The simplest case,
for directly injecting a variable is to use the syntax `{{ .VariableName }}`. For example:

```yaml
query:
  url: https://jsonplaceholder.typicode.com/users/{{ .the_user_id }}
```

The injection may happen in all string based attributes and even in the body_json attribute. Also, for the body_json, 
the injection respect the data type of the variable read. For example, if a previous query read a variable `the_user_id`
as an integer:

```yaml
query:
  body_json: 
    user_name: user_{{ .the_user_id }}
    user_id: {{ .the_user_id }}
```

will generate:

```json
{
  "user_name": "user_1",
  "user_id": 1
}
```

### The response

| Attribute name | Type | Description |
| --- | --- | --- |
| validation | An object Validation | The validation rules |
| capture | An object Capture | The validation rules |

Both validation and capture are optional.

#### The validation

| Attribute name | Type | Description |
| --- | --- | --- |
| status_codes | List of uint | The accepted HTTP code. For example: 200, 201, etc. |
| headers | map[string]string | A list of values to be defined in the header |
| body_text | string | The body of the response as plain text|
| body_json | map[string]interface{} | A list of path in the JSON and the value they must have |

For the values defined as string, the string is evaluated as a RegExp. For example, for checking that the response have
a keep-something header and HTTP status OK:

```yaml
response:
  check:
    status_code:
      - 200
    headers:
      Connection: keep-.*
```

The body_json is a bit specific. As generally, it is expected to check a single or few specific values inside a full 
JSON, the check is done the key of the validation as path and the value of the validation as the value expected in the 
JSON, limited to int, float, boolean and string. Strings are also evaluated as RegExp. For example, for validating the
following JSON: 

```json
{
  "user": {
    "name": "user_1",
    "id": 1
  },
  "company": {
    "name": "company_1",
    "id": 1,
    "address": {
      "city": "Limassol"
    }
  }
}
```

it is possible to use:

```yaml
response:
  validation:
    body_json: 
      "user.id": 1
      "company.address.city": [lL]imassol.*
```


#### The capture

| Attribute name | Type | Description |
| --- | --- | --- |
| headers | map[string]string | A list of pair of variable name / header name  |
| body_text | string | The name of the variable in which to capture the full body as text|
| body_json | map[string]string | A list of pair of variable name / JSON path  |

For example to capture _Connection_ header and the _userId_ of the JSON body:

```yaml
response:
  capture:
    headers:
      "Connection": connection_mode
    body_json: 
      "userId": id_of_connected_user
```

# Full example
The following example run two GET queries against the Typicode. The test is run just one time and will inject the data 
coming from the first query in the second.

```yaml
test_name: Test with the Typicode API
continue_on_stage_failure: false
swarm:
  number_of_runs: 1
  creation_rate: 1
stages:
  - stage_name: Basic API usage
    delay_before: 50
    delay_after: 100
    actions:
      - action_name: Get a todo 
        query:
          url: https://jsonplaceholder.typicode.com/todos/1
          method: GET
          headers:
            Accept: application/json
        response:
          validation:
            status_code:
              - 200
          capture:
            body_json:
              # Bind the field { "userId": "..." } to the variables the_user_id
              userId: the_user_id

      - action_name: Get the user of the previous todo
        query:
          # Use the variable the_user_id to patch URL
          url: https://jsonplaceholder.typicode.com/users/{{ .the_user_id }}
          method: GET
          headers:
            Accept: application/json
        response:
          validation:
            status_code:
              - 200
            body_json:
              # Check the the field { "company": {"name": "Romaguera-Crona"} } with a RegExp
              company.name: Roma.uera-.*
```

The _continue_on_stage_failure_ is not useful in this example as there is only a single stage.  

# License

Copyright 2019 Thomas Wuillemin  <thomas.wuillemin@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this project or its content except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
