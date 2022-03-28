package toml

type WorkflowDefintionTOML struct {
	Workflows []WorkflowTOML `toml:"Workflow"`
}

type WorkflowTOML struct {
	Identifier  string
	Active      bool
	Name        string
	Description string
	Trigger     WorkflowTriggerTOML
	Steps       []WorkflowStepTOML `toml:"Step"`
}

type WorkflowTriggerTOML struct {
	Name        string
	Description string
	Variety     string
	Meta        map[string]string
}

type WorkflowStepTOML struct {
	Active      bool
	Name        string
	Description string
	Variety     string
	Meta        map[string]string
}
