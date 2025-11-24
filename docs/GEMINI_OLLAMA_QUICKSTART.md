# Quick Start: Google AI Studio + Ollama Integration

**Цель:** Быстро добавить Gemini и улучшить Ollama за 2-3 часа

---

## Part 1: Google Gemini Integration (60 минут)

### Step 1: Get API Key (5 min)

```bash
# 1. Перейти на https://ai.google.dev/
# 2. Нажать "Get API Key"
# 3. Создать новый проект или выбрать существующий
# 4. Скопировать API key

export GOOGLE_API_KEY="AIza..."

# Тест
curl "https://generativelanguage.googleapis.com/v1beta/models?key=$GOOGLE_API_KEY"
```

### Step 2: Add Dependencies (5 min)

```bash
cd /home/user/root

# Добавить SDK
go get google.golang.org/genai@latest

# Или использовать REST API (без зависимостей)
# Уже есть net/http в проекте
```

### Step 3: Implement Client (30 min)

**Создать файл:** `tools/protodocs/hldgen/llm_client_gemini.go`

```go
package hldgen

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

// GeminiClient implements LLMClient for Google Gemini via REST API
type GeminiClient struct {
    cfg        ProviderConfig
    httpClient *http.Client
    apiKey     string
}

// NewGeminiClient creates a new Gemini client
func NewGeminiClient(cfg ProviderConfig) (*GeminiClient, error) {
    if cfg.APIKey == "" {
        return nil, fmt.Errorf("google API key not configured")
    }

    return &GeminiClient{
        cfg:    cfg,
        apiKey: cfg.APIKey,
        httpClient: &http.Client{
            Timeout: 90 * time.Second,
        },
    }, nil
}

// geminiRequest represents the API request
type geminiRequest struct {
    Contents []geminiContent          `json:"contents"`
    GenerationConfig geminiGenConfig  `json:"generationConfig,omitempty"`
    SafetySettings []geminiSafety     `json:"safetySettings,omitempty"`
}

type geminiContent struct {
    Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
    Text string `json:"text"`
}

type geminiGenConfig struct {
    Temperature     float64 `json:"temperature,omitempty"`
    MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
}

type geminiSafety struct {
    Category  string `json:"category"`
    Threshold string `json:"threshold"`
}

// geminiResponse represents the API response
type geminiResponse struct {
    Candidates []struct {
        Content struct {
            Parts []struct {
                Text string `json:"text"`
            } `json:"parts"`
        } `json:"content"`
        FinishReason string `json:"finishReason"`
    } `json:"candidates"`
    UsageMetadata struct {
        PromptTokenCount     int `json:"promptTokenCount"`
        CandidatesTokenCount int `json:"candidatesTokenCount"`
    } `json:"usageMetadata"`
}

// Generate implements LLMClient.Generate
func (g *GeminiClient) Generate(ctx context.Context, prompt string, config LLMConfig) (*LLMResponse, error) {
    // Construct request
    reqBody := geminiRequest{
        Contents: []geminiContent{
            {
                Parts: []geminiPart{
                    {Text: prompt},
                },
            },
        },
        GenerationConfig: geminiGenConfig{
            Temperature:     g.cfg.Temperature,
            MaxOutputTokens: g.cfg.MaxTokens,
        },
        SafetySettings: []geminiSafety{
            {Category: "HARM_CATEGORY_HARASSMENT", Threshold: "BLOCK_MEDIUM_AND_ABOVE"},
            {Category: "HARM_CATEGORY_HATE_SPEECH", Threshold: "BLOCK_MEDIUM_AND_ABOVE"},
            {Category: "HARM_CATEGORY_DANGEROUS_CONTENT", Threshold: "BLOCK_MEDIUM_AND_ABOVE"},
        },
    }

    reqJSON, err := json.Marshal(reqBody)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal request: %w", err)
    }

    // Build URL
    model := g.cfg.Model
    if model == "" {
        model = "gemini-1.5-pro"
    }
    url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
        model, g.apiKey)

    // Create HTTP request
    httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqJSON))
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }

    httpReq.Header.Set("Content-Type", "application/json")

    // Execute request
    resp, err := g.httpClient.Do(httpReq)
    if err != nil {
        return nil, fmt.Errorf("gemini API request failed: %w", err)
    }
    defer resp.Body.Close()

    // Read response
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response: %w", err)
    }

    // Check status
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("gemini API error (status %d): %s", resp.StatusCode, string(body))
    }

    // Parse response
    var geminiResp geminiResponse
    if err := json.Unmarshal(body, &geminiResp); err != nil {
        return nil, fmt.Errorf("failed to parse response: %w", err)
    }

    // Extract content
    if len(geminiResp.Candidates) == 0 {
        return nil, fmt.Errorf("no candidates in response")
    }

    candidate := geminiResp.Candidates[0]
    if len(candidate.Content.Parts) == 0 {
        return nil, fmt.Errorf("no content parts in response")
    }

    text := candidate.Content.Parts[0].Text

    // Calculate tokens
    tokensUsed := geminiResp.UsageMetadata.PromptTokenCount +
        geminiResp.UsageMetadata.CandidatesTokenCount

    return &LLMResponse{
        Content:      text,
        TokensUsed:   tokensUsed,
        Model:        model,
        Provider:     "google",
        Confidence:   0.87,  // Between Claude (0.90) and Ollama (0.85)
        FinishReason: candidate.FinishReason,
    }, nil
}

func (g *GeminiClient) GetModelName() string {
    if g.cfg.Model == "" {
        return "gemini-1.5-pro"
    }
    return g.cfg.Model
}

func (g *GeminiClient) GetProviderName() string {
    return "google"
}
```

### Step 4: Register in Router (5 min)

**Редактировать:** `tools/protodocs/hldgen/llm_client.go`

```go
func createLLMClient(cfg ProviderConfig) (LLMClient, error) {
    switch cfg.Name {
    case "anthropic":
        return NewAnthropicClient(cfg)
    case "openai":
        return NewOpenAIClient(cfg)
    case "ollama":
        return NewOllamaClient(cfg)
    case "google", "gemini":  // ← ADD THIS
        return NewGeminiClient(cfg)
    default:
        return nil, fmt.Errorf("unsupported provider: %s", cfg.Name)
    }
}
```

### Step 5: Configure (5 min)

**Создать/обновить:** `config/hldgen.yaml` или `.env`

```yaml
llm:
  providers:
    - name: "google"
      model: "gemini-1.5-pro"
      temperature: 0.2
      max_tokens: 8192
      api_key: "${GOOGLE_API_KEY}"
      weight: 0.7
```

### Step 6: Test (10 min)

**Создать:** `tools/protodocs/hldgen/llm_client_gemini_test.go`

```go
package hldgen

import (
    "context"
    "os"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestGeminiClient_Basic(t *testing.T) {
    apiKey := os.Getenv("GOOGLE_API_KEY")
    if apiKey == "" {
        t.Skip("GOOGLE_API_KEY not set")
    }

    cfg := ProviderConfig{
        Name:        "google",
        Model:       "gemini-1.5-flash",  // Faster/cheaper for tests
        APIKey:      apiKey,
        Temperature: 0.0,
        MaxTokens:   1024,
    }

    client, err := NewGeminiClient(cfg)
    require.NoError(t, err)

    ctx := context.Background()
    prompt := "Write a one-sentence summary of: service UserService { rpc GetUser(ID) returns (User); }"

    resp, err := client.Generate(ctx, prompt, LLMConfig{})
    require.NoError(t, err)

    assert.NotEmpty(t, resp.Content)
    assert.Greater(t, resp.TokensUsed, 0)
    assert.Equal(t, "google", resp.Provider)
    assert.Contains(t, resp.Content, "User")  // Should mention UserService

    t.Logf("Generated: %s", resp.Content)
    t.Logf("Tokens: %d", resp.TokensUsed)
}

func TestGeminiClient_ErrorHandling(t *testing.T) {
    cfg := ProviderConfig{
        Name:   "google",
        APIKey: "invalid_key",
    }

    client, err := NewGeminiClient(cfg)
    require.NoError(t, err)

    ctx := context.Background()
    _, err = client.Generate(ctx, "test", LLMConfig{})
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "API error")
}
```

**Запустить тест:**

```bash
export GOOGLE_API_KEY="your_key_here"
go test ./tools/protodocs/hldgen -v -run TestGeminiClient
```

---

## Part 2: Ollama Production Hardening (60 минут)

### Step 1: Model Selection & Download (15 min)

```bash
# Install Ollama (if not installed)
curl -fsSL https://ollama.com/install.sh | sh

# Download recommended models
ollama pull llama3.1:70b-q5_k_m   # Production (48GB, best quality)
ollama pull qwen2.5:32b           # Development (19GB, fast)
ollama pull gemma2:9b             # Fallback (5GB, very fast)

# Verify models
ollama list

# Test generation
ollama run llama3.1:70b-q5_k_m "Summarize gRPC in one sentence"
```

### Step 2: Enhanced Ollama Client (30 min)

**Обновить:** `tools/protodocs/hldgen/llm_client.go` (секция Ollama)

```go
// OllamaClient implements LLMClient for local Ollama models
type OllamaClient struct {
    cfg         ProviderConfig
    httpClient  *http.Client
    baseURL     string
    healthCheck *time.Time
    isHealthy   bool
}

// NewOllamaClient creates a new Ollama client with health checking
func NewOllamaClient(cfg ProviderConfig) (*OllamaClient, error) {
    baseURL := cfg.BaseURL
    if baseURL == "" {
        baseURL = "http://localhost:11434"
    }

    client := &OllamaClient{
        cfg:     cfg,
        baseURL: baseURL,
        httpClient: &http.Client{
            Timeout: 120 * time.Second,
            Transport: &http.Transport{
                MaxIdleConns:       10,
                IdleConnTimeout:    30 * time.Second,
                DisableCompression: false,
            },
        },
    }

    // Initial health check
    if err := client.checkHealth(); err != nil {
        return nil, fmt.Errorf("ollama server unhealthy: %w", err)
    }

    // Warm up model (load into memory)
    go client.warmUp()

    return client, nil
}

// checkHealth verifies Ollama server is responsive
func (o *OllamaClient) checkHealth() error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    req, err := http.NewRequestWithContext(ctx, "GET", o.baseURL+"/api/tags", nil)
    if err != nil {
        return err
    }

    resp, err := o.httpClient.Do(req)
    if err != nil {
        o.isHealthy = false
        return fmt.Errorf("health check failed: %w", err)
    }
    defer resp.Body.Close()

    o.isHealthy = resp.StatusCode == http.StatusOK
    now := time.Now()
    o.healthCheck = &now

    return nil
}

// warmUp pre-loads the model into memory
func (o *OllamaClient) warmUp() {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Minimal prompt to trigger model load
    _, _ = o.Generate(ctx, "Hi", LLMConfig{})
}

// Generate with retry logic
func (o *OllamaClient) Generate(ctx context.Context, prompt string, config LLMConfig) (*LLMResponse, error) {
    // Check health every 5 minutes
    if o.healthCheck == nil || time.Since(*o.healthCheck) > 5*time.Minute {
        if err := o.checkHealth(); err != nil {
            return nil, err
        }
    }

    if !o.isHealthy {
        return nil, fmt.Errorf("ollama server is unhealthy")
    }

    // Retry with exponential backoff
    var lastErr error
    for attempt := 0; attempt < 3; attempt++ {
        resp, err := o.generateOnce(ctx, prompt, config)
        if err == nil {
            return resp, nil
        }

        lastErr = err

        // Don't retry on context errors
        if ctx.Err() != nil {
            return nil, ctx.Err()
        }

        // Exponential backoff: 1s, 2s, 4s
        if attempt < 2 {
            backoff := time.Duration(1<<uint(attempt)) * time.Second
            time.Sleep(backoff)
        }
    }

    return nil, fmt.Errorf("ollama failed after 3 attempts: %w", lastErr)
}

// generateOnce makes a single generation attempt
func (o *OllamaClient) generateOnce(ctx context.Context, prompt string, config LLMConfig) (*LLMResponse, error) {
    model := o.cfg.Model
    if model == "" {
        model = "llama3.1:70b-q5_k_m"  // Updated default
    }

    reqBody := map[string]interface{}{
        "model":  model,
        "prompt": prompt,
        "stream": false,
        "options": map[string]interface{}{
            "temperature": o.cfg.Temperature,
            "num_predict": o.cfg.MaxTokens,
        },
    }

    reqJSON, err := json.Marshal(reqBody)
    if err != nil {
        return nil, err
    }

    req, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/api/generate", bytes.NewReader(reqJSON))
    if err != nil {
        return nil, err
    }

    req.Header.Set("Content-Type", "application/json")

    resp, err := o.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("ollama request failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("ollama error (status %d): %s", resp.StatusCode, string(body))
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    var ollamaResp struct {
        Response         string `json:"response"`
        TotalDuration    int64  `json:"total_duration"`
        LoadDuration     int64  `json:"load_duration"`
        PromptEvalCount  int    `json:"prompt_eval_count"`
        EvalCount        int    `json:"eval_count"`
    }

    if err := json.Unmarshal(body, &ollamaResp); err != nil {
        return nil, err
    }

    tokensUsed := ollamaResp.PromptEvalCount + ollamaResp.EvalCount

    return &LLMResponse{
        Content:      ollamaResp.Response,
        TokensUsed:   tokensUsed,
        Model:        model,
        Provider:     "ollama",
        Confidence:   0.85,
        FinishReason: "stop",
    }, nil
}
```

### Step 3: Model Configuration (5 min)

**Обновить:** `config/hldgen.yaml`

```yaml
llm:
  providers:
    - name: "ollama"
      model: "llama3.1:70b-q5_k_m"  # Production quality
      base_url: "http://localhost:11434"
      temperature: 0.1
      max_tokens: 4096
      weight: 0.5

  # Альтернативный Ollama для dev
  dev_providers:
    - name: "ollama"
      model: "qwen2.5:32b"  # Faster for development
      base_url: "http://localhost:11434"
```

### Step 4: GPU Monitoring (10 min)

**Создать:** `tools/protodocs/hldgen/gpu_monitor.go`

```go
package hldgen

import (
    "fmt"
    "os/exec"
    "strconv"
    "strings"
)

// GPUMonitor checks GPU availability and usage
type GPUMonitor struct{}

// GPUMetrics represents GPU utilization
type GPUMetrics struct {
    UtilizationPercent int
    MemoryUsedMB       int
    MemoryTotalMB      int
    MemoryAvailableMB  int
}

// GetMetrics queries GPU status via nvidia-smi
func (m *GPUMonitor) GetMetrics() (*GPUMetrics, error) {
    cmd := exec.Command("nvidia-smi",
        "--query-gpu=utilization.gpu,memory.used,memory.total",
        "--format=csv,noheader,nounits")

    output, err := cmd.Output()
    if err != nil {
        return nil, fmt.Errorf("nvidia-smi failed: %w", err)
    }

    // Parse output: "85, 38912, 49152"
    parts := strings.Split(strings.TrimSpace(string(output)), ",")
    if len(parts) != 3 {
        return nil, fmt.Errorf("unexpected nvidia-smi output")
    }

    util, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
    used, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
    total, _ := strconv.Atoi(strings.TrimSpace(parts[2]))

    return &GPUMetrics{
        UtilizationPercent: util,
        MemoryUsedMB:       used,
        MemoryTotalMB:      total,
        MemoryAvailableMB:  total - used,
    }, nil
}

// IsAvailable checks if GPU is ready for use
func (m *GPUMonitor) IsAvailable() bool {
    metrics, err := m.GetMetrics()
    if err != nil {
        return false
    }

    // GPU available if < 80% utilized and > 10GB free
    return metrics.UtilizationPercent < 80 &&
        metrics.MemoryAvailableMB > 10000
}
```

**Использование:**

```go
// В LLMRouter
monitor := &GPUMonitor{}

if monitor.IsAvailable() {
    // Use Ollama
    return ollamaClient.Generate(ctx, prompt, config)
} else {
    // Fallback to cloud
    return anthropicClient.Generate(ctx, prompt, config)
}
```

---

## Part 3: Integration Testing (30 min)

### Test Full Pipeline

**Создать:** `tools/protodocs/hldgen/integration_test.go`

```go
package hldgen

import (
    "context"
    "os"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestLLMRouter_AllProviders(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    // Test all 4 providers
    providers := []ProviderConfig{
        {
            Name:        "anthropic",
            Model:       "claude-3-5-sonnet",
            APIKey:      os.Getenv("ANTHROPIC_API_KEY"),
            Temperature: 0.0,
        },
        {
            Name:        "openai",
            Model:       "gpt-4",
            APIKey:      os.Getenv("OPENAI_API_KEY"),
            Temperature: 0.0,
        },
        {
            Name:        "google",
            Model:       "gemini-1.5-flash",
            APIKey:      os.Getenv("GOOGLE_API_KEY"),
            Temperature: 0.0,
        },
        {
            Name:        "ollama",
            Model:       "qwen2.5:32b",
            BaseURL:     "http://localhost:11434",
            Temperature: 0.0,
        },
    }

    prompt := "In one sentence, what is gRPC?"

    for _, cfg := range providers {
        t.Run(cfg.Name, func(t *testing.T) {
            if cfg.APIKey == "" && cfg.Name != "ollama" {
                t.Skipf("%s API key not set", cfg.Name)
            }

            client, err := createLLMClient(cfg)
            require.NoError(t, err)

            ctx := context.Background()
            resp, err := client.Generate(ctx, prompt, LLMConfig{})

            require.NoError(t, err)
            assert.NotEmpty(t, resp.Content)
            assert.Contains(t, resp.Content, "RPC")  // Should mention RPC
            assert.Greater(t, resp.TokensUsed, 0)

            t.Logf("[%s] Response: %s", cfg.Name, resp.Content)
            t.Logf("[%s] Tokens: %d", cfg.Name, resp.TokensUsed)
        })
    }
}

func TestLLMRouter_CostComparison(t *testing.T) {
    // Compare costs for same prompt
    prompt := "Explain gRPC services in 100 words"

    costs := map[string]float64{
        "anthropic": 0.003,  // $3/1M input tokens
        "openai":    0.01,   // $10/1M
        "google":    0.00125,// $1.25/1M
        "ollama":    0.0,    // Free
    }

    for provider, costPer1M := range costs {
        estimatedTokens := 150  // ~100 words = 150 tokens
        estimatedCost := float64(estimatedTokens) / 1000000 * costPer1M

        t.Logf("%s: ~$%.6f for this prompt", provider, estimatedCost)
    }

    // Expected output:
    // anthropic: ~$0.000450
    // openai:    ~$0.001500
    // google:    ~$0.000188 ← Cheapest cloud option
    // ollama:    ~$0.000000 ← Free but slower
}
```

**Run all tests:**

```bash
export ANTHROPIC_API_KEY="sk-..."
export OPENAI_API_KEY="sk-..."
export GOOGLE_API_KEY="AIza..."

go test ./tools/protodocs/hldgen -v -run TestLLMRouter
```

---

## Part 4: Configuration Examples

### Development Config

```yaml
# config/dev.yaml
llm:
  router:
    strategy: "fastest"

  providers:
    # Local first for dev
    - name: "ollama"
      model: "qwen2.5:32b"
      weight: 0.9

    # Cheap cloud fallback
    - name: "google"
      model: "gemini-1.5-flash"
      weight: 0.5

refinement:
  max_rounds: 1  # Fast iterations
```

### Production Config

```yaml
# config/prod.yaml
llm:
  router:
    strategy: "quality_first"

  providers:
    # Best quality
    - name: "anthropic"
      model: "claude-3-5-sonnet"
      weight: 0.9

    # Cost-effective
    - name: "google"
      model: "gemini-1.5-pro"
      weight: 0.7

    # Local fallback
    - name: "ollama"
      model: "llama3.1:70b-q5_k_m"
      weight: 0.5

  cost_limits:
    daily_budget_usd: 100.0

refinement:
  max_rounds: 3
  consensus_threshold: 0.88
```

---

## Verification Checklist

После внедрения проверьте:

```bash
# ✅ Gemini works
go test ./tools/protodocs/hldgen -v -run TestGeminiClient

# ✅ Ollama health check works
curl http://localhost:11434/api/tags

# ✅ All providers available
go test ./tools/protodocs/hldgen -v -run TestLLMRouter_AllProviders

# ✅ GPU monitoring works (if GPU available)
nvidia-smi

# ✅ Full pipeline works
go run ./cmd/protodocs-hld generate --config config/prod.yaml
```

---

## Cost Analysis

**Scenario:** 1000 documentation generations per month

| Provider | Cost/1M Tokens | Avg Tokens | Monthly Cost |
|----------|----------------|------------|--------------|
| Claude 3.5 | $3 input | 15K input | $45 |
| GPT-4 | $10 input | 15K input | $150 |
| **Gemini 1.5 Pro** | **$1.25 input** | 15K input | **$18.75** 🏆 |
| Ollama | $0 | 15K input | $0 (but hardware cost) |

**Recommended split for cost optimization:**
- 40% Claude (complex tasks) = $18
- 40% Gemini (medium tasks) = $7.50
- 20% Ollama (simple tasks) = $0

**Total: ~$25.50/month** vs $45 (Claude only)

---

## Troubleshooting

### Gemini Issues

**Error: API key not valid**
```bash
# Re-create key
echo $GOOGLE_API_KEY
# Should start with "AIza"

# Test directly
curl "https://generativelanguage.googleapis.com/v1beta/models?key=$GOOGLE_API_KEY"
```

**Error: Resource exhausted (quota)**
```
Free tier: 15 RPM
Solution: Enable billing → 1000 RPM
```

### Ollama Issues

**Error: Connection refused**
```bash
# Check if running
systemctl status ollama

# Start if stopped
systemctl start ollama

# Check logs
journalctl -u ollama -f
```

**Error: Model not found**
```bash
# List available models
ollama list

# Pull missing model
ollama pull llama3.1:70b-q5_k_m
```

**Slow generation (> 30s)**
```bash
# Check GPU usage
nvidia-smi

# If CPU-only, expect 5-10x slower
# Solution: Use smaller model or add GPU
ollama pull qwen2.5:32b  # Faster model
```

---

## Next Steps

После базовой интеграции:

1. **Мониторинг** - добавить metrics для всех провайдеров
2. **A/B Testing** - сравнить качество выхода разных моделей
3. **Prompt Optimization** - настроить промпты под каждую модель
4. **Cost Dashboard** - визуализация затрат в Grafana

Готово к production! 🚀
