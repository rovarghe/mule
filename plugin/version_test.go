package plugin

import (
	"testing"
)

func TestVersionEquals(t *testing.T) {
	if v1_0_0copy != v1_0_0 {
		t.Fail()
	}
}

func TestVersionCompare(t *testing.T) {
	if v1_0_0.Compare(&v1_0_0copy) != 0 {
		t.Fatal("Equal compare failed")
	}
	if v1_0_0.Compare(&v1_0_1) != -1 {
		t.Fatal("Less than failure")
	}
	if v1_0_1.Compare(&v1_0_0) != 1 {
		t.Fatal("Greather than failure")
	}
	if v2_0_0beta.Compare(&v2_0_0rel) != -1 {
		t.Fatal("Label compare failed")
	}

}

func TestParseVersion(t *testing.T) {
	var table = []struct {
		v Version
		s string
	}{
		{Version{1, 0, 0, ""}, "1.0.0"},
		{Version{1, 0, 1, ""}, "1.0.1"},
		{Version{1, 0, 0, ""}, "1"},
		{Version{2, 0, 1, "beta"}, "2.0.1-beta"},
		{Version{2, 0, 1, "beta.012312"}, "2.0.1-beta.012312"},
	}

	for _, r := range table {
		if v, e := ParseVersion(r.s); e != nil || *v != r.v {
			t.Error("Failed for", r.s, e, v)
		}
	}

}
