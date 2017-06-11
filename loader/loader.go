/*
Package loader provides functions to 'register' a Plugin in the order of its dependencies.

It can also 'unregister' plugins in the reverse order. The actual actions are configurable
through RegisterFunc and UnregisterFunc
*/
package loader

import (
	"context"
	"errors"
	"fmt"

	"github.com/rovarghe/mule/plugin"
)

type (
	// RegisterFunc is an initializer function
	// The context is created by the client
	RegisterFunc func(
		ctx context.Context,
		state *LoadedPlugin) (context.Context, error)

	// UnregisterFunc is expected to reverse the actions of RegisterFunc
	// There is no requirement the same Context be used as Load(), but normally it would be
	// the same.
	UnregisterFunc func(ctx context.Context, state *LoadedPlugin) (context.Context, error)

	// RegistrationState indicates whether the state of the registration.
	// It passes through 3 phases: DependenciesRegistered, PluginRegistered and DependentsRegistered
	RegistrationState int

	// NoRootsLoadError is returned if there are no plugins to be loaded, or if dependencies are circular
	NoRootsLoadError struct{}

	// UnresolvedDependency is returned within UnresolvedDependenciesLoadError for each dependency that could
	// not be satisfied
	UnresolvedDependency struct {
		Plugin     plugin.Plugin
		Dependency map[plugin.Dependency][]plugin.Plugin
	}

	// UnresolvedDependenciesLoadError is returned as the error type if the list of plugins are incomplete and
	// if any one of the plugins' dependencies chould not be satisfied
	UnresolvedDependenciesLoadError struct {
		unresolved *unresolvedType
		all        *map[plugin.ID]pluginList
	}

	pluginList []*LoadedPlugin

	unresolvedType map[*LoadedPlugin]interface{}

	// LoadedPlugins stores information about plugins as seen by the loader
	LoadedPlugins struct {
		unresolved unresolvedType
		loaded     pluginList
		roots      pluginList
		all        map[plugin.ID]pluginList
	}

	// LoadedPlugin stores information about a particular plugin as determined by
	// the loader.
	LoadedPlugin struct {
		plugin       plugin.Plugin
		dependencies map[plugin.Dependency]*LoadedPlugin
		dependents   pluginList
		state        RegistrationState
	}
)

const (
	// DependenciesRegistered state indicates that all the parents of the plugiin have been successfully
	// registered
	DependenciesRegistered RegistrationState = iota
	// PluginRegistered indicates the plugin itself has been registered
	PluginRegistered
	// DependentsRegistered indicates all the children of the plugin, i.e. all its dependents have been
	// successfully registered
	DependentsRegistered
)

func (n *LoadedPlugin) isResolved() bool {
	return len(n.plugin.Dependencies()) == len(n.dependencies)
}

func resolve(node *LoadedPlugin, all *map[plugin.ID]pluginList) bool {

	var unresolved = 0
	for _, d := range node.plugin.Dependencies() {
		flag := false
		versions := (*all)[d.ID]
		if versions != nil {
			// does any version satisfy dependency
			for _, v := range versions {
				if plugin.Satisfies(v.plugin, d) {
					// Link parent and child
					node.dependencies[d] = v
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

	return unresolved == 0
}

func flattenRoots(state *LoadedPlugins, ctx context.Context, RegisterFunc RegisterFunc) error {
	var seen = &map[*LoadedPlugin]bool{}

	for _, node := range state.roots {
		err := flattenPluginNodes(node, seen, state, ctx, RegisterFunc)
		if err != nil {
			return err
		}
	}
	return nil
}

func flattenPluginNodes(n *LoadedPlugin,
	seen *map[*LoadedPlugin]bool,
	state *LoadedPlugins,
	ctx context.Context,
	RegisterFunc RegisterFunc) error {

	// Walk down this node if all dependencies are already walked

	ii := 0
	for _, dn := range n.dependencies {
		if !(*seen)[dn] {
			return nil
		}
		ii++
	}

	state.loaded = append(state.loaded, n)
	n.state = DependenciesRegistered

	(*seen)[n] = true

	// Call the initializer
	ctx, err := RegisterFunc(ctx, n)
	if err != nil {
		return err
	}
	n.state = PluginRegistered

	for _, d := range n.dependents {
		if !(*seen)[d] {
			err = flattenPluginNodes(d, seen, state, ctx, RegisterFunc)
			if err != nil {
				return err
			}
		}
	}
	n.state = DependentsRegistered

	return nil
}

func (ur *unresolvedType) String() string {
	var sep = ""
	var str = ""
	for n := range *ur {

		str = fmt.Sprintf("%s%v%s", str, (*n).plugin, sep)
		sep = ","
	}
	return str
}

func (e NoRootsLoadError) Error() string {
	return "No roots detected"
}

// Error returns a formatted string error message
func (e UnresolvedDependenciesLoadError) Error() string {
	str := ""
	for _, ud := range e.UnresolvedDependencies() {
		str = fmt.Sprintln("Cannot resolve plugin [", ud.Plugin.ID(), ud.Plugin.Version(), "]")

		for d, p := range ud.Dependency {
			str = fmt.Sprintf("%s Missing dependency: %s\n", str, d.String())
			str = fmt.Sprintf("%s Candidates:\n", str)
			for _, plugin := range p {
				str = fmt.Sprintf("%s   %s %s\n", str, plugin.ID(), plugin.Version().String())
			}
		}

	}
	return str
}

// UnresolvedDependencies lists all the dependencies that could not be resolved
func (e *UnresolvedDependenciesLoadError) UnresolvedDependencies() []UnresolvedDependency {
	var list = []UnresolvedDependency{}
	for u := range *e.unresolved {
		var ud = UnresolvedDependency{
			Plugin:     u.plugin,
			Dependency: map[plugin.Dependency][]plugin.Plugin{},
		}
		for _, d := range u.plugin.Dependencies() {

			var candidates = []plugin.Plugin{}
			for _, c := range (*e.all)[d.ID] {
				candidates = append(candidates, c.plugin)
			}

			ud.Dependency[d] = candidates

		}
		list = append(list, ud)

	}
	return list
}

// Load goes through each plugin in order of its depedencies and pass
// it to the RegisterFunc to do whatever initialization it wants to do.
func Load(ctx context.Context, plugins []plugin.Plugin, RegisterFunc RegisterFunc) (context.Context, *LoadedPlugins, error) {
	state := &LoadedPlugins{
		unresolved: unresolvedType{},
		loaded:     pluginList{},
		roots:      pluginList{},
		all:        map[plugin.ID]pluginList{},
	}

	if len(plugins) == 0 {
		return ctx, state, nil
	}

	for _, p := range plugins {
		node := &LoadedPlugin{
			plugin:       p,
			dependencies: map[plugin.Dependency]*LoadedPlugin{},
			dependents:   pluginList{},
		}
		if !resolve(node, &state.all) {
			state.unresolved[node] = true
		}

		if len(p.Dependencies()) == 0 {
			state.roots = append(state.roots, node)
		}

		state.all[p.ID()] = append(state.all[p.ID()], node)

	}

	for nomas := true; nomas && len(state.unresolved) > 0; {
		nomas = false
		for n := range state.unresolved {
			if resolve(n, &state.all) {
				delete(state.unresolved, n)
				nomas = true
			}
		}
	}

	if len(state.unresolved) != 0 {
		return ctx, state, UnresolvedDependenciesLoadError{
			unresolved: &state.unresolved,
			all:        &state.all,
		}
	}

	if len(state.roots) == 0 {
		return ctx, state, errors.New("No roots")
	}

	err := flattenRoots(state, ctx, RegisterFunc)
	return ctx, state, err

}

// Unload deregisters the plugins that were successfuly registered by Load()
func (state *LoadedPlugins) Unload(ctx context.Context, unRegisterFunc UnregisterFunc) (context.Context, error) {
	var i = len(state.loaded)
	var err error
	for ; i > 0; i-- {
		ctx, err = unRegisterFunc(ctx, state.loaded[i-1])
		if err != nil {
			break
		}
	}
	state.loaded = state.loaded[:i]
	return ctx, err
}

func (state *LoadedPlugins) Count() int {
	return len(state.loaded)
}

func (state *LoadedPlugins) Get(i int) *LoadedPlugin {
	return state.loaded[i]
}

func (n *LoadedPlugin) Plugin() plugin.Plugin {
	return n.plugin
}

func (n *LoadedPlugin) Dependencies() map[plugin.Dependency]*LoadedPlugin {
	var ret = map[plugin.Dependency]*LoadedPlugin{}
	for d, dn := range n.dependencies {
		ret[d] = dn
	}
	return ret
}

func (n *LoadedPlugin) Dependents() []*LoadedPlugin {
	var ret = make([]*LoadedPlugin, len(n.dependents))

	for i, dn := range n.dependents {
		ret[i] = dn
	}
	return ret
}

func (n *LoadedPlugin) State() RegistrationState {
	return n.state
}
