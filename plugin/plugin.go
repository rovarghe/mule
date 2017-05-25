package plugin

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"strings"
)

type Provider struct {
	Name    string
	Website string
	Contact string
}

type Dependency struct {
	ID    PluginID
	Range Range
}

type Renderer interface {
	Render(ctx *context.Context, w http.ResponseWriter) error
}

type Forwarder interface {
	Forward(ctx *context.Context, r *http.Request)
}

type HandlerFunc func(ctx *context.Context, r *http.Request, next Forwarder) (*context.Context, interface{})

type PluginID string

type Plugin struct {
	ID           PluginID
	Provider     Provider
	Version      Version
	Dependencies []Dependency
	WebContext   string
	HandlerFunc  HandlerFunc
}

func (p *Plugin) Equals(other *Plugin) bool {
	return reflect.DeepEqual(*p, *other)
}

func (p *Plugin) Satisfies(d *Dependency) bool {
	return p.ID == d.ID && p.Version.IsWithin(&d.Range)
}

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
		ID:    PluginID(id),
		Range: *r,
	}, nil

}
