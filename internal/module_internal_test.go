package internal

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/rovarghe/mule/internal/builtin"
	"github.com/rovarghe/mule/schema"
)

type MockHttpWriter struct {
	MockHeader http.Header
	bytes      *[]byte
}

func (w *MockHttpWriter) Write(b []byte) (int, error) {
	var bytes []byte
	if w.bytes == nil {
		bytes = b
	} else {
		bytes = append(*w.bytes, b...)
	}

	w.bytes = &bytes

	return len(b), nil

}

func (w MockHttpWriter) WriteHeader(i int) {
}

func (w MockHttpWriter) Header() http.Header {
	return w.MockHeader
}

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

	ctx, err = Process(ctx, &http.Request{
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
			"Content-Type": []string{"application/json"},
		},
	}
	ctx, err = Process(ctx, httpReq)
	mockWriter := MockHttpWriter{}
	ctx, err = Render(ctx, httpReq, &mockWriter)

	fmt.Println(string(*mockWriter.bytes))
}
