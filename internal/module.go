package internal

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/rovarghe/mule/loader"
	"github.com/rovarghe/mule/plugin"
	"github.com/rovarghe/mule/schema"
)

type (
	routesCtxKeyType string
	routersImpl      map[plugin.ID]pathSpecRoutersList

	pluginServeFunc struct {
		id            plugin.ID
		stateReducer  schema.StateReducer
		renderReducer schema.RenderReducer
	}

	pluginServeFuncList []pluginServeFunc

	pathSpecRoutersList struct {
		defaultPathSpec         schema.PathSpec
		pathSpecServFuncListMap map[schema.PathSpec]pluginServeFuncList
	}

	moduleLoadingContext struct {
		allRouters *routersImpl
	}

	pluginLoadingContext struct {
		moduleLoadingContext
		loadedPlugin *loader.LoadedPlugin
	}

	parentLoadingContext struct {
		pluginLoadingContext
		parentId plugin.ID
	}

	pathSpecLoadingContext struct {
		parentLoadingContext
		pathSpec schema.PathSpec
	}
)

/*
func (r *routersImpl) String() string {

	str := "{"
	sep := ""
	for k, v := range *r {
		str = fmt.Sprintf("%s%s%s:%v", str, sep, k, v)
		sep = ","
	}
	str = str + "}"
	return str
}
*/

func (psr parentLoadingContext) Default() schema.Router {
	all := *psr.allRouters
	return psr.Get(all[psr.parentId].defaultPathSpec)

}

func (psr parentLoadingContext) Get(ps schema.PathSpec) schema.Router {
	return pathSpecLoadingContext{parentLoadingContext: psr, pathSpec: ps}
}

func (psr pathSpecLoadingContext) AddRoute(ps schema.PathSpec, sf schema.StateReducer, rf schema.RenderReducer) {
	currentPluginId := psr.loadedPlugin.Plugin().ID()
	all := *psr.allRouters
	psrl := all[psr.parentId]

	if len(psrl.pathSpecServFuncListMap) == 0 {
		psrl.defaultPathSpec = ps
	}

	psf := pluginServeFunc{
		id:            currentPluginId,
		stateReducer:  sf,
		renderReducer: rf,
	}
	if psrl.pathSpecServFuncListMap == nil {
		psrl.pathSpecServFuncListMap = map[schema.PathSpec]pluginServeFuncList{}
	}
	if len(psrl.pathSpecServFuncListMap[ps]) == 0 {
		psrl.pathSpecServFuncListMap[ps] = pluginServeFuncList{psf}
	} else {
		psrl.pathSpecServFuncListMap[ps] = append(psrl.pathSpecServFuncListMap[ps], psf)
	}

	all[psr.parentId] = psrl
}

func (pr pluginLoadingContext) Get(id plugin.ID) schema.Routers {

	// There is an implicit dependency on the RootModuleID/"bootstrap"
	// All others need to be explicit.
	if id != schema.RootModuleID {
		check := false
		for _, d := range pr.loadedPlugin.Plugin().Dependencies() {
			if d.ID == id {
				check = true
				break
			}
		}

		if !check {
			// This is a dependency specification issue
			// To Get routers from a Module, there needs to be a dependency to that module
			panic(fmt.Sprintf("Invalid access, module '%s' is not a dependency of '%s'. Contact module provider.", string(id), pr.loadedPlugin.Plugin().ID()))
		}
	}

	return parentLoadingContext{pluginLoadingContext: pr, parentId: id}
}

type notFoundType struct{}

func notFoundServeFunc(state schema.State, ctx schema.StateReducerContext, r *http.Request, p schema.DefaultStateReducer) (schema.State, error) {
	return notFoundType{}, nil
}

func defaultRenderer(state schema.State, ctx schema.RenderReducerContext, r *http.Request, w http.ResponseWriter, parent schema.DefaultRenderReducer) (schema.State, error) {

	if _, ok := state.(notFoundType); ok {
		http.NotFound(w, r)
	} else {
		panic("Cannot handle state of type unknown")
	}

	return nil, nil
}

var (
	bootstrapModule = schema.Module{
		Plugin:   plugin.NewPlugin("bootstrap", plugin.Version{1, 0, 0, ""}, []plugin.Dependency{}),
		Startup:  nil,
		Shutdown: nil,
	}

	emptyPathSpec = schema.PathSpec("")

	//modules      = []schema.Module{bootstrapModule}
	moduleCtxKey = routesCtxKeyType("moduleContext")
)

func newModuleLoadingContext() moduleLoadingContext {
	return moduleLoadingContext{
		allRouters: &routersImpl{
			bootstrapModule.ID(): pathSpecRoutersList{
				defaultPathSpec: emptyPathSpec,
				pathSpecServFuncListMap: map[schema.PathSpec]pluginServeFuncList{
					emptyPathSpec: pluginServeFuncList{
						pluginServeFunc{
							id:            bootstrapModule.ID(),
							stateReducer:  notFoundServeFunc,
							renderReducer: defaultRenderer,
						},
					},
				},
			},
		},
	}
}

func LoadModules(ctx context.Context, modules []schema.Module) (context.Context, error) {

	var plugins = make([]plugin.Plugin, len(modules))

	for i := 0; i < len(modules); i++ {
		plugins[i] = modules[i]
	}
	ctx = context.WithValue(ctx, moduleCtxKey, newModuleLoadingContext())

	ctx, loadedPlugins, err := loader.Load(ctx, plugins, startModule)

	if err != nil {
		log.Println("Load incomplete,", loadedPlugins.Count(), "modules loaded")
	}

	return ctx, err

}

func startModule(ctx context.Context, lp *loader.LoadedPlugin) (context.Context, error) {

	mCtx := ctx.Value(moduleCtxKey).(moduleLoadingContext)

	plugin := lp.Plugin()
	module := plugin.(schema.Module)

	if module.ID() == bootstrapModule.ID() {
		log.Println("Bootstrapped")
		return ctx, nil
	}

	if module.Startup == nil {
		log.Printf("Starting module: (%s, %s)", plugin.ID(), plugin.Version())
		return ctx, nil
	}

	log.Printf("Starting module: %s %s", plugin.ID(), plugin.Version())

	mLoadingCtx := pluginLoadingContext{
		moduleLoadingContext: mCtx,
		loadedPlugin:         lp,
	}

	return module.Startup(ctx, mLoadingCtx)
}
