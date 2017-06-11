package router

import (
	"context"
	"fmt"
	"net/http"
	"strings"
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
	w.WriteHeader(404)
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
