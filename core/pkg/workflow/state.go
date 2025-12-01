package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"
)

// StateStore stores workflow execution state
type StateStore struct {
	db *gorm.DB
	mu sync.RWMutex
}

// WorkflowState persisted workflow state
type WorkflowState struct {
	ID           string                 `gorm:"primaryKey"`
	WorkflowID   string                 `gorm:"index"`
	ExecutionID  string                 `gorm:"uniqueIndex"`
	Status       WorkflowStatus         `gorm:"index"`
	CurrentStep  string                 `gorm:"index"`
	Input        string                 `gorm:"type:jsonb"` // JSON serialized
	Output       string                 `gorm:"type:jsonb"` // JSON serialized
	Variables    string                 `gorm:"type:jsonb"` // JSON serialized
	StepResults  string                 `gorm:"type:jsonb"` // JSON serialized
	Error        string                 `gorm:"type:text"`
	StartedAt    time.Time              `gorm:"index"`
	CompletedAt  *time.Time             `gorm:"index"`
	UpdatedAt    time.Time              `gorm:"autoUpdateTime"`
	Metadata     map[string]interface{} `gorm:"-"` // Not stored in DB
}

// EventLog workflow event log
type EventLog struct {
	ID          uint           `gorm:"primaryKey"`
	ExecutionID string         `gorm:"index"`
	StepID      string         `gorm:"index"`
	EventType   string         `gorm:"index"` // started, completed, failed, retried
	Message     string         `gorm:"type:text"`
	Data        string         `gorm:"type:jsonb"`
	Timestamp   time.Time      `gorm:"index"`
}

// NewStateStore creates a new state store
func NewStateStore(db *gorm.DB) (*StateStore, error) {
	// Auto-migrate tables
	if err := db.AutoMigrate(&WorkflowState{}, &EventLog{}); err != nil {
		return nil, fmt.Errorf("failed to migrate tables: %w", err)
	}

	return &StateStore{
		db: db,
	}, nil
}

// SaveState saves workflow execution state
func (s *StateStore) SaveState(execution *Execution) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	execution.mu.RLock()
	defer execution.mu.RUnlock()

	state := &WorkflowState{
		ID:          execution.ID,
		WorkflowID:  execution.WorkflowID,
		ExecutionID: execution.ID,
		Status:      execution.Status,
		CurrentStep: execution.CurrentStep,
		StartedAt:   execution.StartedAt,
		CompletedAt: execution.CompletedAt,
	}

	if execution.Error != nil {
		state.Error = execution.Error.Error()
	}

	// Serialize complex fields
	if inputJSON, err := json.Marshal(execution.Input); err == nil {
		state.Input = string(inputJSON)
	}

	if outputJSON, err := json.Marshal(execution.Output); err == nil {
		state.Output = string(outputJSON)
	}

	if execution.Context != nil {
		execution.Context.mu.RLock()
		if variablesJSON, err := json.Marshal(execution.Context.Variables); err == nil {
			state.Variables = string(variablesJSON)
		}
		execution.Context.mu.RUnlock()
	}

	if resultsJSON, err := json.Marshal(execution.StepResults); err == nil {
		state.StepResults = string(resultsJSON)
	}

	return s.db.Save(state).Error
}

// LoadState loads workflow execution state
func (s *StateStore) LoadState(executionID string) (*Execution, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var state WorkflowState
	if err := s.db.Where("execution_id = ?", executionID).First(&state).Error; err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	execution := &Execution{
		ID:          state.ID,
		WorkflowID:  state.WorkflowID,
		Status:      state.Status,
		CurrentStep: state.CurrentStep,
		StartedAt:   state.StartedAt,
		CompletedAt: state.CompletedAt,
		Input:       make(map[string]interface{}),
		Output:      make(map[string]interface{}),
		StepResults: make(map[string]*StepResult),
		Context: &ExecutionContext{
			WorkflowID:  state.WorkflowID,
			ExecutionID: state.ExecutionID,
			Variables:   make(map[string]interface{}),
			StepResults: make(map[string]interface{}),
			Metadata:    make(map[string]string),
		},
	}

	if state.Error != "" {
		execution.Error = fmt.Errorf("%s", state.Error)
	}

	// Deserialize complex fields
	if state.Input != "" {
		json.Unmarshal([]byte(state.Input), &execution.Input)
	}

	if state.Output != "" {
		json.Unmarshal([]byte(state.Output), &execution.Output)
	}

	if state.Variables != "" {
		json.Unmarshal([]byte(state.Variables), &execution.Context.Variables)
	}

	if state.StepResults != "" {
		json.Unmarshal([]byte(state.StepResults), &execution.StepResults)
	}

	return execution, nil
}

// DeleteState deletes workflow execution state
func (s *StateStore) DeleteState(executionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.db.Where("execution_id = ?", executionID).Delete(&WorkflowState{}).Error
}

// ListStates lists all workflow states
func (s *StateStore) ListStates(workflowID string, status WorkflowStatus, limit int) ([]*WorkflowState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var states []*WorkflowState
	query := s.db.Model(&WorkflowState{})

	if workflowID != "" {
		query = query.Where("workflow_id = ?", workflowID)
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Order("started_at DESC").Find(&states).Error; err != nil {
		return nil, err
	}

	return states, nil
}

// LogEvent logs a workflow event
func (s *StateStore) LogEvent(executionID, stepID, eventType, message string, data map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	event := &EventLog{
		ExecutionID: executionID,
		StepID:      stepID,
		EventType:   eventType,
		Message:     message,
		Timestamp:   time.Now(),
	}

	if dataJSON, err := json.Marshal(data); err == nil {
		event.Data = string(dataJSON)
	}

	return s.db.Create(event).Error
}

// GetEvents gets events for an execution
func (s *StateStore) GetEvents(executionID string, limit int) ([]*EventLog, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var events []*EventLog
	query := s.db.Where("execution_id = ?", executionID)

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Order("timestamp DESC").Find(&events).Error; err != nil {
		return nil, err
	}

	return events, nil
}

// CleanupOldStates removes old completed/failed states
func (s *StateStore) CleanupOldStates(olderThan time.Duration) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().Add(-olderThan)

	result := s.db.Where("completed_at < ? AND status IN ?", cutoff, []WorkflowStatus{StatusCompleted, StatusFailed, StatusCancelled}).
		Delete(&WorkflowState{})

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

// StatefulWorkflowEngine workflow engine with state persistence
type StatefulWorkflowEngine struct {
	*WorkflowEngine
	stateStore *StateStore
}

// NewStatefulWorkflowEngine creates a new stateful workflow engine
func NewStatefulWorkflowEngine(stateStore *StateStore) *StatefulWorkflowEngine {
	return &StatefulWorkflowEngine{
		WorkflowEngine: NewWorkflowEngine(),
		stateStore:     stateStore,
	}
}

// StartExecution starts a workflow execution with state persistence
func (e *StatefulWorkflowEngine) StartExecution(ctx context.Context, workflowID string, input map[string]interface{}) (*Execution, error) {
	execution, err := e.WorkflowEngine.StartExecution(ctx, workflowID, input)
	if err != nil {
		return nil, err
	}

	// Save initial state
	if err := e.stateStore.SaveState(execution); err != nil {
		return nil, fmt.Errorf("failed to save initial state: %w", err)
	}

	// Log start event
	e.stateStore.LogEvent(execution.ID, "", "started", "Workflow execution started", nil)

	// Monitor execution and save state periodically
	go e.monitorExecution(ctx, execution)

	return execution, nil
}

// monitorExecution monitors execution and saves state
func (e *StatefulWorkflowEngine) monitorExecution(ctx context.Context, execution *Execution) {
	ticker := time.NewTicker(5 * time.Second) // Save state every 5 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			e.stateStore.SaveState(execution)
			return
		case <-ticker.C:
			execution.mu.RLock()
			status := execution.Status
			execution.mu.RUnlock()

			// Save current state
			e.stateStore.SaveState(execution)

			// Exit if execution is complete
			if status == StatusCompleted || status == StatusFailed || status == StatusCancelled {
				return
			}
		}
	}
}

// ResumeExecution resumes a paused or failed execution
func (e *StatefulWorkflowEngine) ResumeExecution(ctx context.Context, executionID string) error {
	// Load state from store
	execution, err := e.stateStore.LoadState(executionID)
	if err != nil {
		return fmt.Errorf("failed to load execution state: %w", err)
	}

	// Check if execution can be resumed
	if execution.Status != StatusPaused && execution.Status != StatusFailed {
		return fmt.Errorf("execution cannot be resumed: status=%s", execution.Status)
	}

	// Get workflow
	workflow, err := e.GetWorkflow(execution.WorkflowID)
	if err != nil {
		return err
	}

	// Update status to running
	execution.mu.Lock()
	execution.Status = StatusRunning
	execution.mu.Unlock()

	// Save state
	e.stateStore.SaveState(execution)

	// Log resume event
	e.stateStore.LogEvent(execution.ID, "", "resumed", "Workflow execution resumed", nil)

	// Continue execution
	go e.executeWorkflow(ctx, workflow, execution)

	return nil
}
