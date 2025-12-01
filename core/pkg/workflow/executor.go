package workflow

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ParallelExecutor executes steps in parallel
type ParallelExecutor struct {
	maxWorkers int
}

// NewParallelExecutor creates a new parallel executor
func NewParallelExecutor(maxWorkers int) *ParallelExecutor {
	if maxWorkers <= 0 {
		maxWorkers = 5 // default
	}
	return &ParallelExecutor{
		maxWorkers: maxWorkers,
	}
}

// Execute executes steps in parallel
func (p *ParallelExecutor) Execute(ctx context.Context, steps []Step, execCtx *ExecutionContext) map[string]*StepResult {
	results := make(map[string]*StepResult)
	resultsMu := sync.Mutex{}

	// Create worker pool
	stepsChan := make(chan Step, len(steps))
	resultsChan := make(chan struct {
		id     string
		result *StepResult
	}, len(steps))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < p.maxWorkers && i < len(steps); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for step := range stepsChan {
				result := executeStepWithContext(ctx, step, execCtx)
				resultsChan <- struct {
					id     string
					result *StepResult
				}{
					id:     step.ID,
					result: result,
				}
			}
		}()
	}

	// Send steps to workers
	go func() {
		for _, step := range steps {
			stepsChan <- step
		}
		close(stepsChan)
	}()

	// Collect results
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	for res := range resultsChan {
		resultsMu.Lock()
		results[res.id] = res.result
		resultsMu.Unlock()
	}

	return results
}

// executeStepWithContext executes a step with context
func executeStepWithContext(ctx context.Context, step Step, execCtx *ExecutionContext) *StepResult {
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

		var output interface{}
		var err error

		if step.Action != nil {
			output, err = step.Action(ctx, execCtx)
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

// LoopExecutor executes steps in a loop
type LoopExecutor struct{}

// NewLoopExecutor creates a new loop executor
func NewLoopExecutor() *LoopExecutor {
	return &LoopExecutor{}
}

// ForEach executes a step for each item
func (l *LoopExecutor) ForEach(ctx context.Context, step Step, items []interface{}, execCtx *ExecutionContext) []*StepResult {
	results := make([]*StepResult, 0, len(items))

	for i, item := range items {
		// Set current item in context
		execCtx.Set(fmt.Sprintf("item_%d", i), item)
		execCtx.Set("current_item", item)
		execCtx.Set("current_index", i)

		result := executeStepWithContext(ctx, step, execCtx)
		results = append(results, result)

		// Break on error if no error handling
		if result.Error != nil && len(step.OnFailure) == 0 {
			break
		}
	}

	return results
}

// While executes a step while condition is true
func (l *LoopExecutor) While(ctx context.Context, step Step, condition ConditionFunc, execCtx *ExecutionContext, maxIterations int) []*StepResult {
	results := make([]*StepResult, 0)
	iteration := 0

	for {
		if maxIterations > 0 && iteration >= maxIterations {
			break
		}

		// Check condition
		shouldContinue, err := condition(execCtx)
		if err != nil || !shouldContinue {
			break
		}

		execCtx.Set("iteration", iteration)

		result := executeStepWithContext(ctx, step, execCtx)
		results = append(results, result)

		// Break on error
		if result.Error != nil {
			break
		}

		iteration++
	}

	return results
}

// ConditionalExecutor executes steps based on conditions
type ConditionalExecutor struct{}

// NewConditionalExecutor creates a new conditional executor
func NewConditionalExecutor() *ConditionalExecutor {
	return &ConditionalExecutor{}
}

// IfThenElse executes if-then-else logic
func (c *ConditionalExecutor) IfThenElse(
	ctx context.Context,
	condition ConditionFunc,
	thenStep Step,
	elseStep *Step,
	execCtx *ExecutionContext,
) *StepResult {
	shouldExecute, err := condition(execCtx)
	if err != nil {
		return &StepResult{
			StepID:    "condition",
			Status:    StatusFailed,
			Error:     err,
			StartedAt: time.Now(),
		}
	}

	if shouldExecute {
		return executeStepWithContext(ctx, thenStep, execCtx)
	}

	if elseStep != nil {
		return executeStepWithContext(ctx, *elseStep, execCtx)
	}

	return &StepResult{
		StepID:    "condition",
		Status:    StatusCompleted,
		Output:    "condition not met, no else branch",
		StartedAt: time.Now(),
	}
}

// Switch executes steps based on switch cases
func (c *ConditionalExecutor) Switch(
	ctx context.Context,
	value interface{},
	cases map[interface{}]Step,
	defaultStep *Step,
	execCtx *ExecutionContext,
) *StepResult {
	// Check if value matches any case
	if step, exists := cases[value]; exists {
		return executeStepWithContext(ctx, step, execCtx)
	}

	// Execute default case
	if defaultStep != nil {
		return executeStepWithContext(ctx, *defaultStep, execCtx)
	}

	return &StepResult{
		StepID:    "switch",
		Status:    StatusCompleted,
		Output:    fmt.Sprintf("no matching case for value: %v", value),
		StartedAt: time.Now(),
	}
}
