package workflow

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// WorkflowStatus represents workflow execution status
type WorkflowStatus string

const (
	StatusPending   WorkflowStatus = "pending"
	StatusRunning   WorkflowStatus = "running"
	StatusCompleted WorkflowStatus = "completed"
	StatusFailed    WorkflowStatus = "failed"
	StatusCancelled WorkflowStatus = "cancelled"
	StatusPaused    WorkflowStatus = "paused"
)

// Workflow represents a workflow definition
type Workflow struct {
	ID          string
	Name        string
	Description string
	Version     string
	Steps       []Step
	Config      map[string]interface{}
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Step represents a workflow step
type Step struct {
	ID           string
	Name         string
	Type         StepType
	Action       ActionFunc
	Condition    ConditionFunc
	OnSuccess    []string // Next step IDs on success
	OnFailure    []string // Next step IDs on failure
	RetryPolicy  *RetryPolicy
	Timeout      time.Duration
	Parameters   map[string]interface{}
	Metadata     map[string]string
}

// StepType represents the type of step
type StepType string

const (
	StepTypeTask      StepType = "task"
	StepTypeCondition StepType = "condition"
	StepTypeParallel  StepType = "parallel"
	StepTypeLoop      StepType = "loop"
	StepTypeWait      StepType = "wait"
	StepTypeSubflow   StepType = "subflow"
)

// ActionFunc function to execute for a step
type ActionFunc func(context.Context, *ExecutionContext) (interface{}, error)

// ConditionFunc function to evaluate condition
type ConditionFunc func(*ExecutionContext) (bool, error)

// RetryPolicy retry configuration
type RetryPolicy struct {
	MaxAttempts int
	Delay       time.Duration
	BackoffRate float64 // Exponential backoff multiplier
}

// Execution represents a workflow execution instance
type Execution struct {
	ID           string
	WorkflowID   string
	Status       WorkflowStatus
	CurrentStep  string
	Input        map[string]interface{}
	Output       map[string]interface{}
	Context      *ExecutionContext
	StepResults  map[string]*StepResult
	StartedAt    time.Time
	CompletedAt  *time.Time
	Error        error
	mu           sync.RWMutex
}

// ExecutionContext context for workflow execution
type ExecutionContext struct {
	WorkflowID   string
	ExecutionID  string
	Variables    map[string]interface{}
	StepResults  map[string]interface{}
	Metadata     map[string]string
	mu           sync.RWMutex
}

// StepResult result of step execution
type StepResult struct {
	StepID      string
	Status      WorkflowStatus
	Output      interface{}
	Error       error
	Attempts    int
	StartedAt   time.Time
	CompletedAt *time.Time
	Duration    time.Duration
}

// WorkflowEngine manages workflow execution
type WorkflowEngine struct {
	workflows  map[string]*Workflow
	executions map[string]*Execution
	mu         sync.RWMutex
}

// NewWorkflowEngine creates a new workflow engine
func NewWorkflowEngine() *WorkflowEngine {
	return &WorkflowEngine{
		workflows:  make(map[string]*Workflow),
		executions: make(map[string]*Execution),
	}
}

// RegisterWorkflow registers a workflow
func (e *WorkflowEngine) RegisterWorkflow(workflow *Workflow) error {
	if workflow.ID == "" {
		workflow.ID = fmt.Sprintf("workflow-%d", time.Now().UnixNano())
	}
	workflow.CreatedAt = time.Now()
	workflow.UpdatedAt = time.Now()

	e.mu.Lock()
	e.workflows[workflow.ID] = workflow
	e.mu.Unlock()

	return nil
}

// GetWorkflow gets a workflow by ID
func (e *WorkflowEngine) GetWorkflow(workflowID string) (*Workflow, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	workflow, exists := e.workflows[workflowID]
	if !exists {
		return nil, fmt.Errorf("workflow not found: %s", workflowID)
	}

	return workflow, nil
}

// StartExecution starts a workflow execution
func (e *WorkflowEngine) StartExecution(ctx context.Context, workflowID string, input map[string]interface{}) (*Execution, error) {
	workflow, err := e.GetWorkflow(workflowID)
	if err != nil {
		return nil, err
	}

	execution := &Execution{
		ID:          fmt.Sprintf("exec-%d", time.Now().UnixNano()),
		WorkflowID:  workflowID,
		Status:      StatusRunning,
		Input:       input,
		Output:      make(map[string]interface{}),
		StepResults: make(map[string]*StepResult),
		StartedAt:   time.Now(),
		Context: &ExecutionContext{
			WorkflowID:  workflowID,
			ExecutionID: fmt.Sprintf("exec-%d", time.Now().UnixNano()),
			Variables:   input,
			StepResults: make(map[string]interface{}),
			Metadata:    make(map[string]string),
		},
	}

	e.mu.Lock()
	e.executions[execution.ID] = execution
	e.mu.Unlock()

	// Execute workflow in background
	go e.executeWorkflow(ctx, workflow, execution)

	return execution, nil
}

// executeWorkflow executes a workflow
func (e *WorkflowEngine) executeWorkflow(ctx context.Context, workflow *Workflow, execution *Execution) {
	defer func() {
		if r := recover(); r != nil {
			execution.mu.Lock()
			execution.Status = StatusFailed
			execution.Error = fmt.Errorf("panic: %v", r)
			now := time.Now()
			execution.CompletedAt = &now
			execution.mu.Unlock()
		}
	}()

	// Execute steps in order
	for i, step := range workflow.Steps {
		select {
		case <-ctx.Done():
			execution.mu.Lock()
			execution.Status = StatusCancelled
			execution.Error = ctx.Err()
			now := time.Now()
			execution.CompletedAt = &now
			execution.mu.Unlock()
			return
		default:
		}

		execution.mu.Lock()
		execution.CurrentStep = step.ID
		execution.mu.Unlock()

		result := e.executeStep(ctx, &step, execution.Context)

		execution.mu.Lock()
		execution.StepResults[step.ID] = result
		execution.mu.Unlock()

		if result.Error != nil {
			// Check if there are OnFailure steps
			if len(step.OnFailure) > 0 {
				// Continue to failure handler steps
				continue
			}

			execution.mu.Lock()
			execution.Status = StatusFailed
			execution.Error = result.Error
			now := time.Now()
			execution.CompletedAt = &now
			execution.mu.Unlock()
			return
		}

		// Check if this is the last step
		if i == len(workflow.Steps)-1 {
			execution.mu.Lock()
			execution.Status = StatusCompleted
			now := time.Now()
			execution.CompletedAt = &now
			execution.mu.Unlock()
			return
		}
	}
}

// executeStep executes a single step
func (e *WorkflowEngine) executeStep(ctx context.Context, step *Step, execCtx *ExecutionContext) *StepResult {
	result := &StepResult{
		StepID:    step.ID,
		Status:    StatusRunning,
		StartedAt: time.Now(),
	}

	// Apply timeout if configured
	if step.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, step.Timeout)
		defer cancel()
	}

	// Execute with retry policy
	maxAttempts := 1
	if step.RetryPolicy != nil {
		maxAttempts = step.RetryPolicy.MaxAttempts
	}

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		result.Attempts = attempt

		// Execute based on step type
		var output interface{}
		var err error

		switch step.Type {
		case StepTypeTask:
			if step.Action != nil {
				output, err = step.Action(ctx, execCtx)
			}

		case StepTypeCondition:
			if step.Condition != nil {
				condResult, condErr := step.Condition(execCtx)
				if condErr != nil {
					err = condErr
				} else {
					output = condResult
				}
			}

		case StepTypeWait:
			if duration, ok := step.Parameters["duration"].(time.Duration); ok {
				time.Sleep(duration)
			}

		case StepTypeSubflow:
			// Execute subflow (simplified)
			output = map[string]interface{}{"subflow": "completed"}

		default:
			err = fmt.Errorf("unknown step type: %s", step.Type)
		}

		if err == nil {
			result.Status = StatusCompleted
			result.Output = output
			now := time.Now()
			result.CompletedAt = &now
			result.Duration = time.Since(result.StartedAt)

			// Store result in context
			execCtx.mu.Lock()
			execCtx.StepResults[step.ID] = output
			execCtx.mu.Unlock()

			return result
		}

		lastErr = err

		// Retry with backoff
		if attempt < maxAttempts && step.RetryPolicy != nil {
			delay := step.RetryPolicy.Delay
			if step.RetryPolicy.BackoffRate > 0 {
				for i := 1; i < attempt; i++ {
					delay = time.Duration(float64(delay) * step.RetryPolicy.BackoffRate)
				}
			}
			time.Sleep(delay)
		}
	}

	result.Status = StatusFailed
	result.Error = lastErr
	now := time.Now()
	result.CompletedAt = &now
	result.Duration = time.Since(result.StartedAt)

	return result
}

// GetExecution gets an execution by ID
func (e *WorkflowEngine) GetExecution(executionID string) (*Execution, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	execution, exists := e.executions[executionID]
	if !exists {
		return nil, fmt.Errorf("execution not found: %s", executionID)
	}

	return execution, nil
}

// CancelExecution cancels a workflow execution
func (e *WorkflowEngine) CancelExecution(executionID string) error {
	execution, err := e.GetExecution(executionID)
	if err != nil {
		return err
	}

	execution.mu.Lock()
	defer execution.mu.Unlock()

	if execution.Status != StatusRunning {
		return fmt.Errorf("execution not running: %s", executionID)
	}

	execution.Status = StatusCancelled
	now := time.Now()
	execution.CompletedAt = &now

	return nil
}

// ListExecutions lists all executions for a workflow
func (e *WorkflowEngine) ListExecutions(workflowID string) []*Execution {
	e.mu.RLock()
	defer e.mu.RUnlock()

	executions := make([]*Execution, 0)
	for _, exec := range e.executions {
		if exec.WorkflowID == workflowID {
			executions = append(executions, exec)
		}
	}

	return executions
}

// ListWorkflows lists all workflows
func (e *WorkflowEngine) ListWorkflows() []*Workflow {
	e.mu.RLock()
	defer e.mu.RUnlock()

	workflows := make([]*Workflow, 0, len(e.workflows))
	for _, workflow := range e.workflows {
		workflows = append(workflows, workflow)
	}

	return workflows
}

// DeleteWorkflow deletes a workflow
func (e *WorkflowEngine) DeleteWorkflow(workflowID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.workflows[workflowID]; !exists {
		return fmt.Errorf("workflow not found: %s", workflowID)
	}

	delete(e.workflows, workflowID)
	return nil
}

// Set sets a variable in execution context
func (ctx *ExecutionContext) Set(key string, value interface{}) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.Variables[key] = value
}

// Get gets a variable from execution context
func (ctx *ExecutionContext) Get(key string) (interface{}, bool) {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	value, exists := ctx.Variables[key]
	return value, exists
}

// GetStepResult gets a step result
func (ctx *ExecutionContext) GetStepResult(stepID string) (interface{}, bool) {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	result, exists := ctx.StepResults[stepID]
	return result, exists
}
