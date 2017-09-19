package internal

import (
	"fmt"
	"testing"

	"github.com/rovarghe/mule/test"
)

func TestNewPathSpec(t *testing.T) {
	ps := newPathSpec("foo/bar")

	test.Asserte(t, len(ps.pathArgs) == 0, "Unexpected parameters")

	ps = newPathSpec("foo/{bar}")

	test.Asserte(t, len(ps.pathArgs) == 1, "Expecting 1 parameter")
	fmt.Printf("%v\n", ps)

}
