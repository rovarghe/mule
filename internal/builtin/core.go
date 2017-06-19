package builtin

import (
	"context"
	"net/http"

	"github.com/rovarghe/mule/plugin"
	"github.com/rovarghe/mule/schema"
)

var (
	version0 = plugin.Version{0, 0, 0, ""}

	CoreModule = schema.Module{
		Plugin: plugin.NewPlugin(schema.CoreModuleID, plugin.Version{1, 0, 0, ""}, []plugin.Dependency{
			plugin.Dependency{
				ID:    plugin.ID("bootstrap"),
				Range: plugin.Range{version0, version0, true, true},
			},
		}),
		Startup:  coreStartupFunc,
		Shutdown: coreShutdownFunc,
	}
)

func coreHandler(ctx context.Context, r *http.Request, parent schema.ContextHandler, next schema.ContextHandler) context.Context {
	return nil
}

func coreStartupFunc(ctx context.Context, base schema.BaseRouters) (context.Context, error) {
	routers := base.Get("bootstrap")
	routers.Default().AddRoute(schema.PathSpec("/"), coreHandler)
	return ctx, nil
}

func coreShutdownFunc(ctx context.Context) (context.Context, error) {
	return nil, nil
}
