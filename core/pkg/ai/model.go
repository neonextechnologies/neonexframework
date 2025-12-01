package ai

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ModelType represents the type of AI model
type ModelType string

const (
	ModelTypeTextGeneration    ModelType = "text_generation"
	ModelTypeTextClassification ModelType = "text_classification"
	ModelTypeImageClassification ModelType = "image_classification"
	ModelTypeObjectDetection   ModelType = "object_detection"
	ModelTypeEmbedding         ModelType = "embedding"
	ModelTypeSentiment         ModelType = "sentiment"
	ModelTypeNER               ModelType = "named_entity_recognition"
	ModelTypeTranslation       ModelType = "translation"
	ModelTypeQuestionAnswering ModelType = "question_answering"
	ModelTypeSummarization     ModelType = "summarization"
)

// ModelStatus represents model deployment status
type ModelStatus string

const (
	ModelStatusLoading  ModelStatus = "loading"
	ModelStatusReady    ModelStatus = "ready"
	ModelStatusError    ModelStatus = "error"
	ModelStatusUnloaded ModelStatus = "unloaded"
)

// Model represents an AI/ML model
type Model struct {
	ID          string
	Name        string
	Version     string
	Type        ModelType
	Status      ModelStatus
	Endpoint    string // Local path or API endpoint
	Provider    string // openai, huggingface, local, custom
	Config      map[string]interface{}
	Metadata    map[string]string
	LoadedAt    time.Time
	LastUsedAt  time.Time
	RequestCount int64
	mu          sync.RWMutex
}

// ModelManager manages AI/ML models
type ModelManager struct {
	models    map[string]*Model
	providers map[string]ModelProvider
	cache     *InferenceCache
	mu        sync.RWMutex
}

// ModelProvider interface for different AI providers
type ModelProvider interface {
	LoadModel(config *ModelConfig) (*Model, error)
	UnloadModel(modelID string) error
	Predict(ctx context.Context, modelID string, input *InferenceInput) (*InferenceOutput, error)
	GetMetrics(modelID string) *ModelMetrics
}

// ModelConfig configuration for loading a model
type ModelConfig struct {
	ID       string
	Name     string
	Version  string
	Type     ModelType
	Provider string
	Endpoint string
	APIKey   string
	Config   map[string]interface{}
	Metadata map[string]string
}

// InferenceInput input for model inference
type InferenceInput struct {
	ModelID    string
	Data       interface{} // Text, image bytes, etc.
	Parameters map[string]interface{}
	Metadata   map[string]string
}

// InferenceOutput output from model inference
type InferenceOutput struct {
	ModelID   string
	Result    interface{}
	Metadata  map[string]interface{}
	Latency   time.Duration
	Timestamp time.Time
}

// ModelMetrics metrics for a model
type ModelMetrics struct {
	ModelID       string
	RequestCount  int64
	TotalLatency  time.Duration
	AvgLatency    time.Duration
	ErrorCount    int64
	LastRequestAt time.Time
}

// NewModelManager creates a new model manager
func NewModelManager() *ModelManager {
	return &ModelManager{
		models:    make(map[string]*Model),
		providers: make(map[string]ModelProvider),
		cache:     NewInferenceCache(1000, 1*time.Hour),
	}
}

// RegisterProvider registers an AI provider
func (m *ModelManager) RegisterProvider(name string, provider ModelProvider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.providers[name] = provider
}

// LoadModel loads a model
func (m *ModelManager) LoadModel(config *ModelConfig) (*Model, error) {
	m.mu.Lock()
	
	// Check if model already loaded
	if model, exists := m.models[config.ID]; exists {
		m.mu.Unlock()
		return model, nil
	}
	m.mu.Unlock()

	// Get provider
	provider := m.getProvider(config.Provider)
	if provider == nil {
		return nil, fmt.Errorf("provider not found: %s", config.Provider)
	}

	// Load model using provider
	model, err := provider.LoadModel(config)
	if err != nil {
		return nil, fmt.Errorf("failed to load model: %w", err)
	}

	// Register model
	m.mu.Lock()
	m.models[config.ID] = model
	m.mu.Unlock()

	return model, nil
}

// UnloadModel unloads a model
func (m *ModelManager) UnloadModel(modelID string) error {
	m.mu.Lock()
	model, exists := m.models[modelID]
	if !exists {
		m.mu.Unlock()
		return fmt.Errorf("model not found: %s", modelID)
	}
	
	provider := m.getProvider(model.Provider)
	m.mu.Unlock()

	if provider == nil {
		return fmt.Errorf("provider not found: %s", model.Provider)
	}

	// Unload from provider
	if err := provider.UnloadModel(modelID); err != nil {
		return err
	}

	// Remove from manager
	m.mu.Lock()
	delete(m.models, modelID)
	m.mu.Unlock()

	return nil
}

// Predict performs inference on a model
func (m *ModelManager) Predict(ctx context.Context, input *InferenceInput) (*InferenceOutput, error) {
	// Check cache first
	if cached := m.cache.Get(input); cached != nil {
		return cached, nil
	}

	// Get model
	model := m.getModel(input.ModelID)
	if model == nil {
		return nil, fmt.Errorf("model not found: %s", input.ModelID)
	}

	if model.Status != ModelStatusReady {
		return nil, fmt.Errorf("model not ready: %s (status: %s)", input.ModelID, model.Status)
	}

	// Get provider
	provider := m.getProvider(model.Provider)
	if provider == nil {
		return nil, fmt.Errorf("provider not found: %s", model.Provider)
	}

	// Perform inference
	startTime := time.Now()
	output, err := provider.Predict(ctx, input.ModelID, input)
	if err != nil {
		model.mu.Lock()
		model.mu.Unlock()
		return nil, fmt.Errorf("inference failed: %w", err)
	}

	// Update model stats
	model.mu.Lock()
	model.LastUsedAt = time.Now()
	model.RequestCount++
	model.mu.Unlock()

	output.Latency = time.Since(startTime)
	output.Timestamp = time.Now()

	// Cache result
	m.cache.Set(input, output)

	return output, nil
}

// GetModel gets a model by ID
func (m *ModelManager) GetModel(modelID string) *Model {
	return m.getModel(modelID)
}

func (m *ModelManager) getModel(modelID string) *Model {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.models[modelID]
}

func (m *ModelManager) getProvider(name string) ModelProvider {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.providers[name]
}

// ListModels lists all loaded models
func (m *ModelManager) ListModels() []*Model {
	m.mu.RLock()
	defer m.mu.RUnlock()

	models := make([]*Model, 0, len(m.models))
	for _, model := range m.models {
		models = append(models, model)
	}
	return models
}

// GetMetrics gets metrics for a model
func (m *ModelManager) GetMetrics(modelID string) *ModelMetrics {
	model := m.getModel(modelID)
	if model == nil {
		return nil
	}

	provider := m.getProvider(model.Provider)
	if provider != nil {
		return provider.GetMetrics(modelID)
	}

	// Return basic metrics
	model.mu.RLock()
	defer model.mu.RUnlock()

	return &ModelMetrics{
		ModelID:       model.ID,
		RequestCount:  model.RequestCount,
		LastRequestAt: model.LastUsedAt,
	}
}

// GetAllMetrics gets metrics for all models
func (m *ModelManager) GetAllMetrics() map[string]*ModelMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metrics := make(map[string]*ModelMetrics)
	for id := range m.models {
		if m := m.GetMetrics(id); m != nil {
			metrics[id] = m
		}
	}
	return metrics
}

// WarmUp pre-loads models for faster first inference
func (m *ModelManager) WarmUp(modelIDs []string) error {
	for _, id := range modelIDs {
		model := m.getModel(id)
		if model == nil {
			continue
		}

		// Perform dummy inference to warm up
		ctx := context.Background()
		input := &InferenceInput{
			ModelID: id,
			Data:    "warmup",
		}
		_, _ = m.Predict(ctx, input)
	}
	return nil
}

// Close closes the model manager
func (m *ModelManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Unload all models
	for id, model := range m.models {
		provider := m.getProvider(model.Provider)
		if provider != nil {
			provider.UnloadModel(id)
		}
	}

	m.models = make(map[string]*Model)
	return nil
}
