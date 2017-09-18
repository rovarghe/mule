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

func BaseServeFunc(state schema.State, ctx schema.StateReducerContext, r *http.Request, parent schema.DefaultStateReducer) (schema.State, error) {
	return state, nil
}

func BaseRenderFunc(state schema.State, ctx schema.RenderReducerContext, r *http.Request, w http.ResponseWriter, parent schema.DefaultRenderReducer) (schema.State, error) {
	return state, nil
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

func MavenServeFunc(state schema.State, ctx schema.StateReducerContext, r *http.Request, parent schema.DefaultStateReducer) (schema.State, error) {
	return ctx, nil
}

func MavenRenderFunc(state schema.State, ctx schema.RenderReducerContext, r *http.Request, w http.ResponseWriter, parent schema.DefaultRenderReducer) (schema.State, error) {

	return ctx, nil
}
