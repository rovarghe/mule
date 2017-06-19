package internal_test

import (
	"testing"

	"github.com/rovarghe/mule/internal"
	"github.com/rovarghe/mule/internal/builtin"
	"github.com/rovarghe/mule/schema"
)

func TestBuiltInModules(t *testing.T) {
	for _, m := range []schema.Module{
		builtin.CoreModule,
	} {
		internal.AddModule(m)
	}

	internal.LoadModules()
}
