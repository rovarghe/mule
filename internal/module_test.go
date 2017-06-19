package internal_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/rovarghe/mule/internal"
	"github.com/rovarghe/mule/internal/builtin"
	"github.com/rovarghe/mule/schema"
)

func listOfModules() []schema.Module {
	var modules = []schema.Module{}
	for _, m := range []schema.Module{
		builtin.CoreModule,
	} {
		modules = append(modules, m)
	}
	return modules

}
func TestBuiltInModules(t *testing.T) {
	_, err := internal.LoadModules(context.Background(), listOfModules())
	if err != nil {
		t.Error(err)
	}
}

func TestProcess(t *testing.T) {
	ctx, err := internal.LoadModules(context.Background(), listOfModules())
	if err != nil {
		t.Error("Modules not loaded", err)
		return
	}

	internal.Process(ctx, &http.Request{
		RequestURI: "/",
	})

}
