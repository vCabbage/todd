/*
    ToDD TestRun definition

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package objects

import (
	"encoding/json"
	"fmt"
)

// TestRunObject is a specific implementation of BaseObject. It represents the "testrun" object in ToDD, which
// is a set of parameters under which a test should be run
type TestRunObject struct {
	BaseObject `yaml:",inline"` // the ",inline" tag is necessary for the go-yaml package to properly see the outer struct fields
	Spec       struct {
		TargetType string            `json:"targettype" yaml:"targettype"`
		Source     map[string]string `json:"source" yaml:"source"`
		Target     Target            `json:"target" yaml:"target"` // targettype of "group" uses this as a map, targettype of "uncontrolled" uses this as a slice.
		//App        string            `json:"app" yaml:"app"`  //TODO(mierdin): temporarily commenting out because App is defined in Source and Target now.
	} `json:"spec" yaml:"spec"`
}

// GetSpec is a simple function to return the "Spec" attribute of a TestRunObject
func (t TestRunObject) GetSpec() string {
	return fmt.Sprint(t.Spec)
}

// Target unmarshals YAML to either a map or a slice.
type Target struct {
	Map   map[string]string
	Slice []string
}

// UnmarshalYAML fulfills the yaml.Unmarshaler interface.
func (t *Target) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if unmarshal(&t.Map) == nil {
		return nil
	}
	return unmarshal(&t.Slice)
}

// MarshalJSON fulfills the json.Marshaller interface.
func (t Target) MarshalJSON() ([]byte, error) {
	if t.Map != nil {
		return json.Marshal(t.Map)
	}
	return json.Marshal(t.Slice)
}

// UnmarshalJSON fulfills the json.Unmarshaler interface.
func (t *Target) UnmarshalJSON(b []byte) error {
	if json.Unmarshal(b, &t.Map) == nil {
		return nil
	}
	return json.Unmarshal(b, &t.Slice)
}

type SourceOverrides struct {
	Group string `json:"source_group"`
	App   string `json:"source_app"`
	Args  string `json:"source_args"`
}

// AnySet returns whether and of the fields are set.
func (s *SourceOverrides) AnySet() bool {
	return s.Group != "" || s.App != "" || s.Args != ""
}

// Apply overrides source params as necessary
func (s *SourceOverrides) Apply(source map[string]string) {
	if s.App != "" {
		source["app"] = s.App
	}
	if s.Args != "" {
		source["args"] = s.Args
	}
	if s.Group != "" {
		source["name"] = s.Group
	}
}
