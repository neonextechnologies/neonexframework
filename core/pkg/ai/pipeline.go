package ai

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Pipeline represents an ML inference pipeline
type Pipeline struct {
	ID          string
	Name        string
	Description string
	Steps       []PipelineStep
	Config      map[string]interface{}
	CreatedAt   time.Time
}

// PipelineStep represents a step in the pipeline
type PipelineStep struct {
	Name       string
	Type       StepType
	ModelID    string
	Transform  TransformFunc
	Parameters map[string]interface{}
}

// StepType represents the type of pipeline step
type StepType string

const (
	StepTypePreprocess  StepType = "preprocess"
	StepTypeModel       StepType = "model"
	StepTypePostprocess StepType = "postprocess"
	StepTypeTransform   StepType = "transform"
)

// TransformFunc function for transforming data
type TransformFunc func(context.Context, interface{}) (interface{}, error)

// PipelineManager manages ML pipelines
type PipelineManager struct {
	pipelines    map[string]*Pipeline
	modelManager *ModelManager
	mu           sync.RWMutex
}

// PipelineResult result of pipeline execution
type PipelineResult struct {
	PipelineID  string
	Input       interface{}
	Output      interface{}
	StepResults []StepResult
	Latency     time.Duration
	Timestamp   time.Time
}

// StepResult result of a pipeline step
type StepResult struct {
	StepName  string
	Input     interface{}
	Output    interface{}
	Latency   time.Duration
	Error     error
}

// NewPipelineManager creates a new pipeline manager
func NewPipelineManager(modelManager *ModelManager) *PipelineManager {
	return &PipelineManager{
		pipelines:    make(map[string]*Pipeline),
		modelManager: modelManager,
	}
}

// CreatePipeline creates a new pipeline
func (pm *PipelineManager) CreatePipeline(pipeline *Pipeline) error {
	if pipeline.ID == "" {
		pipeline.ID = fmt.Sprintf("pipeline-%d", time.Now().UnixNano())
	}
	pipeline.CreatedAt = time.Now()

	pm.mu.Lock()
	pm.pipelines[pipeline.ID] = pipeline
	pm.mu.Unlock()

	return nil
}

// GetPipeline gets a pipeline by ID
func (pm *PipelineManager) GetPipeline(pipelineID string) (*Pipeline, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	pipeline, exists := pm.pipelines[pipelineID]
	if !exists {
		return nil, fmt.Errorf("pipeline not found: %s", pipelineID)
	}

	return pipeline, nil
}

// Execute executes a pipeline
func (pm *PipelineManager) Execute(ctx context.Context, pipelineID string, input interface{}) (*PipelineResult, error) {
	startTime := time.Now()

	pipeline, err := pm.GetPipeline(pipelineID)
	if err != nil {
		return nil, err
	}

	result := &PipelineResult{
		PipelineID:  pipelineID,
		Input:       input,
		StepResults: make([]StepResult, 0),
	}

	currentData := input

	// Execute each step
	for _, step := range pipeline.Steps {
		stepStart := time.Now()
		stepResult := StepResult{
			StepName: step.Name,
			Input:    currentData,
		}

		var stepOutput interface{}
		var stepErr error

		switch step.Type {
		case StepTypePreprocess, StepTypePostprocess, StepTypeTransform:
			if step.Transform != nil {
				stepOutput, stepErr = step.Transform(ctx, currentData)
			} else {
				stepOutput = currentData // Pass through
			}

		case StepTypeModel:
			if step.ModelID == "" {
				stepErr = fmt.Errorf("model_id required for model step")
			} else {
				inferenceInput := &InferenceInput{
					ModelID:    step.ModelID,
					Data:       currentData,
					Parameters: step.Parameters,
				}
				inferenceOutput, err := pm.modelManager.Predict(ctx, inferenceInput)
				if err != nil {
					stepErr = err
				} else {
					stepOutput = inferenceOutput.Result
				}
			}

		default:
			stepErr = fmt.Errorf("unknown step type: %s", step.Type)
		}

		stepResult.Output = stepOutput
		stepResult.Error = stepErr
		stepResult.Latency = time.Since(stepStart)

		result.StepResults = append(result.StepResults, stepResult)

		if stepErr != nil {
			result.Latency = time.Since(startTime)
			result.Timestamp = time.Now()
			return result, fmt.Errorf("step %s failed: %w", step.Name, stepErr)
		}

		currentData = stepOutput
	}

	result.Output = currentData
	result.Latency = time.Since(startTime)
	result.Timestamp = time.Now()

	return result, nil
}

// ListPipelines lists all pipelines
func (pm *PipelineManager) ListPipelines() []*Pipeline {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	pipelines := make([]*Pipeline, 0, len(pm.pipelines))
	for _, pipeline := range pm.pipelines {
		pipelines = append(pipelines, pipeline)
	}

	return pipelines
}

// DeletePipeline deletes a pipeline
func (pm *PipelineManager) DeletePipeline(pipelineID string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.pipelines[pipelineID]; !exists {
		return fmt.Errorf("pipeline not found: %s", pipelineID)
	}

	delete(pm.pipelines, pipelineID)
	return nil
}

// Common transform functions

// TextPreprocessor preprocesses text
func TextPreprocessor(ctx context.Context, input interface{}) (interface{}, error) {
	text, ok := input.(string)
	if !ok {
		return nil, fmt.Errorf("expected string input")
	}

	// Simple preprocessing: trim, lowercase
	// Add more sophisticated preprocessing as needed
	return text, nil
}

// JSONExtractor extracts field from JSON response
func JSONExtractor(fieldPath string) TransformFunc {
	return func(ctx context.Context, input interface{}) (interface{}, error) {
		data, ok := input.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("expected map input")
		}

		// Simple field extraction
		// Add nested path support as needed
		if value, exists := data[fieldPath]; exists {
			return value, nil
		}

		return nil, fmt.Errorf("field not found: %s", fieldPath)
	}
}

// BatchProcessor processes data in batches
func BatchProcessor(batchSize int, processFunc TransformFunc) TransformFunc {
	return func(ctx context.Context, input interface{}) (interface{}, error) {
		items, ok := input.([]interface{})
		if !ok {
			return nil, fmt.Errorf("expected slice input")
		}

		results := make([]interface{}, 0, len(items))

		for i := 0; i < len(items); i += batchSize {
			end := i + batchSize
			if end > len(items) {
				end = len(items)
			}

			batch := items[i:end]
			for _, item := range batch {
				result, err := processFunc(ctx, item)
				if err != nil {
					return nil, err
				}
				results = append(results, result)
			}
		}

		return results, nil
	}
}
