package internal

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
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

	loadingCtx, err := LoadModules(context.Background(), coreAndAboutModules())
	if err != nil {
		t.Error(err)
		return
	}

	httpReq := httptest.NewRequest("GET", "/about", strings.NewReader(""))
	httpReq.Header.Add("Accept", "application/json")

	state, processCtx, err := Process(loadingCtx, httpReq)

	mockWriter := httptest.NewRecorder()
	state, err = Render(state, processCtx, httpReq, mockWriter)

	fmt.Println(mockWriter.HeaderMap.Write(os.Stdout))
	fmt.Println(mockWriter.Body.String())
}
