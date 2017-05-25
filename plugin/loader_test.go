package plugin

import (
	"testing"
)

func TestIsResolved(t *testing.T) {
	var mavenPluginNode = &pluginNode{
		plugin: &mavenPlugin,
	}

	var basePluginNode = &pluginNode{
		plugin: &basePlugin,
	}

	if mavenPluginNode.isResolved() {
		t.Error("Should be unrsolved")
	}

	mavenPluginNode.dependencies = map[*Dependency]*pluginNode{
		&mavenPlugin.Dependencies[0]: basePluginNode,
	}

	if !mavenPluginNode.isResolved() {
		t.Error("Should be resolved")
	}
}

func TestLoadPlugins(t *testing.T) {
	loadPlugins(&[]*Plugin{
		&basePlugin, &mavenPlugin,
	})

}
