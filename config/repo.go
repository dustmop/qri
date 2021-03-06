package config

import "github.com/qri-io/jsonschema"

// Repo configures a qri repo
type Repo struct {
	Middleware []string `json:"middleware"`
	Type       string   `json:"type"`
}

// DefaultRepo creates & returns a new default repo configuration
func DefaultRepo() *Repo {
	return &Repo{
		Type:       "fs",
		Middleware: []string{},
	}
}

// Validate validates all fields of repo returning all errors found.
func (cfg Repo) Validate() error {
	schema := jsonschema.Must(`{
    "$schema": "http://json-schema.org/draft-06/schema#",
    "title": "Repo",
    "description": "Config for the qri repository",
    "type": "object",
    "required": ["middleware", "type"],
    "properties": {
      "middleware": {
        "description": "Middleware packages that need to be applied to the repo",
        "type": "array",
        "items": {
          "type": "string"
        }
      },
      "type": {
        "description": "Type of repository",
        "type": "string",
        "enum": [
          "fs"
        ]
      }
    }
  }`)
	return validate(schema, &cfg)
}
