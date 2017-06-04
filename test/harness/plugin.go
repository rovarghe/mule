package harness

import "github.com/rovarghe/mule/plugin"

var V1_0_0 = plugin.Version{1, 0, 0, ""}
var V1_0_0copy = plugin.Version{1, 0, 0, ""}
var V1_0_1 = plugin.Version{1, 0, 1, ""}
var V1_0_2 = plugin.Version{1, 0, 2, ""}
var V2_0_0rel = plugin.Version{2, 0, 0, "rel"}
var V2_0_0beta = plugin.Version{2, 0, 0, "beta"}

var provider0 = plugin.Provider{
	Name: "rovarghe",
	URL:  "github.com/rovarghe/mule",
}

var provider1 = plugin.Provider{
	Name: "example",
	URL:  "example.com/provider1",
}

var provider2 = plugin.Provider{
	Name: "example",
	URL:  "example.com/provider2",
}

var provider3 = plugin.Provider{
	Name: "example",
	URL:  "example.com/provider3",
}

var BasePlugin = plugin.Plugin{
	ID:           plugin.ID("base"),
	Provider:     provider1,
	Dependencies: []plugin.Dependency{},
	Version:      V1_0_0,
}

var MavenPluginCopy = plugin.Plugin{
	ID:       plugin.ID("maven"),
	Provider: provider2,
	Version:  V1_0_2,
	Dependencies: []plugin.Dependency{
		plugin.Dependency{
			ID: "base",
			Range: plugin.Range{
				Minimum:      V1_0_0,
				Maximum:      V1_0_2,
				MinInclusive: true,
				MaxInclusive: false,
			},
		},
	},
}

var MavenPlugin = plugin.Plugin{
	ID:       plugin.ID("maven"),
	Version:  V1_0_2,
	Provider: provider2,
	Dependencies: []plugin.Dependency{
		plugin.Dependency{
			ID: "base",
			Range: plugin.Range{
				Minimum:      V1_0_0,
				Maximum:      V1_0_2,
				MinInclusive: true,
				MaxInclusive: false},
		},
	},
}

var MavenTestPlugin = plugin.Plugin{
	ID:       plugin.ID("maven-test"),
	Version:  V1_0_1,
	Provider: provider2,
	Dependencies: []plugin.Dependency{
		plugin.Dependency{
			ID: "maven",
			Range: plugin.Range{
				Minimum: V1_0_0,
				Maximum: V2_0_0beta,
			},
		},
	},
}

var MavenArtifactPlugin = plugin.Plugin{
	ID:       plugin.ID("maven-artifact"),
	Version:  V1_0_0,
	Provider: provider2,
	Dependencies: []plugin.Dependency{
		plugin.Dependency{
			ID: "maven",
			Range: plugin.Range{
				Minimum:      V1_0_0,
				Maximum:      V1_0_2,
				MaxInclusive: true,
			},
		},
		plugin.Dependency{
			ID: "mvn-test-reports",
			Range: plugin.Range{
				Minimum:      V1_0_0,
				Maximum:      V1_0_0,
				MinInclusive: true,
				MaxInclusive: true,
			},
		},
	},
}

var MvnTestReportsPlugin = plugin.Plugin{
	ID:       plugin.ID("mvn-test-reports"),
	Version:  V1_0_0,
	Provider: provider2,
	Dependencies: []plugin.Dependency{
		plugin.Dependency{
			ID: "maven-test",
			Range: plugin.Range{
				Minimum:      V1_0_1,
				Maximum:      V1_0_1,
				MinInclusive: true,
				MaxInclusive: true,
			},
		},
	},
}

var GitPlugin = plugin.Plugin{
	ID:       plugin.ID("git"),
	Version:  V1_0_2,
	Provider: provider3,
	Dependencies: []plugin.Dependency{
		plugin.Dependency{
			ID: "base",
			Range: plugin.Range{
				Minimum:      V1_0_0,
				Maximum:      V1_0_1,
				MinInclusive: true,
				MaxInclusive: true,
			},
		},
	},
}
