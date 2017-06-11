/*
Package plugin provides a generic way to define Plugins.

A Plugin is a versioned unit of functionality that has dependencies to other Plugins.
The actual functionality is opaque, represented by the Payload attribute
*/
package plugin

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// Dependency is a link from one plugin to another
type Dependency struct {
	ID    ID
	Range Range
}

func (d Dependency) String() string {
	return fmt.Sprintf("%s %s", d.ID, d.Range.String())
}

// ID is a unique ID for the plugin.
// There can be different versions for the same ID
type ID string

func (id ID) Equals(other ID) bool {
	return id == other
}

// Plugin describes a single loadable unit of functionality
// A Plugin has a version and optionally may have dependencies to one or more other Plugins
type Plugin interface {
	ID() ID
	Version() Version
	Dependencies() []Dependency
	//Payload      interface{}
}

type DefaultPlugin struct {
	id           ID
	version      Version
	dependencies []Dependency
}

func (p DefaultPlugin) ID() ID {
	return p.id
}

func (p DefaultPlugin) Version() Version {
	return p.version
}

func (p DefaultPlugin) Dependencies() []Dependency {
	return p.dependencies
}

func NewPlugin(id ID, version Version, dependencies []Dependency) Plugin {
	return DefaultPlugin{id, version, dependencies}
}

// Satisfies returns true if a Plugin has the same ID and falls within the Range
// of a Dependency
func Satisfies(p Plugin, d Dependency) bool {
	return p.ID() == d.ID && p.Version().IsWithin(d.Range)
}

func PluginEquals(f Plugin, s Plugin) bool {
	return f.ID().Equals(s.ID()) &&
		s.Version().Equals(f.Version()) &&
		reflect.DeepEqual(f.Dependencies(), s.Dependencies())
}

// ParseDependency converts a string to a Dependency type
// Returns not-nil error if unable to parse.
func ParseDependency(s string) (*Dependency, error) {
	i := strings.IndexAny(s, " [(")
	if i < 0 {
		return nil, errors.New("No version range")
	}
	id := s[:i]
	s = s[i:]
	r, err := ParseRange(strings.TrimLeft(s, " "))
	if err != nil {
		return nil, err
	}
	return &Dependency{
		ID:    ID(id),
		Range: *r,
	}, nil

}
