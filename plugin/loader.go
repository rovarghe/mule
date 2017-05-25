package plugin

import (
	"errors"
	"fmt"
)

type pluginNode struct {
	plugin       *Plugin
	dependencies map[*Dependency]*pluginNode
	dependents   []*pluginNode
}

func (p *pluginNode) isResolved() bool {
	return len(p.plugin.Dependencies) == len(p.dependencies)
}

type pluginList []*pluginNode

func resolve(node *pluginNode, all *map[PluginID]pluginList) bool {

	var unresolved = 0
	for _, d := range node.plugin.Dependencies {
		flag := false
		versions := (*all)[d.ID]
		if versions != nil {
			// does any version satisfy dependency
			for _, v := range versions {
				if v.plugin.Satisfies(&d) {
					// Link parent and child
					node.dependencies[&d] = node
					v.dependents = append(v.dependents, node)
					flag = true
					break
				}
			}
		}
		if !flag {
			unresolved++
		}
	}

	return unresolved > 0
}

func flattenRoots(roots *pluginList) *[]*Plugin {
	var flattened = &[]*Plugin{}
	for _, node := range *roots {
		flattened = flattenPluginNodes(node, flattened)
	}
	return flattened
}

func flattenPluginNodes(n *pluginNode, list *[]*Plugin) *[]*Plugin {
	newlist := append(*list, n.plugin)
	for _, d := range n.dependents {
		flattenPluginNodes(d, &newlist)
	}
	return &newlist
}

func loadPlugins(plugins *[]*Plugin) (*[]*Plugin, error) {
	var unresolved map[*pluginNode]interface{}
	var roots pluginList
	var all = map[PluginID]pluginList{}

	if len(*plugins) == 0 {
		return &[]*Plugin{}, nil
	}

	for _, p := range *plugins {
		node := &pluginNode{
			plugin: p,
		}
		if !resolve(node, &all) {
			unresolved[node] = true
		}

		if len(p.Dependencies) == 0 {
			roots = append(roots, node)
		}

		all[p.ID] = append(all[p.ID], node)

	}

	for nomas := true; nomas && len(unresolved) > 0; {
		nomas = false
		for n := range unresolved {
			if resolve(n, &all) {
				delete(unresolved, n)
				nomas = true
			}
		}
	}

	if len(unresolved) != 0 {
		return nil, errors.New(fmt.Sprint("Unresolved Plugins", unresolved))
	}

	if len(roots) == 0 {
		return nil, errors.New("No roots")
	}

	return flattenRoots(&roots), nil

}
