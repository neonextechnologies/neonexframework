package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// OpenAIProvider provider for OpenAI API
type OpenAIProvider struct {
	apiKey  string
	baseURL string
	client  *http.Client
	metrics map[string]*ModelMetrics
	mu      sync.RWMutex
}

// OpenAIConfig configuration for OpenAI
type OpenAIConfig struct {
	APIKey  string
	BaseURL string // Optional, defaults to https://api.openai.com/v1
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(config *OpenAIConfig) *OpenAIProvider {
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	return &OpenAIProvider{
		apiKey:  config.APIKey,
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
		metrics: make(map[string]*ModelMetrics),
	}
}

// LoadModel loads an OpenAI model
func (p *OpenAIProvider) LoadModel(config *ModelConfig) (*Model, error) {
	model := &Model{
		ID:       config.ID,
		Name:     config.Name,
		Version:  config.Version,
		Type:     config.Type,
		Status:   ModelStatusReady,
		Endpoint: config.Endpoint,
		Provider: "openai",
		Config:   config.Config,
		Metadata: config.Metadata,
		LoadedAt: time.Now(),
	}

	// Initialize metrics
	p.mu.Lock()
	p.metrics[config.ID] = &ModelMetrics{
		ModelID: config.ID,
	}
	p.mu.Unlock()

	return model, nil
}

// UnloadModel unloads a model (no-op for OpenAI)
func (p *OpenAIProvider) UnloadModel(modelID string) error {
	p.mu.Lock()
	delete(p.metrics, modelID)
	p.mu.Unlock()
	return nil
}

// Predict performs inference using OpenAI API
func (p *OpenAIProvider) Predict(ctx context.Context, modelID string, input *InferenceInput) (*InferenceOutput, error) {
	startTime := time.Now()

	// Build request based on model type
	var result interface{}
	var err error

	switch input.Parameters["type"] {
	case "chat":
		result, err = p.chatCompletion(ctx, modelID, input)
	case "completion":
		result, err = p.completion(ctx, modelID, input)
	case "embedding":
		result, err = p.embedding(ctx, modelID, input)
	default:
		result, err = p.chatCompletion(ctx, modelID, input) // Default to chat
	}

	if err != nil {
		p.recordMetrics(modelID, time.Since(startTime), true)
		return nil, err
	}

	p.recordMetrics(modelID, time.Since(startTime), false)

	return &InferenceOutput{
		ModelID:   modelID,
		Result:    result,
		Metadata:  map[string]interface{}{},
		Latency:   time.Since(startTime),
		Timestamp: time.Now(),
	}, nil
}

// chatCompletion performs chat completion
func (p *OpenAIProvider) chatCompletion(ctx context.Context, modelID string, input *InferenceInput) (interface{}, error) {
	messages := []map[string]string{
		{"role": "user", "content": fmt.Sprintf("%v", input.Data)},
	}

	// Add system message if provided
	if systemMsg, ok := input.Parameters["system"]; ok {
		messages = append([]map[string]string{
			{"role": "system", "content": fmt.Sprintf("%v", systemMsg)},
		}, messages...)
	}

	requestBody := map[string]interface{}{
		"model":    modelID,
		"messages": messages,
	}

	// Add optional parameters
	if temp, ok := input.Parameters["temperature"]; ok {
		requestBody["temperature"] = temp
	}
	if maxTokens, ok := input.Parameters["max_tokens"]; ok {
		requestBody["max_tokens"] = maxTokens
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// completion performs text completion
func (p *OpenAIProvider) completion(ctx context.Context, modelID string, input *InferenceInput) (interface{}, error) {
	requestBody := map[string]interface{}{
		"model":  modelID,
		"prompt": input.Data,
	}

	// Add optional parameters
	if temp, ok := input.Parameters["temperature"]; ok {
		requestBody["temperature"] = temp
	}
	if maxTokens, ok := input.Parameters["max_tokens"]; ok {
		requestBody["max_tokens"] = maxTokens
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// embedding generates embeddings
func (p *OpenAIProvider) embedding(ctx context.Context, modelID string, input *InferenceInput) (interface{}, error) {
	requestBody := map[string]interface{}{
		"model": modelID,
		"input": input.Data,
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// GetMetrics returns model metrics
func (p *OpenAIProvider) GetMetrics(modelID string) *ModelMetrics {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.metrics[modelID]
}

// recordMetrics records inference metrics
func (p *OpenAIProvider) recordMetrics(modelID string, latency time.Duration, isError bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	metrics := p.metrics[modelID]
	if metrics == nil {
		metrics = &ModelMetrics{ModelID: modelID}
		p.metrics[modelID] = metrics
	}

	metrics.RequestCount++
	metrics.TotalLatency += latency
	metrics.AvgLatency = metrics.TotalLatency / time.Duration(metrics.RequestCount)
	metrics.LastRequestAt = time.Now()

	if isError {
		metrics.ErrorCount++
	}
}
