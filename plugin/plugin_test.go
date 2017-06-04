package plugin_test

import (
	"strings"
	"testing"

	"github.com/rovarghe/mule/plugin"
	"github.com/rovarghe/mule/test/harness"
)

func TestParseDependency(t *testing.T) {
	var table = []struct {
		d plugin.Dependency
		s string
	}{
		{
			plugin.Dependency{
				plugin.ID("foo"),
				plugin.Range{
					Minimum:      plugin.Version{1, 0, 0, ""},
					MinInclusive: true,
					Maximum:      plugin.Version{2, 0, 0, ""},
					MaxInclusive: true,
				},
			}, "foo [1.0.0,2.0.0]"}, // one space
		{
			plugin.Dependency{
				plugin.ID("foo"),
				plugin.Range{
					Minimum:      plugin.Version{1, 0, 0, ""},
					MinInclusive: false,
					Maximum:      plugin.Version{2, 0, 0, ""},
					MaxInclusive: true,
				},
			}, "foo(1.0.0,2.0.0]"}, // no space
		{
			plugin.Dependency{
				plugin.ID("foo"),
				plugin.Range{
					Minimum:      plugin.Version{1, 0, 0, ""},
					MinInclusive: false,
					Maximum:      plugin.Version{2, 0, 0, ""},
					MaxInclusive: false,
				},
			}, "foo (1.0.0,2.0.0)"},
	}

	for _, r := range table {
		d, err := plugin.ParseDependency(r.s)
		if err != nil {

			t.Error(err, "input=", r.s)
			continue
		}
		if *d != r.d {
			t.Error("Mismatched", *d, r.d)
		}
		s_a, s_e := strings.Replace(d.String(), " ", "", -1), strings.Replace(r.s, " ", "", -1)

		if s_a != s_e {
			t.Error("String representation error, expecting", s_e, "got", s_a)
		}

	}

}

func TestPluginSatisfiesDependency(t *testing.T) {

	var table = []struct {
		result    bool
		depedency plugin.Dependency
	}{
		{true, plugin.Dependency{
			ID: plugin.ID("maven"),
			Range: plugin.Range{
				Minimum:      harness.V1_0_2,
				MinInclusive: true,
				Maximum:      harness.V2_0_0beta,
				MaxInclusive: false,
			},
		}},
		{true, plugin.Dependency{
			ID: plugin.ID("maven"),
			Range: plugin.Range{
				Minimum:      harness.V1_0_2,
				MinInclusive: true,
				Maximum:      harness.V1_0_2,
				MaxInclusive: true,
			},
		}},
		{false, plugin.Dependency{
			ID: plugin.ID("maven"),
			Range: plugin.Range{
				Minimum:      harness.V1_0_2,
				MinInclusive: false,
				Maximum:      harness.V2_0_0beta,
				MaxInclusive: true,
			},
		}},
		{false, plugin.Dependency{
			ID: plugin.ID("maven"),
			Range: plugin.Range{
				Minimum:      harness.V1_0_1,
				MinInclusive: true,
				Maximum:      harness.V1_0_2,
				MaxInclusive: false,
			},
		}},
	}

	for _, r := range table {
		if harness.MavenPlugin.Satisfies(&r.depedency) != r.result {
			t.Fatal("failure for dependency ", r.depedency, harness.MavenPlugin)
		}
	}

}
func TestPluginEquals(t *testing.T) {
	if !harness.MavenPlugin.Equals(&harness.MavenPluginCopy) {
		t.Fail()
	}
}
