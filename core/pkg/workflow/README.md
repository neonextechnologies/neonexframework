# Workflow Engine Package

NeonexCore Workflow Engine provides a powerful and flexible system for orchestrating complex workflows with support for conditional logic, loops, parallel execution, state persistence, and comprehensive error handling.

## Features

- **Workflow Definition**: Define workflows using Go code, YAML, or JSON
- **Step Types**: Task, Condition, Parallel, Loop, Wait, Subflow
- **Conditional Logic**: If-then-else and switch statements
- **Loops**: ForEach and While loops
- **Parallel Execution**: Execute multiple steps concurrently
- **Retry Logic**: Configurable retry policies with exponential backoff
- **State Persistence**: Save and resume workflow execution
- **Event Logging**: Track workflow execution history
- **Timeout Support**: Per-step timeout configuration
- **Error Handling**: Custom error handling with OnSuccess/OnFailure paths

## Installation

```go
import "github.com/neonexcore/pkg/workflow"
```

## Quick Start

### Basic Workflow

```go
// Create workflow engine
engine := workflow.NewWorkflowEngine()

// Build workflow using fluent API
wf := workflow.NewWorkflowBuilder("order-processing").
    Description("Process customer orders").
    Version("1.0.0").
    AddStep("validate", "Validate Order").
        Action(func(ctx context.Context, execCtx *workflow.ExecutionContext) (interface{}, error) {
            // Validate order
            return map[string]interface{}{"valid": true}, nil
        }).
        Retry(3, 1*time.Second, 2.0).
        Timeout(30 * time.Second).
    Then("process", "Process Payment").
        Action(func(ctx context.Context, execCtx *workflow.ExecutionContext) (interface{}, error) {
            // Process payment
            return map[string]interface{}{"payment_id": "PAY123"}, nil
        }).
    Then("notify", "Send Notification").
        Action(func(ctx context.Context, execCtx *workflow.ExecutionContext) (interface{}, error) {
            // Send notification
            return map[string]interface{}{"sent": true}, nil
        }).
    End().
    Build()

// Register workflow
engine.RegisterWorkflow(wf)

// Execute workflow
execution, err := engine.StartExecution(context.Background(), wf.ID, map[string]interface{}{
    "order_id": "ORD123",
    "amount":   99.99,
})
```

### YAML Workflow Definition

```yaml
name: order-processing
description: Process customer orders
version: 1.0.0
config:
  timeout: 300s

steps:
  - id: validate
    name: Validate Order
    type: task
    action_type: validate_order
    timeout: 30s
    retry:
      max_attempts: 3
      delay: 1s
      backoff_rate: 2.0
    on_success:
      - process
    on_failure:
      - notify_error

  - id: process
    name: Process Payment
    type: task
    action_type: process_payment
    timeout: 60s
    on_success:
      - notify
    on_failure:
      - refund

  - id: notify
    name: Send Notification
    type: task
    action_type: send_notification
```

Load and execute YAML workflow:

```go
// Define action registry
actionRegistry := map[string]workflow.ActionFunc{
    "validate_order": func(ctx context.Context, execCtx *workflow.ExecutionContext) (interface{}, error) {
        // Validation logic
        return map[string]interface{}{"valid": true}, nil
    },
    "process_payment": func(ctx context.Context, execCtx *workflow.ExecutionContext) (interface{}, error) {
        // Payment processing
        return map[string]interface{}{"payment_id": "PAY123"}, nil
    },
    "send_notification": func(ctx context.Context, execCtx *workflow.ExecutionContext) (interface{}, error) {
        // Send notification
        return nil, nil
    },
}

// Load workflow from YAML
yamlData := []byte(yamlContent)
wf, err := workflow.FromYAML(yamlData, actionRegistry)
if err != nil {
    log.Fatal(err)
}

// Register and execute
engine.RegisterWorkflow(wf)
execution, err := engine.StartExecution(context.Background(), wf.ID, map[string]interface{}{
    "order_id": "ORD123",
})
```

## Advanced Features

### Conditional Execution

```go
condExecutor := workflow.NewConditionalExecutor()

// If-Then-Else
condition := func(execCtx *workflow.ExecutionContext) (bool, error) {
    amount, _ := execCtx.Get("amount")
    return amount.(float64) > 100, nil
}

thenStep := workflow.Step{
    ID:   "premium",
    Name: "Premium Processing",
    Action: func(ctx context.Context, execCtx *workflow.ExecutionContext) (interface{}, error) {
        return "Premium service", nil
    },
}

elseStep := workflow.Step{
    ID:   "standard",
    Name: "Standard Processing",
    Action: func(ctx context.Context, execCtx *workflow.ExecutionContext) (interface{}, error) {
        return "Standard service", nil
    },
}

result := condExecutor.IfThenElse(ctx, condition, thenStep, &elseStep, execCtx)

// Switch Statement
value := "premium"
cases := map[interface{}]workflow.Step{
    "premium": premiumStep,
    "standard": standardStep,
    "basic": basicStep,
}

result := condExecutor.Switch(ctx, value, cases, &defaultStep, execCtx)
```

### Loop Execution

```go
loopExecutor := workflow.NewLoopExecutor()

// ForEach Loop
items := []interface{}{"item1", "item2", "item3"}
step := workflow.Step{
    ID: "process_item",
    Action: func(ctx context.Context, execCtx *workflow.ExecutionContext) (interface{}, error) {
        item, _ := execCtx.Get("current_item")
        // Process item
        return item, nil
    },
}

results := loopExecutor.ForEach(ctx, step, items, execCtx)

// While Loop
condition := func(execCtx *workflow.ExecutionContext) (bool, error) {
    iteration, _ := execCtx.Get("iteration")
    return iteration.(int) < 5, nil
}

results := loopExecutor.While(ctx, step, condition, execCtx, 10) // max 10 iterations
```

### Parallel Execution

```go
parallelExecutor := workflow.NewParallelExecutor(5) // 5 workers

steps := []workflow.Step{
    {
        ID:   "task1",
        Name: "Task 1",
        Action: func(ctx context.Context, execCtx *workflow.ExecutionContext) (interface{}, error) {
            return "result1", nil
        },
    },
    {
        ID:   "task2",
        Name: "Task 2",
        Action: func(ctx context.Context, execCtx *workflow.ExecutionContext) (interface{}, error) {
            return "result2", nil
        },
    },
}

results := parallelExecutor.Execute(ctx, steps, execCtx)
```

### State Persistence

```go
// Create state store with database
stateStore, err := workflow.NewStateStore(db)
if err != nil {
    log.Fatal(err)
}

// Create stateful engine
engine := workflow.NewStatefulWorkflowEngine(stateStore)

// Start execution (state is automatically saved)
execution, err := engine.StartExecution(ctx, workflowID, input)

// Resume paused/failed execution
err = engine.ResumeExecution(ctx, executionID)

// Query execution state
states, err := stateStore.ListStates(workflowID, workflow.StatusRunning, 10)

// Get event logs
events, err := stateStore.GetEvents(executionID, 100)

// Cleanup old states (older than 30 days)
deleted, err := stateStore.CleanupOldStates(30 * 24 * time.Hour)
```

## Workflow Step Types

### Task Step
Execute a single action:
```go
step := workflow.Step{
    Type: workflow.StepTypeTask,
    Action: func(ctx context.Context, execCtx *workflow.ExecutionContext) (interface{}, error) {
        // Your logic here
        return result, nil
    },
}
```

### Condition Step
Evaluate a condition:
```go
step := workflow.Step{
    Type: workflow.StepTypeCondition,
    Condition: func(execCtx *workflow.ExecutionContext) (bool, error) {
        // Return true or false
        return true, nil
    },
}
```

### Wait Step
Wait for a duration:
```go
step := workflow.Step{
    Type: workflow.StepTypeWait,
    Parameters: map[string]interface{}{
        "duration": 5 * time.Second,
    },
}
```

### Subflow Step
Execute another workflow:
```go
step := workflow.Step{
    Type: workflow.StepTypeSubflow,
    Parameters: map[string]interface{}{
        "workflow_id": "sub-workflow-id",
        "input": map[string]interface{}{"key": "value"},
    },
}
```

## Error Handling

### Retry Policy

```go
step := workflow.Step{
    RetryPolicy: &workflow.RetryPolicy{
        MaxAttempts: 3,
        Delay:       1 * time.Second,
        BackoffRate: 2.0, // 1s, 2s, 4s delays
    },
}
```

### OnSuccess/OnFailure Paths

```go
step := workflow.Step{
    OnSuccess: []string{"next_step"},
    OnFailure: []string{"error_handler", "notify"},
}
```

## Monitoring and Logging

### Get Execution Status

```go
execution, err := engine.GetExecution(executionID)
fmt.Printf("Status: %s\n", execution.Status)
fmt.Printf("Current Step: %s\n", execution.CurrentStep)

// Get step results
for stepID, result := range execution.StepResults {
    fmt.Printf("Step %s: %+v\n", stepID, result)
}
```

### Event Logging

```go
// Log custom events
stateStore.LogEvent(executionID, stepID, "custom", "Custom message", map[string]interface{}{
    "key": "value",
})

// Query events
events, err := stateStore.GetEvents(executionID, 100)
for _, event := range events {
    fmt.Printf("[%s] %s: %s\n", event.Timestamp, event.EventType, event.Message)
}
```

## Best Practices

1. **Use Timeouts**: Always set appropriate timeouts for steps
2. **Implement Retry Logic**: Use retry policies for network operations
3. **Handle Errors**: Define OnFailure paths for critical steps
4. **State Persistence**: Use StatefulWorkflowEngine for long-running workflows
5. **Parallel Execution**: Use parallel executor for independent tasks
6. **Monitor Execution**: Log events and track execution progress
7. **Version Workflows**: Use version field for workflow management
8. **Clean Old States**: Regularly cleanup completed executions

## Complete Example

See `examples/workflow_example.go` for comprehensive examples including:
- Basic workflow execution
- YAML workflow loading
- Conditional logic (if-then-else, switch)
- Loops (foreach, while)
- Parallel execution
- State persistence and resumption
- Error handling and retry
- Event logging and monitoring

## Architecture

The workflow engine consists of:
- **WorkflowEngine**: Main orchestration engine
- **Workflow**: Workflow definition with steps
- **Execution**: Runtime execution instance
- **ExecutionContext**: Shared context for step execution
- **StateStore**: Persistent state storage
- **Executors**: Specialized executors (parallel, loop, conditional)
- **DSL Parser**: YAML/JSON workflow parser

## Performance

- Lightweight execution: ~1ms per step overhead
- Parallel execution: Configurable worker pool
- State persistence: Async with periodic saves
- Event logging: Non-blocking with batching
- Memory efficient: Stream-based processing

## Thread Safety

All components are thread-safe:
- Workflow registration uses RWMutex
- Execution state uses RWMutex
- Context variables use RWMutex
- State store operations are synchronized
