package plugin_test

import (
	"testing"

	"github.com/rovarghe/mule/plugin"
	"github.com/rovarghe/mule/test/harness"
)

func TestVersionEquals(t *testing.T) {
	if harness.V1_0_0copy != harness.V1_0_0 {
		t.Fail()
	}
}

func TestVersionCompare(t *testing.T) {
	if harness.V1_0_0.Compare(harness.V1_0_0copy) != 0 {
		t.Fatal("Equal compare failed")
	}
	if harness.V1_0_0.Compare(harness.V1_0_1) != -1 {
		t.Fatal("Less than failure")
	}
	if harness.V1_0_1.Compare(harness.V1_0_0) != 1 {
		t.Fatal("Greather than failure")
	}
	if harness.V2_0_0beta.Compare(harness.V2_0_0rel) != -1 {
		t.Fatal("Label compare failed")
	}

}

func TestParseVersion(t *testing.T) {
	var table = []struct {
		v plugin.Version
		s string
	}{
		{plugin.Version{1, 0, 0, ""}, "1.0.0"},
		{plugin.Version{1, 0, 1, ""}, "1.0.1"},
		{plugin.Version{1, 0, 0, ""}, "1"},
		{plugin.Version{2, 0, 1, "beta"}, "2.0.1-beta"},
		{plugin.Version{2, 0, 1, "beta.012312"}, "2.0.1-beta.012312"},
	}

	for _, r := range table {
		if v, e := plugin.ParseVersion(r.s); e != nil || *v != r.v {
			t.Error("Failed for", r.s, e, v)
		}
	}

	v, e := plugin.ParseRange("[1.2.3,1.2.3]")
	if e != nil {
		t.Error(e)
	}
	switch {
	case v.Maximum.Compare(plugin.Version{1, 2, 3, ""}) != 0:
		t.Error("Max version range parse error")
	case v.Minimum.Compare(plugin.Version{1, 2, 3, ""}) != 0:
		t.Error("Min version range parse error")
	case v.MaxInclusive != true:
		t.Error("Max inclusive parse error")
	case v.MinInclusive != true:
		t.Error("Min inclusive parse error")
	}

}
