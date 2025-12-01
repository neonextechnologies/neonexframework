# AI Integration

Integrate AI and Large Language Models (LLMs) into your NeonEx applications. Learn OpenAI integration, embeddings, vector search, and AI-powered features.

## Table of Contents

- [Introduction](#introduction)
- [Quick Start](#quick-start)
- [OpenAI Integration](#openai-integration)
- [Embeddings](#embeddings)
- [Vector Search](#vector-search)
- [AI Features](#ai-features)
- [Caching](#caching)
- [Best Practices](#best-practices)

## Introduction

NeonEx provides comprehensive AI integration with:

- **OpenAI API**: GPT-4, ChatGPT, DALL-E
- **Embeddings**: Text embeddings for semantic search
- **Vector Store**: Store and search embeddings
- **Streaming**: Real-time response streaming
- **Caching**: Response caching for performance
- **Rate Limiting**: API quota management

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "neonex/core/pkg/ai"
)

func main() {
    // Initialize AI provider
    provider := ai.NewOpenAIProvider(&ai.OpenAIConfig{
        APIKey: "sk-...",
        Model:  "gpt-4",
    })
    
    // Create chat completion
    ctx := context.Background()
    
    resp, err := provider.ChatCompletion(ctx, &ai.ChatRequest{
        Messages: []ai.Message{
            {
                Role:    "system",
                Content: "You are a helpful assistant.",
            },
            {
                Role:    "user",
                Content: "What is the capital of France?",
            },
        },
    })
    
    if err != nil {
        panic(err)
    }
    
    fmt.Println(resp.Content)
}
```

## OpenAI Integration

### Configuration

```go
type AIConfig struct {
    Provider string
    APIKey   string
    Model    string
    
    // Optional
    MaxTokens   int
    Temperature float64
    TopP        float64
    Stop        []string
}

config := &AIConfig{
    Provider:    "openai",
    APIKey:      os.Getenv("OPENAI_API_KEY"),
    Model:       "gpt-4",
    MaxTokens:   2000,
    Temperature: 0.7,
    TopP:        1.0,
}

provider := ai.NewProvider(config)
```

### Chat Completions

```go
func ChatCompletion(provider ai.Provider) {
    resp, err := provider.ChatCompletion(ctx, &ai.ChatRequest{
        Messages: []ai.Message{
            {Role: "system", Content: "You are a helpful assistant."},
            {Role: "user", Content: "Explain quantum computing in simple terms."},
        },
        MaxTokens:   500,
        Temperature: 0.7,
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Response: %s\n", resp.Content)
    fmt.Printf("Tokens used: %d\n", resp.Usage.TotalTokens)
}
```

### Streaming Responses

```go
func StreamingChat(provider ai.Provider) {
    stream, err := provider.ChatCompletionStream(ctx, &ai.ChatRequest{
        Messages: []ai.Message{
            {Role: "user", Content: "Write a short story about a robot."},
        },
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    defer stream.Close()
    
    for {
        chunk, err := stream.Recv()
        if err == io.EOF {
            break
        }
        if err != nil {
            log.Fatal(err)
        }
        
        fmt.Print(chunk.Content)
    }
}
```

### Function Calling

```go
func ChatWithFunctions(provider ai.Provider) {
    functions := []ai.Function{
        {
            Name:        "get_weather",
            Description: "Get the current weather for a location",
            Parameters: ai.FunctionParameters{
                Type: "object",
                Properties: map[string]ai.Property{
                    "location": {
                        Type:        "string",
                        Description: "The city and state, e.g. San Francisco, CA",
                    },
                    "unit": {
                        Type: "string",
                        Enum: []string{"celsius", "fahrenheit"},
                    },
                },
                Required: []string{"location"},
            },
        },
    }
    
    resp, err := provider.ChatCompletion(ctx, &ai.ChatRequest{
        Messages: []ai.Message{
            {Role: "user", Content: "What's the weather in Paris?"},
        },
        Functions: functions,
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    if resp.FunctionCall != nil {
        // Execute function
        result := executeFunction(resp.FunctionCall.Name, resp.FunctionCall.Arguments)
        
        // Continue conversation
        resp2, _ := provider.ChatCompletion(ctx, &ai.ChatRequest{
            Messages: []ai.Message{
                {Role: "user", Content: "What's the weather in Paris?"},
                {Role: "assistant", FunctionCall: resp.FunctionCall},
                {Role: "function", Name: "get_weather", Content: result},
            },
        })
        
        fmt.Println(resp2.Content)
    }
}
```

### Image Generation

```go
func GenerateImage(provider ai.Provider) {
    resp, err := provider.ImageGeneration(ctx, &ai.ImageRequest{
        Prompt: "A futuristic city at sunset with flying cars",
        Size:   "1024x1024",
        N:      1,
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    for _, image := range resp.Images {
        fmt.Printf("Image URL: %s\n", image.URL)
    }
}
```

## Embeddings

### Generate Embeddings

```go
type EmbeddingService struct {
    provider ai.Provider
    cache    cache.Cache
}

func (es *EmbeddingService) GetEmbedding(ctx context.Context, text string) ([]float64, error) {
    // Check cache
    cacheKey := fmt.Sprintf("embedding:%s", hash(text))
    if cached, err := es.cache.Get(ctx, cacheKey); err == nil {
        var embedding []float64
        json.Unmarshal(cached, &embedding)
        return embedding, nil
    }
    
    // Generate embedding
    resp, err := es.provider.CreateEmbedding(ctx, &ai.EmbeddingRequest{
        Input: text,
        Model: "text-embedding-ada-002",
    })
    
    if err != nil {
        return nil, err
    }
    
    embedding := resp.Embeddings[0]
    
    // Cache result
    data, _ := json.Marshal(embedding)
    es.cache.Set(ctx, cacheKey, data, 7*24*time.Hour)
    
    return embedding, nil
}
```

### Batch Embeddings

```go
func (es *EmbeddingService) GetBatchEmbeddings(ctx context.Context, texts []string) ([][]float64, error) {
    // Process in batches of 100
    batchSize := 100
    var allEmbeddings [][]float64
    
    for i := 0; i < len(texts); i += batchSize {
        end := i + batchSize
        if end > len(texts) {
            end = len(texts)
        }
        
        batch := texts[i:end]
        
        resp, err := es.provider.CreateEmbedding(ctx, &ai.EmbeddingRequest{
            Input: batch,
            Model: "text-embedding-ada-002",
        })
        
        if err != nil {
            return nil, err
        }
        
        allEmbeddings = append(allEmbeddings, resp.Embeddings...)
        
        // Rate limiting
        time.Sleep(100 * time.Millisecond)
    }
    
    return allEmbeddings, nil
}
```

## Vector Search

### Vector Store

```go
type VectorStore struct {
    db *gorm.DB
}

type Document struct {
    ID        int64
    Content   string
    Embedding pgvector.Vector `gorm:"type:vector(1536)"`
    Metadata  datatypes.JSON
}

func (vs *VectorStore) Store(ctx context.Context, doc *Document) error {
    return vs.db.WithContext(ctx).Create(doc).Error
}

func (vs *VectorStore) Search(ctx context.Context, queryEmbedding []float64, limit int) ([]Document, error) {
    var docs []Document
    
    err := vs.db.WithContext(ctx).
        Raw(`
            SELECT *, embedding <=> ? AS distance
            FROM documents
            ORDER BY distance
            LIMIT ?
        `, pgvector.NewVector(queryEmbedding), limit).
        Scan(&docs).Error
    
    return docs, err
}
```

### Semantic Search

```go
type SemanticSearch struct {
    embeddingService *EmbeddingService
    vectorStore      *VectorStore
}

func (ss *SemanticSearch) Search(ctx context.Context, query string, limit int) ([]Document, error) {
    // Generate query embedding
    queryEmbedding, err := ss.embeddingService.GetEmbedding(ctx, query)
    if err != nil {
        return nil, err
    }
    
    // Search similar documents
    results, err := ss.vectorStore.Search(ctx, queryEmbedding, limit)
    if err != nil {
        return nil, err
    }
    
    return results, nil
}
```

### Hybrid Search

```go
func (ss *SemanticSearch) HybridSearch(ctx context.Context, query string, limit int) ([]Document, error) {
    // Full-text search
    var textResults []Document
    ss.vectorStore.db.
        Where("content ILIKE ?", "%"+query+"%").
        Limit(limit).
        Find(&textResults)
    
    // Semantic search
    queryEmbedding, _ := ss.embeddingService.GetEmbedding(ctx, query)
    semanticResults, _ := ss.vectorStore.Search(ctx, queryEmbedding, limit)
    
    // Combine and deduplicate
    seen := make(map[int64]bool)
    var combined []Document
    
    for _, doc := range textResults {
        if !seen[doc.ID] {
            combined = append(combined, doc)
            seen[doc.ID] = true
        }
    }
    
    for _, doc := range semanticResults {
        if !seen[doc.ID] {
            combined = append(combined, doc)
            seen[doc.ID] = true
        }
    }
    
    return combined, nil
}
```

## AI Features

### Content Generation

```go
type ContentGenerator struct {
    provider ai.Provider
}

func (cg *ContentGenerator) GenerateBlogPost(ctx context.Context, topic string) (string, error) {
    resp, err := cg.provider.ChatCompletion(ctx, &ai.ChatRequest{
        Messages: []ai.Message{
            {
                Role: "system",
                Content: "You are an expert content writer. Generate high-quality, engaging blog posts.",
            },
            {
                Role: "user",
                Content: fmt.Sprintf("Write a blog post about: %s", topic),
            },
        },
        MaxTokens:   2000,
        Temperature: 0.8,
    })
    
    if err != nil {
        return "", err
    }
    
    return resp.Content, nil
}

func (cg *ContentGenerator) GenerateProductDescription(ctx context.Context, product *Product) (string, error) {
    prompt := fmt.Sprintf(`
        Generate a compelling product description for:
        Name: %s
        Category: %s
        Features: %s
        Price: $%.2f
    `, product.Name, product.Category, strings.Join(product.Features, ", "), product.Price)
    
    resp, err := cg.provider.ChatCompletion(ctx, &ai.ChatRequest{
        Messages: []ai.Message{
            {Role: "system", Content: "You are a marketing expert."},
            {Role: "user", Content: prompt},
        },
        MaxTokens:   500,
        Temperature: 0.7,
    })
    
    return resp.Content, err
}
```

### Sentiment Analysis

```go
type SentimentAnalyzer struct {
    provider ai.Provider
}

func (sa *SentimentAnalyzer) Analyze(ctx context.Context, text string) (string, float64, error) {
    resp, err := sa.provider.ChatCompletion(ctx, &ai.ChatRequest{
        Messages: []ai.Message{
            {
                Role: "system",
                Content: "Analyze the sentiment of the following text. Respond with POSITIVE, NEGATIVE, or NEUTRAL and a confidence score (0-1).",
            },
            {
                Role: "user",
                Content: text,
            },
        },
        Temperature: 0.3,
    })
    
    if err != nil {
        return "", 0, err
    }
    
    // Parse response
    parts := strings.Split(resp.Content, ",")
    sentiment := strings.TrimSpace(parts[0])
    confidence, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
    
    return sentiment, confidence, nil
}
```

### Smart Recommendations

```go
type RecommendationEngine struct {
    provider   ai.Provider
    vectorDB   *VectorStore
    embeddings *EmbeddingService
}

func (re *RecommendationEngine) GetRecommendations(ctx context.Context, userID int64, limit int) ([]Product, error) {
    // Get user history
    var history []Product
    db.Where("user_id = ?", userID).Order("created_at DESC").Limit(10).Find(&history)
    
    // Create user profile embedding
    userProfile := strings.Join(getProductDescriptions(history), " ")
    userEmbedding, err := re.embeddings.GetEmbedding(ctx, userProfile)
    if err != nil {
        return nil, err
    }
    
    // Find similar products
    var recommendations []Product
    err = db.Raw(`
        SELECT p.*, embedding <=> ? AS similarity
        FROM products p
        WHERE p.id NOT IN (
            SELECT product_id FROM user_purchases WHERE user_id = ?
        )
        ORDER BY similarity
        LIMIT ?
    `, pgvector.NewVector(userEmbedding), userID, limit).Scan(&recommendations).Error
    
    return recommendations, err
}
```

### Chatbot

```go
type Chatbot struct {
    provider ai.Provider
    memory   map[string][]ai.Message
    mu       sync.RWMutex
}

func (cb *Chatbot) Chat(ctx context.Context, userID, message string) (string, error) {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    // Get conversation history
    history := cb.memory[userID]
    if history == nil {
        history = []ai.Message{
            {
                Role:    "system",
                Content: "You are a helpful customer service assistant.",
            },
        }
    }
    
    // Add user message
    history = append(history, ai.Message{
        Role:    "user",
        Content: message,
    })
    
    // Generate response
    resp, err := cb.provider.ChatCompletion(ctx, &ai.ChatRequest{
        Messages:    history,
        MaxTokens:   500,
        Temperature: 0.7,
    })
    
    if err != nil {
        return "", err
    }
    
    // Update history
    history = append(history, ai.Message{
        Role:    "assistant",
        Content: resp.Content,
    })
    
    // Keep last 10 messages
    if len(history) > 10 {
        history = history[len(history)-10:]
    }
    
    cb.memory[userID] = history
    
    return resp.Content, nil
}
```

## Caching

### Response Cache

```go
type CachedAIProvider struct {
    provider ai.Provider
    cache    cache.Cache
}

func (cap *CachedAIProvider) ChatCompletion(ctx context.Context, req *ai.ChatRequest) (*ai.ChatResponse, error) {
    // Generate cache key
    cacheKey := generateCacheKey(req)
    
    // Check cache
    if cached, err := cap.cache.Get(ctx, cacheKey); err == nil {
        var resp ai.ChatResponse
        json.Unmarshal(cached, &resp)
        return &resp, nil
    }
    
    // Call provider
    resp, err := cap.provider.ChatCompletion(ctx, req)
    if err != nil {
        return nil, err
    }
    
    // Cache response (24 hours)
    data, _ := json.Marshal(resp)
    cap.cache.Set(ctx, cacheKey, data, 24*time.Hour)
    
    return resp, nil
}

func generateCacheKey(req *ai.ChatRequest) string {
    data, _ := json.Marshal(req)
    hash := sha256.Sum256(data)
    return fmt.Sprintf("ai:chat:%x", hash)
}
```

### Rate Limiting

```go
type RateLimitedProvider struct {
    provider ai.Provider
    limiter  *rate.Limiter
}

func NewRateLimitedProvider(provider ai.Provider, requestsPerMinute int) *RateLimitedProvider {
    return &RateLimitedProvider{
        provider: provider,
        limiter:  rate.NewLimiter(rate.Every(time.Minute/time.Duration(requestsPerMinute)), requestsPerMinute),
    }
}

func (rlp *RateLimitedProvider) ChatCompletion(ctx context.Context, req *ai.ChatRequest) (*ai.ChatResponse, error) {
    if err := rlp.limiter.Wait(ctx); err != nil {
        return nil, err
    }
    
    return rlp.provider.ChatCompletion(ctx, req)
}
```

## Best Practices

### 1. Error Handling

```go
func SafeAICall(provider ai.Provider, req *ai.ChatRequest) (string, error) {
    maxRetries := 3
    
    for i := 0; i < maxRetries; i++ {
        resp, err := provider.ChatCompletion(ctx, req)
        if err == nil {
            return resp.Content, nil
        }
        
        // Check if retryable
        if isRateLimitError(err) {
            time.Sleep(time.Duration(i+1) * time.Second)
            continue
        }
        
        return "", err
    }
    
    return "", fmt.Errorf("max retries exceeded")
}
```

### 2. Cost Management

```go
type CostTracker struct {
    totalTokens int64
    totalCost   float64
    mu          sync.Mutex
}

func (ct *CostTracker) TrackUsage(usage ai.Usage) {
    ct.mu.Lock()
    defer ct.mu.Unlock()
    
    ct.totalTokens += int64(usage.TotalTokens)
    
    // GPT-4 pricing (example)
    inputCost := float64(usage.PromptTokens) * 0.03 / 1000
    outputCost := float64(usage.CompletionTokens) * 0.06 / 1000
    
    ct.totalCost += inputCost + outputCost
}
```

### 3. Prompt Engineering

```go
func BuildPrompt(context, instruction, input string) string {
    return fmt.Sprintf(`
Context: %s

Instruction: %s

Input: %s

Output:`, context, instruction, input)
}

// Example
prompt := BuildPrompt(
    "You are an expert software engineer.",
    "Write a function that validates email addresses.",
    "Language: Go",
)
```

### 4. Testing

```go
type MockAIProvider struct{}

func (m *MockAIProvider) ChatCompletion(ctx context.Context, req *ai.ChatRequest) (*ai.ChatResponse, error) {
    return &ai.ChatResponse{
        Content: "Mock response",
        Usage: ai.Usage{
            PromptTokens:     10,
            CompletionTokens: 20,
            TotalTokens:      30,
        },
    }, nil
}

func TestContentGenerator(t *testing.T) {
    mock := &MockAIProvider{}
    generator := &ContentGenerator{provider: mock}
    
    content, err := generator.GenerateBlogPost(ctx, "AI in Healthcare")
    
    assert.NoError(t, err)
    assert.NotEmpty(t, content)
}
```

---

**Next Steps:**
- Learn about [Vector Databases](../database/vector.md)
- Explore [Caching](cache.md) for performance
- See [Queue](queue.md) for async processing

**Related Topics:**
- [OpenAI API](https://platform.openai.com/docs)
- [LangChain](https://langchain.com/)
- [Vector Search](https://www.pinecone.io/)
