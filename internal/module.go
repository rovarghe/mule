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

func (pctx processContext) URI() schema.PathSpec {
	uri := strings.Join(pctx.uriParts[0:pctx.uriIndex], "/")
	return schema.PathSpec(uri)
}

func (pctx processContext) Final() bool {
	return pctx.uriIndex == len(pctx.uriParts)-1
}

func (pctx processContext) PathParameters() map[string]string {
	return map[string]string{}
}

type processContextKeyType string

//type renderContextKeyType string

var processContextKey = processContextKeyType("processContext")

//var renderContextKey = renderContextKeyType("renderContext")

func Process(ctx context.Context, req *http.Request) (schema.State, context.Context, error) {
	uri := req.RequestURI

	moduleCtx := ctx.Value(moduleCtxKey).(moduleLoadingContext)

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

	pctxStack, state, err := stateReduce(ctx, req, pCtx, []processContext{})
	if err != nil {
		return nil, nil, err
	}
	ctx = context.WithValue(ctx, processContextKey, pctxStack)

	return state, ctx, err

}

func stateReduce(state schema.State, req *http.Request,
	pctx processContext, pctxStack []processContext) ([]processContext, schema.State, error) {

	parentHandler := func(state schema.State, r *http.Request) (schema.State, error) {
		if pctx.funcIndex == 0 {
			return state, nil
		}
		// Make a copy
		parentCtx := pctx
		parentCtx.depth++
		parentCtx.funcIndex--

		// Stack does not grow when calling parent
		_, state, err := stateReduce(state, req, parentCtx, pctxStack)
		return state, err
	}

	psf := pctx.currentRoutersForPathSpec[pctx.funcIndex]

	state, err := psf.stateReducer(state, pctx, req, parentHandler)

	if err != nil {
		return pctxStack, state, err
	}

	// 'Next' processing starts here.
	// Proceed only if not processing a parent call.
	if pctx.depth > 0 {
		return pctxStack, state, nil
	}
	// Add current pctx to stack
	pctxStack = append(pctxStack, pctx)

	currentUriIndex := pctx.uriIndex + 1

	// If no more URI parts to handle
	if currentUriIndex == len(pctx.uriParts) {
		// We ran out of paths, before sub-routes, so just return current ctx

		return pctxStack, state, nil
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

			return stateReduce(state, req, pctx, pctxStack)
		}
	}

	state, err = notFoundServeFunc(state, pctx, req, nil)
	return pctxStack, state, err
}

type renderContext processContext

func (rctx renderContext) URI() schema.PathSpec {
	return processContext(rctx).URI()
}

func (rctx renderContext) Final() bool {
	return rctx.uriIndex == 0
}

func Render(state schema.State, processCtx context.Context, req *http.Request, w http.ResponseWriter) (schema.State, error) {

	var err error

	fmt.Println(req)

	pCtxStack := processCtx.Value(processContextKey).([]processContext)
	if pCtxStack == nil {
		panic(fmt.Errorf("Render called without a process context"))
	}

	for i := len(pCtxStack) - 1; i >= 0; i-- {
		rctx := renderContext(pCtxStack[i])
		rctx.funcIndex = len(rctx.currentRoutersForPathSpec) - 1
		state, err = renderer(state, req, w, rctx)
		if err != nil {
			break
		}

	}

	return state, err
}

func renderer(state schema.State, req *http.Request, w http.ResponseWriter, rctx renderContext) (schema.State, error) {

	renderReducer := rctx.currentRoutersForPathSpec[rctx.funcIndex].renderReducer

	parentRenderer := func(state schema.State, r *http.Request, w http.ResponseWriter) (schema.State, error) {
		if rctx.funcIndex == 0 {
			return state, nil
		}

		parentCtx := rctx
		parentCtx.funcIndex--
		parentCtx.depth++
		return renderer(state, req, w, parentCtx)

	}

	state, err := renderReducer(state, rctx, req, w, parentRenderer)

	return state, err
}
