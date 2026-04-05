package maven

// Basic Maven domain concepts for initial scaffolding.
// These types will be expanded for full modeling of modules, plugins, etc.

type Goal string

type Phase string

type Plugin struct {
	GroupID    string
	ArtifactID string
	Version    string
}

type MavenInvocation struct {
	Goals   []Goal
	Phases  []Phase
	Modules []string // module artifactIDs or paths
	Flags   []string // extra flags for advanced usage
}

type MavenOpts struct {
	NoClean    bool
	ResumeFrom string
}
