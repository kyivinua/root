package enricher

import (
	"context"
	"time"
)

// EnrichmentTarget represents the target for enrichment
type EnrichmentTarget interface {
	GetIdentifier() string
	GetCurrentDocs() string
	SetEnrichedDocs(string)
	GetContext() map[string]interface{}
}

// LLMClient provides abstract LLM access
type LLMClient interface {
	GenerateCompletion(ctx context.Context, prompt string) (string, error)
	GenerateCompletionWithConfig(ctx context.Context, prompt string, config map[string]interface{}) (string, error)
	GetModelName() string
	GetProviderName() string
}

// RAGRetriever retrieves relevant context from vector store
type RAGRetriever interface {
	RetrieveContext(ctx context.Context, query string, topK int) ([]RAGDocument, error)
	IndexDocument(ctx context.Context, doc RAGDocument) error
	DeleteByID(ctx context.Context, id string) error
}

// RAGDocument represents a document in the vector store
type RAGDocument struct {
	ID       string
	Content  string
	Metadata map[string]interface{}
	Score    float64
}

// EnrichmentCache provides caching for enriched documentation
type EnrichmentCache interface {
	Get(ctx context.Context, key string) (string, bool)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context) error
}

// PromptTemplateEngine renders prompt templates
type PromptTemplateEngine interface {
	Render(templateName string, data map[string]interface{}) (string, error)
	RegisterTemplate(name string, template string) error
}

// SmartStrategy decides when to use RAG vs. base LLM
type SmartStrategy interface {
	ShouldUseRAG(target EnrichmentTarget, metadata map[string]interface{}) bool
	GetTopKForTarget(target EnrichmentTarget) int
}

// SafetyGuard validates enriched content
type SafetyGuard interface {
	Validate(ctx context.Context, original string, enriched string, target EnrichmentTarget) (*SafetyReport, error)
}

// SafetyReport contains validation results
type SafetyReport struct {
	Passed          bool
	Confidence      float64
	FailureReasons  []string
	EntropyScore    float64
	JudgeVerdict    string
	DetectedIssues  []SafetyIssue
	Timestamp       time.Time
}

// SafetyIssue represents a specific safety concern
type SafetyIssue struct {
	Type     string
	Severity string
	Message  string
	Location string
}

// PolicyEngine enforces tenant isolation and approval workflows
type PolicyEngine interface {
	Evaluate(ctx context.Context, target EnrichmentTarget, tenant string) (*PolicyDecision, error)
	GetPolicyForTenant(tenant string) (*TenantPolicy, error)
}

// PolicyDecision represents the result of policy evaluation
type PolicyDecision struct {
	Allowed         bool
	RequiresApproval bool
	Reason          string
	Tenant          string
	ApprovedBy      string
	ApprovedAt      *time.Time
}

// TenantPolicy defines per-tenant enrichment rules
type TenantPolicy struct {
	TenantID            string
	AllowEnrichment     bool
	RequireApproval     bool
	MaxTokens           int
	AllowedVisibilities []string
	CustomRules         map[string]interface{}
}

// EnrichmentMetrics collects enrichment statistics
type EnrichmentMetrics interface {
	RecordEnrichment(target string, duration time.Duration, tokensUsed int, success bool)
	RecordCacheHit(target string)
	RecordCacheMiss(target string)
	RecordRAGRetrieval(query string, resultsCount int, duration time.Duration)
	RecordSafetyCheck(passed bool, duration time.Duration)
	RecordCost(provider string, model string, tokens int, cost float64)
}

// TraceSink records enrichment audit trail
type TraceSink interface {
	RecordTrace(ctx context.Context, trace *EnrichmentTrace) error
	QueryTraces(ctx context.Context, filter TraceFilter) ([]*EnrichmentTrace, error)
}

// EnrichmentTrace represents full audit trail for one enrichment
type EnrichmentTrace struct {
	TraceID         string
	Tenant          string
	TargetID        string
	TargetType      string
	Timestamp       time.Time
	Duration        time.Duration
	OriginalDocs    string
	EnrichedDocs    string
	PromptUsed      string
	ModelUsed       string
	RAGContext      []RAGDocument
	SafetyReport    *SafetyReport
	PolicyDecision  *PolicyDecision
	Success         bool
	ErrorMessage    string
	TokensUsed      int
	Cost            float64
	CacheHit        bool
}

// TraceFilter for querying enrichment traces
type TraceFilter struct {
	TenantID   string
	TargetID   string
	StartTime  time.Time
	EndTime    time.Time
	Success    *bool
	Limit      int
	Offset     int
}

// ApiDocModel represents the documentation model
type ApiDocModel struct {
	Modules      []ModuleDoc       `json:"modules"`
	GeneratedAt  time.Time         `json:"generated_at"`
	SourceCommit string            `json:"source_commit"`
	Tools        map[string]string `json:"tools,omitempty"`
	Statistics   map[string]int64  `json:"statistics,omitempty"`
}

// ModuleDoc represents a package/module
type ModuleDoc struct {
	PackageName string        `json:"package_name"`
	FilePath    string        `json:"file_path"`
	Summary     string        `json:"summary"`
	Services    []ServiceDoc  `json:"services"`
	Messages    []MessageDoc  `json:"messages"`
	Enums       []EnumDoc     `json:"enums"`
}

// ServiceDoc represents a gRPC service
type ServiceDoc struct {
	Name     string      `json:"name"`
	FullName string      `json:"full_name"`
	Summary  string      `json:"summary"`
	Methods  []MethodDoc `json:"methods"`
}

// MethodDoc represents a service method
type MethodDoc struct {
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Summary     string `json:"summary"`
	InputType   string `json:"input_type"`
	OutputType  string `json:"output_type"`
	IsStreaming bool   `json:"is_streaming,omitempty"`
}

// MessageDoc represents a protobuf message
type MessageDoc struct {
	Name     string     `json:"name"`
	FullName string     `json:"full_name"`
	Summary  string     `json:"summary"`
	Fields   []FieldDoc `json:"fields"`
}

// FieldDoc represents a message field
type FieldDoc struct {
	Name     string `json:"name"`
	Number   int    `json:"number"`
	Type     string `json:"type"`
	Summary  string `json:"summary"`
	Repeated bool   `json:"repeated,omitempty"`
}

// EnumDoc represents a protobuf enum
type EnumDoc struct {
	Name     string         `json:"name"`
	FullName string         `json:"full_name"`
	Summary  string         `json:"summary"`
	Values   []EnumValueDoc `json:"values"`
}

// EnumValueDoc represents an enum value
type EnumValueDoc struct {
	Name    string `json:"name"`
	Number  int    `json:"number"`
	Summary string `json:"summary"`
}

// MethodTarget implements EnrichmentTarget for method documentation
type MethodTarget struct {
	Method      *MethodDoc
	Service     *ServiceDoc
	Module      *ModuleDoc
	enrichedDoc string
}

func (mt *MethodTarget) GetIdentifier() string {
	return mt.Method.FullName
}

func (mt *MethodTarget) GetCurrentDocs() string {
	return mt.Method.Summary
}

func (mt *MethodTarget) SetEnrichedDocs(docs string) {
	mt.enrichedDoc = docs
	mt.Method.Summary = docs
}

func (mt *MethodTarget) GetContext() map[string]interface{} {
	return map[string]interface{}{
		"method_name":  mt.Method.Name,
		"service_name": mt.Service.Name,
		"package_name": mt.Module.PackageName,
		"input_type":   mt.Method.InputType,
		"output_type":  mt.Method.OutputType,
		"is_streaming": mt.Method.IsStreaming,
	}
}

// MessageTarget implements EnrichmentTarget for message documentation
type MessageTarget struct {
	Message     *MessageDoc
	Module      *ModuleDoc
	enrichedDoc string
}

func (mt *MessageTarget) GetIdentifier() string {
	return mt.Message.FullName
}

func (mt *MessageTarget) GetCurrentDocs() string {
	return mt.Message.Summary
}

func (mt *MessageTarget) SetEnrichedDocs(docs string) {
	mt.enrichedDoc = docs
	mt.Message.Summary = docs
}

func (mt *MessageTarget) GetContext() map[string]interface{} {
	fieldNames := make([]string, len(mt.Message.Fields))
	for i, f := range mt.Message.Fields {
		fieldNames[i] = f.Name
	}
	return map[string]interface{}{
		"message_name": mt.Message.Name,
		"package_name": mt.Module.PackageName,
		"fields":       fieldNames,
		"field_count":  len(mt.Message.Fields),
	}
}
