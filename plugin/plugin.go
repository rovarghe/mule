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

// Provider is the source of the plugin
type Provider struct {
	Name    string
	URL     string
	Contact string
}

// Dependency is a link from one plugin to another
type Dependency struct {
	ID    ID
	Range Range
}

func (d *Dependency) String() string {
	return fmt.Sprintf("%s %s", d.ID, d.Range.String())
}

// ID is a unique ID for the plugin.
// There can be different versions for the same ID
type ID string

// Plugin describes a single loadable unit of functionality
// A Plugin has a version and optionally may have dependencies to one or more other Plugins
type Plugin struct {
	ID           ID
	Provider     Provider
	Version      Version
	Dependencies []Dependency
	Payload      interface{}
}

// Equals checks if one plugin is exactly equal to another
func (p *Plugin) Equals(other *Plugin) bool {
	return reflect.DeepEqual(*p, *other)
}

// Satisfies returns true if a Plugin has the same ID and falls within the Range
// of a Dependency
func (p *Plugin) Satisfies(d *Dependency) bool {
	return p.ID == d.ID && p.Version.IsWithin(&d.Range)
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
