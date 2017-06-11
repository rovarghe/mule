package loader

import (
	"testing"

	"github.com/rovarghe/mule/plugin"

	"github.com/rovarghe/mule/test/harness"
)

func TestIsResolved(t *testing.T) {

	var mavenPluginNode = &LoadedPlugin{
		plugin: harness.MavenPlugin,
	}

	var basePluginNode = &LoadedPlugin{
		plugin: harness.BasePlugin,
	}

	if mavenPluginNode.isResolved() {
		t.Error("Should be unrsolved")
	}

	mavenPluginNode.dependencies = map[plugin.Dependency]*LoadedPlugin{

		harness.MavenPlugin.Dependencies()[0]: basePluginNode,
	}

	if !mavenPluginNode.isResolved() {
		t.Error("Should be resolved")
	}
}
