package internal

import (
	"context"
	"fmt"

	"github.com/rovarghe/mule/loader"
	"github.com/rovarghe/mule/plugin"
	"github.com/rovarghe/mule/schema"
)

func registerModule(ctx context.Context, plugin *loader.LoadedPlugin) (context.Context, error) {
	return ctx, nil
}

func LoadModules(modules []schema.Module) {
	var plugins = make([]*plugin.Plugin, len(modules))

	for i := 0; i < len(modules); i++ {
		plugins[i] = &modules[i].Plugin
	}
	ctx, loadedPlugins, err := loader.Load(context.Background(), &plugins, registerModule)

	fmt.Println(ctx, err, loadedPlugins)
}
