package schema

import (
	"context"
	"net/http"

	"github.com/rovarghe/mule/plugin"
)

type (
	State interface{}

	ReducerContext interface {
		URI() PathSpec
		PathParameters() map[string]string
		Final() bool
	}

	// RenderReducerContext interface {
	// 	URI() PathSpec
	// 	PathParameters() map[string]string
	// 	Final() bool
	// }

	DefaultStateReducer  func(State, *http.Request) (State, error)
	DefaultRenderReducer func(State, *http.Request, http.ResponseWriter) (State, error)

	PathSpec string

	StateReducer func(state State, context ReducerContext, request *http.Request, parent DefaultStateReducer) (State, error)

	RenderReducer func(state State, context ReducerContext, request *http.Request, response http.ResponseWriter, parent DefaultRenderReducer) (State, error)

	PathHandlers map[PathSpec]StateReducer

	Router interface {
		AddRoute(PathSpec, StateReducer, RenderReducer)
	}

	Routers interface {
		Default() Router
		Get(PathSpec) Router
	}

	BaseRouters interface {
		Get(id plugin.ID) Routers
	}

	Starter interface {
		Start(context.Context, BaseRouters) (context.Context, error)
	}

	Stopper interface {
		Stop(context.Context) (context.Context, error)
	}

	StarterFunc func(context.Context, BaseRouters) (context.Context, error)

	StopperFunc func(context.Context) (context.Context, error)

	Module struct {
		plugin.Plugin
		Starter Starter
		Stopper Stopper
	}

	// moduleContextKeyType string

	// ModuleContext interface {
	// 	Set(key interface{}, value interface{})
	// 	Value(key interface{}) interface{}
	// }

	processResultKeyType  string
	processContextKeyType string
	renderResultKeyType   string
)

const (
	RootModuleID      = plugin.ID("bootstrap")
	ProcessResultKey  = processResultKeyType("result")
	ProcessContextKey = processContextKeyType("processContext")
	//RenderResultKey  = renderResultKeyType("rendered")
	// ModuleContextKey = moduleContextKeyType("moduleContextKey")
)

func (f StarterFunc) Start(c context.Context, b BaseRouters) (context.Context, error) {
	return f(c, b)
}

func (f StopperFunc) Stop(c context.Context) (context.Context, error) {
	return f(c)
}
