package internal

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/rovarghe/mule/loader"
	"github.com/rovarghe/mule/plugin"
	"github.com/rovarghe/mule/schema"
)

type (
	routesCtxKeyType string
	routersImpl      map[plugin.ID]pathSpecRoutersList

	pluginServeFunc struct {
		id        plugin.ID
		serveFunc schema.ServeFunc
	}

	pluginServeFuncList []pluginServeFunc

	pathSpecRoutersList struct {
		defaultPathSpec         schema.PathSpec
		pathSpecServFuncListMap map[schema.PathSpec]pluginServeFuncList
	}

	moduleContext struct {
		allRouters *routersImpl
	}

	moduleLoadingContext struct {
		moduleContext
		loadedPlugin *loader.LoadedPlugin
	}

	parentRouters struct {
		moduleLoadingContext
		parentId plugin.ID
	}

	pathSpecRouters struct {
		parentRouters
		pathSpec schema.PathSpec
	}
)

func (psr parentRouters) Default() schema.Router {
	all := *psr.allRouters
	return psr.Get(all[psr.parentId].defaultPathSpec)

}

func (psr parentRouters) Get(ps schema.PathSpec) schema.Router {
	return pathSpecRouters{parentRouters: psr, pathSpec: ps}
}

func (psr pathSpecRouters) AddRoute(ps schema.PathSpec, sf schema.ServeFunc) {
	currentPluginId := psr.loadedPlugin.Plugin().ID()
	all := *psr.allRouters
	psrl := all[psr.parentId]
	if len(psrl.pathSpecServFuncListMap) == 0 {
		psrl.defaultPathSpec = ps
	}
	pssfl := psrl.pathSpecServFuncListMap[ps]

	psrl.pathSpecServFuncListMap[ps] = append(pssfl, pluginServeFunc{
		id:        currentPluginId,
		serveFunc: sf})
	all[psr.parentId] = psrl
}

func (pr moduleLoadingContext) Get(id plugin.ID) schema.Routers {

	check := false
	for _, d := range pr.loadedPlugin.Plugin().Dependencies() {
		if d.ID == id {
			check = true
			break
		}
	}

	if !check {
		panic(fmt.Sprintf("Invalid access, module '%s' is not a dependency of '%s'. Contact module provider.", string(id), pr.loadedPlugin.Plugin().ID()))
	}

	return parentRouters{moduleLoadingContext: pr, parentId: id}
}

type notFoundCtxKeyType string

var notFoundCtxKey = notFoundCtxKeyType("notfound")

func notFoundServeFunc(ctx context.Context, r *http.Request, p schema.ContextHandler, n schema.ContextHandler) (context.Context, error) {
	return n(ctx, r)
	//return context.WithValue(ctx, notFoundCtxKey, nil), nil
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

func newModuleContext() moduleContext {
	return moduleContext{
		allRouters: &routersImpl{
			bootstrapModule.ID(): pathSpecRoutersList{
				defaultPathSpec: emptyPathSpec,
				pathSpecServFuncListMap: map[schema.PathSpec]pluginServeFuncList{
					emptyPathSpec: pluginServeFuncList{
						pluginServeFunc{
							id:        bootstrapModule.ID(),
							serveFunc: notFoundServeFunc,
						},
					},
				},
			},
		},
	}
}

func LoadModules(ctx context.Context, modules []schema.Module) (context.Context, error) {
	// Add the bootstrap module to the beginning
	// Could also have been done to the tail, but adding it to front will help it
	// get all the modules resolve faster
	modules = append([]schema.Module{bootstrapModule}, modules...)

	var plugins = make([]plugin.Plugin, len(modules))

	for i := 0; i < len(modules); i++ {
		plugins[i] = modules[i]
	}
	ctx = context.WithValue(ctx, moduleCtxKey, newModuleContext())

	ctx, loadedPlugins, err := loader.Load(ctx, plugins, startModule)

	if err != nil {
		log.Println("Load incomplete,", loadedPlugins.Count(), "modules loaded")
	}
	return ctx, err

}

func startModule(ctx context.Context, lp *loader.LoadedPlugin) (context.Context, error) {

	mCtx := ctx.Value(moduleCtxKey).(moduleContext)

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

	mLoadingCtx := moduleLoadingContext{
		moduleContext: mCtx,
		loadedPlugin:  lp,
	}

	return module.Startup(ctx, mLoadingCtx)
}

type processContext struct {
	moduleCtx                 moduleContext
	currentModuleID           plugin.ID
	currentRoutersForModule   pathSpecRoutersList
	currentRoutersForPathSpec pluginServeFuncList
	funcIndex                 int
	uriParts                  []string
	uriIndex                  int
}

func Process(ctx context.Context, req *http.Request) (context.Context, error) {
	uri := req.RequestURI

	moduleCtx := ctx.Value(moduleCtxKey).(moduleContext)

	uriParts := strings.Split(uri, "/")
	uriIndex := 0
	pathSpec := schema.PathSpec(uriParts[uriIndex])
	currentModuleID := bootstrapModule.ID()
	currentRoutersForModule := (*moduleCtx.allRouters)[currentModuleID]
	currentRoutersForPathSpec := currentRoutersForModule.pathSpecServFuncListMap[pathSpec]
	funcIndex := len(currentRoutersForPathSpec) - 1

	pCtx := processContext{
		moduleCtx:                 moduleCtx,
		currentModuleID:           currentModuleID,
		currentRoutersForModule:   currentRoutersForModule,
		currentRoutersForPathSpec: currentRoutersForPathSpec,
		funcIndex:                 funcIndex,
		uriParts:                  uriParts,
		uriIndex:                  uriIndex,
	}

	return servReduce(ctx, req, pCtx)

}

func servReduce(ctx context.Context, req *http.Request, pctx processContext) (context.Context, error) {
	parentHandler := func(ctx context.Context, r *http.Request) (context.Context, error) {
		pctx.funcIndex--
		if pctx.funcIndex >= 0 {
			pctx.funcIndex--
			servReduce(ctx, req, pctx)
		}
		return ctx, nil
	}

	nextHandler := func(ctx context.Context, r *http.Request) (context.Context, error) {
		pctx.uriIndex++
		if pctx.uriIndex < len(pctx.uriParts) {
			uriFrag := schema.PathSpec(pctx.uriParts[pctx.uriIndex])
			servFuncList := pctx.currentRoutersForModule.pathSpecServFuncListMap[uriFrag]
			funcIndex := len(servFuncList) - 1
			if funcIndex < 0 {
				// No sub-paths defined, we ran out of routers before uri fragments
				// TODO: handle this in a generic way
				return ctx, nil
			}
			pctx.currentRoutersForPathSpec = servFuncList
			pctx.funcIndex = funcIndex
			pctx.currentModuleID = servFuncList[funcIndex].id
			pctx.currentRoutersForModule = (*pctx.moduleCtx.allRouters)[pctx.currentModuleID]

			return servReduce(ctx, req, pctx)
		} else {
			// No sub-routes defined
			// We ran out of paths, before sub-routes, so its just a no-op
			return ctx, nil
		}

	}

	psf := pctx.currentRoutersForPathSpec[pctx.funcIndex]

	log.Println("Handling part", pctx.uriIndex, "in", pctx.uriParts, "with", psf.id)
	return psf.serveFunc(ctx, req, parentHandler, nextHandler)
}

func Render(ctx context.Context, w http.ResponseWriter, req *http.Request) (context.Context, error) {
	return ctx, nil
}
