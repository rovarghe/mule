package internal

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/rovarghe/mule/internal/builtin"
	"github.com/rovarghe/mule/schema"
)

func onlyCoreModule() []schema.Module {
	return []schema.Module{builtin.CoreModule}
}

func coreAndAboutModules() []schema.Module {
	return append(onlyCoreModule(), builtin.AboutModule)
}

func TestBuiltInModules(t *testing.T) {
	ctx, err := LoadModules(context.Background(), onlyCoreModule())
	if err != nil {
		t.Error(err)
	}
	moduleLoadingCtx := ctx.Value(moduleCtxKey).(moduleLoadingContext)
	allRouters := *(moduleLoadingCtx.allRouters)

	if _, ok := allRouters[schema.RootModuleID]; !ok {
		t.Fatal("Root module not in allRouters")
	}

}

func TestProcess(t *testing.T) {
	ctx, err := LoadModules(context.Background(), onlyCoreModule())
	if err != nil {
		t.Error("Modules not loaded", err)
		return
	}

	Process(ctx, &http.Request{
		RequestURI: "/",
	})
}

func TestProcessAndRender(t *testing.T) {

	ctx, err := LoadModules(context.Background(), coreAndAboutModules())
	if err != nil {
		t.Error(err)
		return
	}

	httpReq := &http.Request{
		RequestURI: "/about",
		Header: http.Header{
			"Accept": []string{"application/json"},
		},
	}
	state, pctxStack, err := Process(ctx, httpReq)

	mockWriter := httptest.NewRecorder()
	state, err = Render(state, pctxStack, httpReq, mockWriter)

	fmt.Println(mockWriter.HeaderMap.Write(os.Stdout))
	fmt.Println(mockWriter.Body.String())
}
