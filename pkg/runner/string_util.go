package runner

import (
	"bytes"
	"strings"
	"text/template"
)

// formatString formats the given string, filling its placeholder with the values
// coming from `variables`
func formatString(str string, variables map[string]interface{}) (string, error) {

	// If the string does not have marker for template, use it directly
	if !strings.Contains(str, "{{") {
		return str, nil
	}

	// Create the template
	tmpl, err := template.New("temp").Parse(str)
	if err != nil {
		return "", err
	}

	// Execute the template
	var tpl bytes.Buffer
	err = tmpl.Execute(&tpl, variables)
	if err != nil {
		return "", err
	}

	// Return the templated output
	return tpl.String(), nil
}
