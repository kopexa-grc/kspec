package core

// Policy represents a collection of checks and groups.
type Policy struct {
	APIVersion    string        `yaml:"apiVersion,omitempty"`
	Kind          string        `yaml:"kind,omitempty"`
	Metadata      Metadata      `yaml:"metadata,omitempty"`
	Groups        []Group       `yaml:"groups,omitempty"`
	Queries       []Check       `yaml:"queries,omitempty"`
	Require       []Requirement `yaml:"require,omitempty"`
	Authors       []Author      `yaml:"authors,omitempty"`
	ScoringSystem string        `yaml:"scoring_system,omitempty"`
}

type Metadata struct {
	Name    string            `yaml:"name"`
	Version string            `yaml:"version,omitempty"`
	Title   string            `yaml:"title,omitempty"`
	License string            `yaml:"license,omitempty"`
	Tags    map[string]string `yaml:"tags,omitempty"`
}

type Requirement struct {
	Provider string `yaml:"provider"`
}

type Author struct {
	Name  string `yaml:"name"`
	Email string `yaml:"email,omitempty"`
}

type Group struct {
	Title  string  `yaml:"title"`
	Filter string  `yaml:"filter,omitempty"` // CEL expression
	Checks []Check `yaml:"checks"`
}

type Check struct {
	UID   string `yaml:"uid,omitempty"`
	ID    string `yaml:"id,omitempty"`
	Title string `yaml:"title,omitempty"`
	// Resource type to query against. e.g. "package", "service"
	// In CEL, we need to know WHICH resource to fetch.
	Resource string `yaml:"resource,omitempty"`
	// Config for the resource fetch, e.g. name=foo
	Config map[string]string `yaml:"config,omitempty"`

	Query       string `yaml:"query,omitempty"` // CEL expression
	Remediation string `yaml:"remediation,omitempty"`
	Severity    string `yaml:"severity,omitempty"`

	Docs  string `yaml:"docs,omitempty"`
	Audit string `yaml:"audit,omitempty"`
	Props []Prop `yaml:"props,omitempty"`
}

type Prop struct {
	UID   string      `yaml:"uid,omitempty"`
	Title string      `yaml:"title,omitempty"`
	MQL   interface{} `yaml:"mql,omitempty"`
}
