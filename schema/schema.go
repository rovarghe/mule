package schema

import (
	"context"
	"net/http"

	"github.com/rovarghe/mule/plugin"
)

type (
	State interface{}

	ContextHandler func(State, *http.Request) (State, error)
	Renderer       func(State, *http.Request, http.ResponseWriter) (State, error)

	PathSpec string

	ServeFunc func(state State, request *http.Request, parent ContextHandler) (State, error)

	RenderFunc func(state State, request *http.Request, response http.ResponseWriter, parent Renderer) (State, error)

	PathHandlers map[PathSpec]ServeFunc

	Router interface {
		AddRoute(PathSpec, ServeFunc, RenderFunc)
	}

	Routers interface {
		Default() Router
		Get(PathSpec) Router
	}

	BaseRouters interface {
		Get(id plugin.ID) Routers
	}

	StartupFunc  func(context.Context, BaseRouters) (context.Context, error)
	ShutdownFunc func(context.Context) (context.Context, error)

	Module struct {
		plugin.Plugin
		Startup  StartupFunc
		Shutdown ShutdownFunc
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
