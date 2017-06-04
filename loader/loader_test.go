package loader_test

import (
	"context"
	"testing"

	"github.com/rovarghe/mule/loader"
	"github.com/rovarghe/mule/plugin"
	"github.com/rovarghe/mule/test/harness"
)

func registerFunc(t *testing.T) loader.RegisterFunc {
	return func(ctx context.Context, ps loader.LoadedPlugin) (context.Context, error) {

		t.Log("Loading plugin", ps.Plugin().ID)
		if ps.State() != loader.DependenciesRegistered {
			t.Error("Unexpected state, dependencies shuld be loaded", ps.Plugin())
		}
		return ctx, nil
	}
}

func unregisterFunc(t *testing.T) loader.UnregisterFunc {
	return func(ctx context.Context, ps loader.LoadedPlugin) (context.Context, error) {
		t.Log("Unloading plugin", ps.Plugin().ID)
		if ps.State() != loader.PluginRegistered && ps.State() != loader.DependentsRegistered {
			t.Error("Unexpected state, expecting PluginLoadedState", ps.State(), ps.Plugin().ID, ps.Plugin().Version)
		}
		return ctx, nil
	}
}

func TestLoadPlugins(t *testing.T) {
	var plugins = []*plugin.Plugin{
		&harness.MavenPlugin, &harness.BasePlugin,
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, 1, "FOO")
	ctx, loaded, err := loader.Load(ctx, &plugins, registerFunc(t))
	if err != nil {
		t.Fatal(err)
	}

	if loaded.Get(0).Plugin() != &harness.BasePlugin {
		t.Fatal("Base should be at 0")
	}
	if loaded.Get(1).Plugin() != &harness.MavenPlugin {
		t.Fatal("Maven plugin should be at 1")
	}
}

func TestLoadPlugins2(t *testing.T) {

	var plugins = []*plugin.Plugin{
		&harness.MvnTestReportsPlugin,
		&harness.MavenArtifactPlugin,
		&harness.MavenTestPlugin,
		&harness.MavenPlugin,
		&harness.GitPlugin,
		&harness.BasePlugin,
	}
	/*
		for i, p := range plugins {
			fmt.Println(i, *p)
		}
	*/
	ctx := context.WithValue(context.Background(), 2, "BAR")
	ctx, loaded, err := loader.Load(ctx, &plugins, registerFunc(t))
	loaded.Unload(ctx, unregisterFunc(t))
	if err != nil {
		switch err.(type) {
		case loader.UnresolvedDependenciesLoadError:
			t.Fatal(err)
		case loader.NoRootsLoadError:
			t.Fatal(err)
		default:
			t.Fatal(err)
		}
	}

	/*
		for i, p := range plugins {
			fmt.Println(i, *p)
		}
	*/

	var orders = map[plugin.PluginID]bool{}
	for i := 0; i < loaded.Count(); i++ {
		p := loaded.Get(i).Plugin()
		checked := false
		switch p.ID {
		case "base":
			if i != 0 {
				t.Error("Expected base")
			}
			checked = true
		case "maven":
			if !orders["base"] {
				t.Error("Expected base before maven")
			}
			checked = true
		case "git":
			if !orders["base"] {
				t.Error("Expected base before git")
			}
			checked = true

		case "maven-test":
			if orders["base"] && orders["maven"] {
				checked = true
			} else {
				t.Error("Expected base and maven before maven-test")
				checked = true
			}
		case "maven-artifact":
			if !orders["mvn-test-reports"] {
				t.Error("Expected mvn-test-reports before maven-artifact")
			}
			checked = true
		case "mvn-test-reports":
			if !orders["maven-test"] {
				t.Error("Expected maven-test before maven-test-reports")
			}
			checked = true

		}

		orders[p.ID] = true
		if !checked {
			t.Error("Failed check", i, p.ID)
		}
	}

}
