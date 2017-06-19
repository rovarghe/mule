package harness

import (
	"context"
	"fmt"
	"net/http"

	"github.com/rovarghe/mule/schema"
)

var BaseModule = schema.Module{
	Plugin:   BasePlugin,
	Startup:  BaseStartupFunc,
	Shutdown: BaseShutdownFunc,
}

func BaseStartupFunc(ctx context.Context, routers schema.BaseRouters) (context.Context, error) {
	fmt.Println("Base startup executed, adding /")
	routers.Get(schema.CoreModuleID).Default().AddRoute("/", BaseServeFunc)
	return ctx, nil

}

func BaseShutdownFunc(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func BaseServeFunc(ctx context.Context, r *http.Request, parent schema.ContextHandler, next schema.ContextHandler) context.Context {
	return ctx
}

var MavenModule = schema.Module{
	Plugin:   MavenPlugin,
	Startup:  MavenStartupFunc,
	Shutdown: MavenShutdownFunc,
}

func MavenStartupFunc(ctx context.Context, routers schema.BaseRouters) (context.Context, error) {
	fmt.Println("Maven startup executed, adding /maven")
	routers.Get(MavenPlugin.ID()).Default().AddRoute("/maven", MavenServeFunc)
	return ctx, nil

}

func MavenShutdownFunc(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func MavenServeFunc(ctx context.Context, r *http.Request, p schema.ContextHandler, n schema.ContextHandler) context.Context {
	return ctx
}
