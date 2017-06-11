package router

import (
	"testing"

	"github.com/rovarghe/mule/test"
)

func TestNewPathSpec(t *testing.T) {
	ps := NewPathSpec("foo/bar")

	test.Asserte(t, len(ps.pathArgs) == 1, "Unexpected parameters")

}
