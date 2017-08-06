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
	routers.Get(schema.RootModuleID).Default().AddRoute("/", BaseServeFunc, BaseRenderFunc)
	return ctx, nil

}

func BaseShutdownFunc(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func BaseServeFunc(ctx context.Context,
	r *http.Request, parent schema.ContextHandler) (interface{}, error) {
	return ctx, nil
}

func BaseRenderFunc(ctx context.Context, r *http.Request, w http.ResponseWriter, parent schema.Renderer) (interface{}, error) {
	return ctx, nil
}

var MavenModule = schema.Module{
	Plugin:   MavenPlugin,
	Startup:  MavenStartupFunc,
	Shutdown: MavenShutdownFunc,
}

func MavenStartupFunc(ctx context.Context, routers schema.BaseRouters) (context.Context, error) {
	fmt.Println("Maven startup executed, adding /maven")
	routers.Get(MavenPlugin.ID()).Default().AddRoute("/maven", MavenServeFunc, MavenRenderFunc)
	return ctx, nil

}

func MavenShutdownFunc(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func MavenServeFunc(ctx context.Context,
	r *http.Request, p schema.ContextHandler) (interface{}, error) {
	return ctx, nil
}

func MavenRenderFunc(ctx context.Context,
	r *http.Request,
	w http.ResponseWriter,
	parent schema.Renderer) (interface{}, error) {

	return ctx, nil
}
