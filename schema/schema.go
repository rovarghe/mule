package schema

import (
	"context"
	"net/http"

	"github.com/rovarghe/mule/plugin"
)

type (
	ContextHandler func(context.Context, *http.Request) (context.Context, error)
	Renderer       func(context.Context, *http.Request, http.ResponseWriter) (context.Context, error)

	PathSpec string

	ServeFunc func(ctx context.Context, request *http.Request, parent ContextHandler) (interface{}, error)

	RenderFunc func(ctx context.Context, request *http.Request, response http.ResponseWriter, parent Renderer) (interface{}, error)

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

	processResultKeyType string
	renderResultKeyType  string
)

const (
	RootModuleID     = plugin.ID("bootstrap")
	ProcessResultKey = processResultKeyType("result")
	RenderResultKey  = renderResultKeyType("rendered")
	// ModuleContextKey = moduleContextKeyType("moduleContextKey")
)
