package internal

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/rovarghe/mule/plugin"
	"github.com/rovarghe/mule/schema"
)

type (
	pathVariables string
	sanitizedPath []string

	pathParameter struct {
		name  string
		regex string
	}

	pathSpec struct {
		sanitizedPath []string
		pathArgs      map[int]pathParameter
	}

	// NextHandler is called to invoke the next function in the chain
	NextHandler func(context.Context, http.ResponseWriter, *http.Request)
)

const (
	pathVariablesKey = pathVariables("pathVariables")
)

// PathParam retrieves the path parameters from the context
// Returns empty string if not found
func PathParam(ctx context.Context, param string) string {
	params := PathParams(ctx)
	if params == nil {
		return ""
	}
	return (*params)[param]
}

// PathParams retrieves all the path parameters from the context
func PathParams(ctx context.Context) *map[string]string {
	params, ok := ctx.Value(pathVariablesKey).(*map[string]string)
	if !ok {
		return nil
	}
	return params
}

// SetPathParam sets parameters in the context
func SetPathParam(ctx context.Context, param string, value string) context.Context {
	params := PathParams(ctx)

	if params == nil {
		params = &map[string]string{}
		ctx = context.WithValue(ctx, pathVariablesKey, params)
	}
	(*params)[param] = value
	return ctx
}

// DefaultNextHandler writes 404 to the writer
func DefaultNextHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

func extractPathParameter(str string) *pathParameter {
	if strings.HasPrefix(str, "{") && strings.HasSuffix(str, "}") {
		var name, regex string
		if i := strings.Index(str, ":"); i < 0 {
			name = str[1 : len(str)-1]
			regex = "[^/]+"
		} else {
			name = str[1:i]
			regex = str[i+1:]
		}
		return &pathParameter{
			name:  name,
			regex: regex,
		}
	}
	return nil

}

func newPathSpec(path string) pathSpec {
	sp := strings.Split(path, "/")
	pa := make(map[int]pathParameter)

	for i := 0; i < len(sp); i++ {
		if param := extractPathParameter(sp[i]); param != nil {
			sp[i] = fmt.Sprintf("{%d}", len(pa))
			pa[len(pa)] = *param
		}
	}
	return pathSpec{
		sanitizedPath: sp,
		pathArgs:      pa,
	}

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

func (rctx renderContext) PathParameters() map[string]string {
	return map[string]string{}
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
