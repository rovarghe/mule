package builtin

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/rovarghe/mule/plugin"
	"github.com/rovarghe/mule/schema"
)

var (
	version1 = plugin.Version{1, 0, 0, ""}

	CoreModule = schema.Module{
		Plugin:   plugin.NewPlugin(plugin.ID("mule"), version1, []plugin.Dependency{}),
		Startup:  coreStartupFunc,
		Shutdown: coreShutdownFunc,
	}

	AboutModule = schema.Module{
		Plugin: plugin.NewPlugin(plugin.ID("about"), version1, []plugin.Dependency{
			plugin.Dependency{
				ID:    CoreModule.Plugin.ID(),
				Range: plugin.Range{version1, version1, true, true},
			},
		}),
		Startup:  aboutStartupFunc,
		Shutdown: nil,
	}
)

func coreHandler(ctx context.Context, r *http.Request, parent schema.ContextHandler) (interface{}, error) {

	ctx, err := parent(ctx, r)
	if err != nil {
		return ctx, err
	}

	return nil, nil
}

func coreRenderer(ctx context.Context, r *http.Request, w http.ResponseWriter, parent schema.Renderer) (interface{}, error) {
	intf := ctx.Value(schema.RenderResultKey)

	if intf != nil && r.Header.Get("Content-Type") == "application/json" {

		js, err := json.Marshal(intf)
		if err != nil {
			log.Fatal("Cannot convert ", intf, " to json")
			return []byte("Internal Server Error"), err

		}

		ctx = context.WithValue(ctx, schema.RenderResultKey, js)
		return parent(ctx, r, w)

	}

	return ctx, nil
}

func coreStartupFunc(ctx context.Context, base schema.BaseRouters) (context.Context, error) {
	routers := base.Get(schema.RootModuleID)
	routers.Default().AddRoute(schema.PathSpec(""), coreHandler, coreRenderer)
	return ctx, nil
}

func coreShutdownFunc(ctx context.Context) (context.Context, error) {

	return nil, nil
}

func aboutStartupFunc(ctx context.Context, base schema.BaseRouters) (context.Context, error) {
	routers := base.Get(CoreModule.ID())
	routers.Default().AddRoute(schema.PathSpec("about"), aboutHandler, aboutRenderer)
	return ctx, nil
}

func aboutHandler(ctx context.Context, r *http.Request, parent schema.ContextHandler) (interface{}, error) {

	return map[string]string{"msg": "About the world"}, nil
}

func aboutRenderer(ctx context.Context, r *http.Request, w http.ResponseWriter, parent schema.Renderer) (interface{}, error) {
	return ctx.Value(schema.ProcessResultKey), nil
}
