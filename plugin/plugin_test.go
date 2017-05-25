package plugin

import (
	"testing"
)

var v1_0_0 = Version{1, 0, 0, ""}
var v1_0_0copy = Version{1, 0, 0, ""}
var v1_0_1 = Version{1, 0, 1, ""}
var v1_0_2 = Version{1, 0, 2, ""}
var v2_0_0rel = Version{2, 0, 0, "rel"}
var v2_0_0beta = Version{2, 0, 0, "beta"}

var provider0 = Provider{
	Name:    "rovarghe",
	Website: "github.com/rovarghe/mule",
}

var provider1 = Provider{
	Name:    "example",
	Website: "example.com/provider1",
}

var provider2 = Provider{
	Name:    "example",
	Website: "example.com/provider2",
}

var provider3 = Provider{
	Name:    "example",
	Website: "example.com/provider3",
}

var basePlugin = Plugin{
	ID:           PluginID("base"),
	Provider:     provider1,
	Dependencies: []Dependency{},
}

var mavenPluginCopy = Plugin{
	ID:       PluginID("maven"),
	Provider: provider2,
	Version:  v1_0_2,
	Dependencies: []Dependency{
		Dependency{
			ID: "base",
			Range: Range{
				Minimum:      v1_0_0,
				Maximum:      v1_0_2,
				MinInclusive: true,
				MaxInclusive: false,
			},
		},
	},
}

var mavenPlugin = Plugin{
	ID:       PluginID("maven"),
	Version:  v1_0_2,
	Provider: provider2,
	Dependencies: []Dependency{
		Dependency{
			ID: "base",
			Range: Range{
				Minimum:      v1_0_0,
				Maximum:      v1_0_2,
				MinInclusive: true,
				MaxInclusive: false},
		},
	},
}

func TestParseDependency(t *testing.T) {
	var table = []struct {
		d Dependency
		s string
	}{
		{
			Dependency{
				PluginID("foo"),
				Range{
					Minimum:      Version{1, 0, 0, ""},
					MinInclusive: true,
					Maximum:      Version{2, 0, 0, ""},
					MaxInclusive: true,
				},
			}, "foo [1.0.0,2.0.0]"}, // one space
		{
			Dependency{
				PluginID("foo"),
				Range{
					Minimum:      Version{1, 0, 0, ""},
					MinInclusive: false,
					Maximum:      Version{2, 0, 0, ""},
					MaxInclusive: true,
				},
			}, "foo(1.0.0,2.0.0]"}, // no space
		{
			Dependency{
				PluginID("foo"),
				Range{
					Minimum:      Version{1, 0, 0, ""},
					MinInclusive: false,
					Maximum:      Version{2, 0, 0, ""},
					MaxInclusive: false,
				},
			}, "foo (1.0.0,2.0.0)"},
	}

	for _, r := range table {
		d, err := ParseDependency(r.s)
		if err != nil {

			t.Error(err, "input=", r.s)
			continue
		}
		if *d != r.d {
			t.Error("Mismatched", *d, r.d)
		}

	}

}
func TestPluginSatisfiesDependency(t *testing.T) {

	var table = []struct {
		result    bool
		depedency Dependency
	}{
		{true, Dependency{
			ID: PluginID("maven"),
			Range: Range{
				Minimum:      v1_0_2,
				MinInclusive: true,
				Maximum:      v2_0_0beta,
				MaxInclusive: false,
			},
		}},
		{true, Dependency{
			ID: PluginID("maven"),
			Range: Range{
				Minimum:      v1_0_2,
				MinInclusive: true,
				Maximum:      v1_0_2,
				MaxInclusive: true,
			},
		}},
		{false, Dependency{
			ID: PluginID("maven"),
			Range: Range{
				Minimum:      v1_0_2,
				MinInclusive: false,
				Maximum:      v2_0_0beta,
				MaxInclusive: true,
			},
		}},
		{false, Dependency{
			ID: PluginID("maven"),
			Range: Range{
				Minimum:      v1_0_1,
				MinInclusive: true,
				Maximum:      v1_0_2,
				MaxInclusive: false,
			},
		}},
	}

	for _, r := range table {
		if mavenPlugin.Satisfies(&r.depedency) != r.result {
			t.Fatal("failure for dependency ", r.depedency, mavenPlugin)
		}
	}

}
func TestPluginEquals(t *testing.T) {
	if !mavenPlugin.Equals(&mavenPluginCopy) {
		t.Fail()
	}
}
