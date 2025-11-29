package agents

import "fmt"

// Agent represents a delegateable persona.
type Agent struct {
	Name        string `json:"name" yaml:"name"`
	Persona     string `json:"persona" yaml:"persona"`
	Description string `json:"description" yaml:"description"`
}

// Validate ensures required fields are present.
func (a Agent) Validate() error {
	if a.Name == "" {
		return fmt.Errorf("name is required")
	}
	if a.Persona == "" {
		return fmt.Errorf("persona is required for agent %q", a.Name)
	}
	if a.Description == "" {
		return fmt.Errorf("description is required for agent %q", a.Name)
	}
	return nil
}
