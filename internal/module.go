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
		id         plugin.ID
		serveFunc  schema.ServeFunc
		renderFunc schema.RenderFunc
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

func (psr pathSpecLoadingContext) AddRoute(ps schema.PathSpec, sf schema.ServeFunc, rf schema.RenderFunc) {
	currentPluginId := psr.loadedPlugin.Plugin().ID()
	all := *psr.allRouters
	psrl := all[psr.parentId]

	if len(psrl.pathSpecServFuncListMap) == 0 {
		psrl.defaultPathSpec = ps
	}

	psf := pluginServeFunc{
		id:         currentPluginId,
		serveFunc:  sf,
		renderFunc: rf,
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

func notFoundServeFunc(ctx context.Context, r *http.Request, p schema.ContextHandler) (interface{}, error) {
	return notFoundType{}, nil
}

func defaultRenderer(ctx context.Context, r *http.Request, w http.ResponseWriter, parent schema.Renderer) (interface{}, error) {

	intf := ctx.Value(schema.RenderResultKey)

	switch x := intf.(type) {
	case notFoundType:
		w.Write([]byte("Not Found"))
	case []byte:
		w.Write(x)
	default:
		panic("Cannot handle intf of type unknown")
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
							id:         bootstrapModule.ID(),
							serveFunc:  notFoundServeFunc,
							renderFunc: defaultRenderer,
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

type processContext struct {
	moduleCtx                 moduleLoadingContext
	currentModuleID           plugin.ID
	currentRoutersForModule   pathSpecRoutersList
	currentRoutersForPathSpec pluginServeFuncList
	funcIndex                 int
	uriParts                  []string
	uriIndex                  int
	depth                     int
}

type processContextKeyType string
type renderContextKeyType string

var processContextKey = processContextKeyType("processContext")

var renderContextKey = renderContextKeyType("renderContext")

func Process(ctx context.Context, req *http.Request) (context.Context, error) {
	uri := req.RequestURI

	moduleCtx := ctx.Value(moduleCtxKey).(moduleLoadingContext)

	uriParts := strings.Split(uri, "/")
	// if strings.HasPrefix(uri, "/") {
	// 	uriParts = uriParts[1:]
	// }
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

	pctxStack, ctx, err := servReduce(ctx, req, pCtx, []processContext{})
	ctx = context.WithValue(ctx, processContextKey, pctxStack)
	return ctx, err

}

func servReduce(ctx context.Context, req *http.Request,
	pctx processContext, pctxStack []processContext) ([]processContext, context.Context, error) {

	parentHandler := func(ctx context.Context, r *http.Request) (context.Context, error) {
		if pctx.funcIndex == 0 {
			return ctx, nil
		}
		// Make a copy
		parentCtx := pctx
		parentCtx.depth++
		parentCtx.funcIndex--

		// Stack does not grow when calling parent
		_, ctx, err := servReduce(ctx, req, parentCtx, pctxStack)
		return ctx, err
	}

	psf := pctx.currentRoutersForPathSpec[pctx.funcIndex]

	intf, err := psf.serveFunc(ctx, req, parentHandler)

	ctx, ok := intf.(context.Context)
	if !ok {
		ctx = context.WithValue(ctx, schema.ProcessResultKey, intf)
	}

	if err != nil {
		return pctxStack, ctx, err
	}

	// 'Next' processing starts here.
	// Proceed only if not processing a parent call.
	if pctx.depth > 0 {
		return pctxStack, ctx, nil
	}
	// Add current pctx to stack
	pctxStack = append(pctxStack, pctx)

	currentUriIndex := pctx.uriIndex + 1

	// If no more URI parts to handle
	if currentUriIndex == len(pctx.uriParts) {
		// We ran out of paths, before sub-routes, so just return current ctx

		return pctxStack, ctx, nil
	}

	uriFrag := schema.PathSpec(pctx.uriParts[currentUriIndex])

	for currentFuncIndex := pctx.funcIndex; currentFuncIndex >= 0; currentFuncIndex-- {
		nextModuleID := pctx.currentRoutersForPathSpec[currentFuncIndex].id
		routersForModule := (*pctx.moduleCtx.allRouters)[nextModuleID]
		servFuncList := routersForModule.pathSpecServFuncListMap[uriFrag]

		funcIndex := len(servFuncList) - 1
		if funcIndex >= 0 {
			pctx.currentModuleID = nextModuleID
			pctx.currentRoutersForModule = routersForModule
			pctx.currentRoutersForPathSpec = servFuncList
			pctx.funcIndex = funcIndex
			pctx.uriIndex = currentUriIndex

			return servReduce(ctx, req, pctx, pctxStack)
		}
	}

	intf, err = notFoundServeFunc(ctx, req, nil)
	return pctxStack, intf.(context.Context), err
}

func Render(ctx context.Context, req *http.Request, w http.ResponseWriter) (context.Context, error) {
	pCtxStack, ok := ctx.Value(processContextKey).([]processContext)
	if !ok {
		panic("Invalid context. Render called without a preceding Process")
	}
	var err error

	for i := len(pCtxStack) - 1; i >= 0; i-- {
		pctx := pCtxStack[i]
		pctx.funcIndex = len(pctx.currentRoutersForPathSpec) - 1
		ctx, err = renderer(ctx, req, w, pctx)
		if err != nil {
			break
		}

	}

	return ctx, err

}

func renderer(ctx context.Context, req *http.Request, w http.ResponseWriter, pctx processContext) (context.Context, error) {

	renderFunc := pctx.currentRoutersForPathSpec[pctx.funcIndex].renderFunc

	parentRenderer := func(ctx context.Context, r *http.Request, w http.ResponseWriter) (context.Context, error) {
		if pctx.funcIndex == 0 {
			return ctx, nil
		}

		parentCtx := pctx
		parentCtx.funcIndex--
		parentCtx.depth++
		return renderer(ctx, req, w, parentCtx)

	}

	intf, err := renderFunc(ctx, req, w, parentRenderer)

	ctxCast, ok := intf.(context.Context)
	if ok {
		ctx = ctxCast
	} else {
		ctx = context.WithValue(ctx, schema.RenderResultKey, intf)
	}

	return ctx, err
}
