package schema

import (
	"context"
	"net/http"

	"github.com/rovarghe/mule/plugin"
)

type (
	NextHandler func(context.Context, http.ResponseWriter, *http.Request)

	PathSpec string

	ServeFunc func(context.Context, http.ResponseWriter, *http.Request, NextHandler) context.Context

	PathHandlers map[PathSpec]ServeFunc

	Router interface {
		AddRoute(PathSpec, ServeFunc)
	}

	Routers interface {
		Default() Router
		Get(PathSpec) Router
	}

	BaseRouters map[plugin.ID]Routers

	StartupFunc  func(context.Context, BaseRouters) (context.Context, error)
	ShutdownFunc func(context.Context) (context.Context, error)

	Module struct {
		plugin.Plugin
		Startup  StartupFunc
		Shutdown ShutdownFunc
	}
)

const (
	RootModuleID = plugin.ID("ROOT")
)
