package workflow

import (
	"encoding/json"
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

// WorkflowDefinition YAML/JSON workflow definition
type WorkflowDefinition struct {
	Name        string                 `yaml:"name" json:"name"`
	Description string                 `yaml:"description" json:"description"`
	Version     string                 `yaml:"version" json:"version"`
	Config      map[string]interface{} `yaml:"config" json:"config"`
	Steps       []StepDefinition       `yaml:"steps" json:"steps"`
}

// StepDefinition YAML/JSON step definition
type StepDefinition struct {
	ID         string                 `yaml:"id" json:"id"`
	Name       string                 `yaml:"name" json:"name"`
	Type       string                 `yaml:"type" json:"type"`
	ActionType string                 `yaml:"action_type,omitempty" json:"action_type,omitempty"`
	OnSuccess  []string               `yaml:"on_success,omitempty" json:"on_success,omitempty"`
	OnFailure  []string               `yaml:"on_failure,omitempty" json:"on_failure,omitempty"`
	Timeout    string                 `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	Retry      *RetryDefinition       `yaml:"retry,omitempty" json:"retry,omitempty"`
	Parameters map[string]interface{} `yaml:"parameters,omitempty" json:"parameters,omitempty"`
	Metadata   map[string]string      `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

// RetryDefinition YAML/JSON retry definition
type RetryDefinition struct {
	MaxAttempts int     `yaml:"max_attempts" json:"max_attempts"`
	Delay       string  `yaml:"delay" json:"delay"`
	BackoffRate float64 `yaml:"backoff_rate,omitempty" json:"backoff_rate,omitempty"`
}

// FromYAML creates a workflow from YAML
func FromYAML(data []byte, actionRegistry map[string]ActionFunc) (*Workflow, error) {
	var def WorkflowDefinition
	if err := yaml.Unmarshal(data, &def); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return buildWorkflowFromDefinition(&def, actionRegistry)
}

// FromJSON creates a workflow from JSON
func FromJSON(data []byte, actionRegistry map[string]ActionFunc) (*Workflow, error) {
	var def WorkflowDefinition
	if err := json.Unmarshal(data, &def); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return buildWorkflowFromDefinition(&def, actionRegistry)
}

// buildWorkflowFromDefinition builds workflow from definition
func buildWorkflowFromDefinition(def *WorkflowDefinition, actionRegistry map[string]ActionFunc) (*Workflow, error) {
	workflow := &Workflow{
		Name:        def.Name,
		Description: def.Description,
		Version:     def.Version,
		Config:      def.Config,
		Steps:       make([]Step, 0, len(def.Steps)),
		CreatedAt:   time.Now(),
	}

	for _, stepDef := range def.Steps {
		step, err := buildStepFromDefinition(&stepDef, actionRegistry)
		if err != nil {
			return nil, fmt.Errorf("failed to build step %s: %w", stepDef.ID, err)
		}
		workflow.Steps = append(workflow.Steps, *step)
	}

	return workflow, nil
}

// buildStepFromDefinition builds step from definition
func buildStepFromDefinition(def *StepDefinition, actionRegistry map[string]ActionFunc) (*Step, error) {
	step := &Step{
		ID:         def.ID,
		Name:       def.Name,
		Type:       StepType(def.Type),
		OnSuccess:  def.OnSuccess,
		OnFailure:  def.OnFailure,
		Parameters: def.Parameters,
		Metadata:   def.Metadata,
	}

	// Parse timeout
	if def.Timeout != "" {
		timeout, err := time.ParseDuration(def.Timeout)
		if err != nil {
			return nil, fmt.Errorf("invalid timeout: %w", err)
		}
		step.Timeout = timeout
	}

	// Parse retry policy
	if def.Retry != nil {
		delay, err := time.ParseDuration(def.Retry.Delay)
		if err != nil {
			return nil, fmt.Errorf("invalid retry delay: %w", err)
		}
		step.RetryPolicy = &RetryPolicy{
			MaxAttempts: def.Retry.MaxAttempts,
			Delay:       delay,
			BackoffRate: def.Retry.BackoffRate,
		}
	}

	// Get action from registry
	if def.ActionType != "" && actionRegistry != nil {
		if action, exists := actionRegistry[def.ActionType]; exists {
			step.Action = action
		}
	}

	return step, nil
}

// ToYAML exports workflow to YAML
func ToYAML(workflow *Workflow) ([]byte, error) {
	def := workflowToDefinition(workflow)
	return yaml.Marshal(def)
}

// ToJSON exports workflow to JSON
func ToJSON(workflow *Workflow) ([]byte, error) {
	def := workflowToDefinition(workflow)
	return json.MarshalIndent(def, "", "  ")
}

// workflowToDefinition converts workflow to definition
func workflowToDefinition(workflow *Workflow) *WorkflowDefinition {
	def := &WorkflowDefinition{
		Name:        workflow.Name,
		Description: workflow.Description,
		Version:     workflow.Version,
		Config:      workflow.Config,
		Steps:       make([]StepDefinition, 0, len(workflow.Steps)),
	}

	for _, step := range workflow.Steps {
		stepDef := StepDefinition{
			ID:         step.ID,
			Name:       step.Name,
			Type:       string(step.Type),
			OnSuccess:  step.OnSuccess,
			OnFailure:  step.OnFailure,
			Parameters: step.Parameters,
			Metadata:   step.Metadata,
		}

		if step.Timeout > 0 {
			stepDef.Timeout = step.Timeout.String()
		}

		if step.RetryPolicy != nil {
			stepDef.Retry = &RetryDefinition{
				MaxAttempts: step.RetryPolicy.MaxAttempts,
				Delay:       step.RetryPolicy.Delay.String(),
				BackoffRate: step.RetryPolicy.BackoffRate,
			}
		}

		def.Steps = append(def.Steps, stepDef)
	}

	return def
}
