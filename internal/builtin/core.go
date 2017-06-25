package builtin

import (
	"context"
	"fmt"
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

func coreHandler(ctx context.Context, r *http.Request, parent schema.ContextHandler) (context.Context, error) {
	fmt.Println("Calling parent in core")
	ctx, err := parent(ctx, r)
	if err != nil {
		return ctx, err
	}
	fmt.Println("Hello World")
	return context.WithValue(ctx, "core", "Hello world"), nil
}

func coreStartupFunc(ctx context.Context, base schema.BaseRouters) (context.Context, error) {
	routers := base.Get(schema.RootModuleID)
	routers.Default().AddRoute(schema.PathSpec(""), coreHandler)
	return ctx, nil
}

func coreShutdownFunc(ctx context.Context) (context.Context, error) {
	return nil, nil
}

func aboutStartupFunc(ctx context.Context, base schema.BaseRouters) (context.Context, error) {
	routers := base.Get(CoreModule.ID())
	routers.Default().AddRoute(schema.PathSpec("about"), aboutHandler)
	return ctx, nil
}

func aboutHandler(ctx context.Context, r *http.Request, parent schema.ContextHandler) (context.Context, error) {
	fmt.Println("About")
	return context.WithValue(ctx, "about", "About the world"), nil
}
