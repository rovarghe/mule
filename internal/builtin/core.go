package builtin

import (
	"context"
	"encoding/json"
	"fmt"
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

func coreHandler(state schema.State, ctx schema.ReducerContext, r *http.Request, parent schema.DefaultStateReducer) (schema.State, error) {

	state, err := parent(state, r)

	return state, err
}

func coreRenderer(state schema.State, ctx schema.ReducerContext, r *http.Request, w http.ResponseWriter, parent schema.DefaultRenderReducer) (schema.State, error) {

	if state != nil {
		contentType := r.Header.Get("Accept")
		switch contentType {
		case "application/json":
			js, err := json.Marshal(state)
			if err != nil {
				log.Fatal("Cannot convert ", state, " to json")
				return []byte("Internal Server Error"), err
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(js)
		default:
			panic(fmt.Sprintf("Unable to handle content type %s", contentType))

		}
	}

	return state, nil
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

func aboutHandler(state schema.State, ctx schema.ReducerContext, r *http.Request, parent schema.DefaultStateReducer) (schema.State, error) {

	return map[string]string{"msg": "About the world"}, nil
}

func aboutRenderer(state schema.State, ctx schema.ReducerContext, r *http.Request, w http.ResponseWriter, parent schema.DefaultRenderReducer) (schema.State, error) {
	return state, nil
}
