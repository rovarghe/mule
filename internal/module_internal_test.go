package internal

import (
	"context"
	"fmt"
	"net/http"
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

	ctx, err = Process(ctx, &http.Request{
		RequestURI: "/",
	})

	fmt.Println("Process returned", ctx.Value("core"))
	fmt.Println("Process returned", ctx.Value("about"))

	ctx, err = LoadModules(context.Background(), coreAndAboutModules())

	ctx, err = Process(ctx, &http.Request{
		RequestURI: "/about",
	})

	fmt.Println("Second process returned", ctx.Value("core"))
	fmt.Println("Second process returned", ctx.Value("about"))
}
