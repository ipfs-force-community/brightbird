package types

// Property Property
// swagger:model property
type Property struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Value       string `json:"value"` //easy for front page
	Require     bool   `json:"require"`
}

// Property DependencyProperty
// swagger:model svcProperty
type DependencyProperty struct {
	Name  string     `json:"name"`
	Value string     `json:"value"`
	Type  PluginType `json:"type"`

	SockPath    string `json:"sockPath"`
	Description string `json:"description"`
	Require     bool   `json:"require"`
}

// SharedPropertyInNode just use to get shared field between deploynode and testitem, no pratical usage
type SharedPropertyInNode interface {
	GetName() string
	GetVersion() string
	GetProperties() []*Property
	GetDependencies() []*DependencyProperty
	GetInstance() *DependencyProperty
}

type DeployNode struct {
	// the name for this test flow
	// required: true
	// min length: 3
	Name string `json:"name"`
	// the version for this test flow
	// required: true
	// min length: 3
	Version      string                `json:"version"`
	Properties   []*Property           `json:"properties"`
	Dependencies []*DependencyProperty `json:"dependencies"`
	Instance     *DependencyProperty   `json:"instance"`
}

func (n DeployNode) GetName() string                        { return n.Name }
func (n DeployNode) GetVersion() string                     { return n.Version }
func (n DeployNode) GetProperties() []*Property             { return n.Properties }
func (n DeployNode) GetDependencies() []*DependencyProperty { return n.Dependencies }
func (n DeployNode) GetInstance() *DependencyProperty       { return n.Instance }

type TestItem struct {
	// the name for this test flow
	// required: true
	// min length: 3
	Name string `json:"name"`
	// the version for this test flow
	// required: true
	// min length: 3
	Version      string                `json:"version"`
	Properties   []*Property           `json:"properties"`
	Dependencies []*DependencyProperty `json:"dependencies"`
	Instance     *DependencyProperty   `json:"Instance"`
}

func (n TestItem) GetName() string                        { return n.Name }
func (n TestItem) GetVersion() string                     { return n.Version }
func (n TestItem) GetProperties() []*Property             { return n.Properties }
func (n TestItem) GetDependencies() []*DependencyProperty { return n.Dependencies }
func (n TestItem) GetInstance() *DependencyProperty       { return n.Instance }
