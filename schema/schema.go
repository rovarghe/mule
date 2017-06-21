package schema

import (
	"context"
	"net/http"

	"github.com/rovarghe/mule/plugin"
)

type (
	ContextHandler func(context.Context, *http.Request) (context.Context, error)

	PathSpec string

	ServeFunc func(ctx context.Context, request *http.Request, parent ContextHandler, next ContextHandler) (context.Context, error)

	PathHandlers map[PathSpec]ServeFunc

	Router interface {
		AddRoute(PathSpec, ServeFunc)
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
)

const (
	CoreModuleID = plugin.ID("mule")
)
