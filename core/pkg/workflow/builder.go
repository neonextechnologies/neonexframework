package workflow

import (
	"time"
)

// WorkflowBuilder fluent API for building workflows
type WorkflowBuilder struct {
	workflow *Workflow
}

// StepBuilder fluent API for building steps
type StepBuilder struct {
	step     *Step
	workflow *Workflow
	builder  *WorkflowBuilder
}

// NewWorkflowBuilder creates a new workflow builder
func NewWorkflowBuilder(name string) *WorkflowBuilder {
	return &WorkflowBuilder{
		workflow: &Workflow{
			Name:      name,
			Steps:     make([]Step, 0),
			Config:    make(map[string]interface{}),
			CreatedAt: time.Now(),
		},
	}
}

// Description sets workflow description
func (b *WorkflowBuilder) Description(desc string) *WorkflowBuilder {
	b.workflow.Description = desc
	return b
}

// Version sets workflow version
func (b *WorkflowBuilder) Version(version string) *WorkflowBuilder {
	b.workflow.Version = version
	return b
}

// Config sets workflow configuration
func (b *WorkflowBuilder) Config(key string, value interface{}) *WorkflowBuilder {
	b.workflow.Config[key] = value
	return b
}

// AddStep adds a new step to the workflow
func (b *WorkflowBuilder) AddStep(id, name string) *StepBuilder {
	step := &Step{
		ID:         id,
		Name:       name,
		Type:       StepTypeTask,
		Parameters: make(map[string]interface{}),
		Metadata:   make(map[string]string),
	}

	return &StepBuilder{
		step:     step,
		workflow: b.workflow,
		builder:  b,
	}
}

// Build builds the workflow
func (b *WorkflowBuilder) Build() *Workflow {
	return b.workflow
}

// Type sets step type
func (s *StepBuilder) Type(stepType StepType) *StepBuilder {
	s.step.Type = stepType
	return s
}

// Action sets step action function
func (s *StepBuilder) Action(action ActionFunc) *StepBuilder {
	s.step.Action = action
	return s
}

// Condition sets step condition function
func (s *StepBuilder) Condition(condition ConditionFunc) *StepBuilder {
	s.step.Condition = condition
	return s
}

// OnSuccess sets next step IDs on success
func (s *StepBuilder) OnSuccess(stepIDs ...string) *StepBuilder {
	s.step.OnSuccess = stepIDs
	return s
}

// OnFailure sets next step IDs on failure
func (s *StepBuilder) OnFailure(stepIDs ...string) *StepBuilder {
	s.step.OnFailure = stepIDs
	return s
}

// Timeout sets step timeout
func (s *StepBuilder) Timeout(timeout time.Duration) *StepBuilder {
	s.step.Timeout = timeout
	return s
}

// Retry sets retry policy
func (s *StepBuilder) Retry(maxAttempts int, delay time.Duration, backoffRate float64) *StepBuilder {
	s.step.RetryPolicy = &RetryPolicy{
		MaxAttempts: maxAttempts,
		Delay:       delay,
		BackoffRate: backoffRate,
	}
	return s
}

// Parameter sets step parameter
func (s *StepBuilder) Parameter(key string, value interface{}) *StepBuilder {
	s.step.Parameters[key] = value
	return s
}

// Metadata sets step metadata
func (s *StepBuilder) Metadata(key, value string) *StepBuilder {
	s.step.Metadata[key] = value
	return s
}

// End completes the step and returns to workflow builder
func (s *StepBuilder) End() *WorkflowBuilder {
	s.workflow.Steps = append(s.workflow.Steps, *s.step)
	return s.builder
}

// Then adds another step
func (s *StepBuilder) Then(id, name string) *StepBuilder {
	// Complete current step
	s.workflow.Steps = append(s.workflow.Steps, *s.step)

	// Create new step
	newStep := &Step{
		ID:         id,
		Name:       name,
		Type:       StepTypeTask,
		Parameters: make(map[string]interface{}),
		Metadata:   make(map[string]string),
	}

	return &StepBuilder{
		step:     newStep,
		workflow: s.workflow,
		builder:  s.builder,
	}
}
