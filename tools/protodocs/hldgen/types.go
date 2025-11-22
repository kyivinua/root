package hldgen

import (
	"time"
)

// =============================================================================
// Core Types
// =============================================================================

// ConsolidatedDocs represents the input documentation from ProtoDocs
type ConsolidatedDocs struct {
	ModuleName    string                 `json:"module_name"`
	Services      []ServiceDoc           `json:"services"`
	Messages      []MessageDoc           `json:"messages"`
	Enums         []EnumDoc              `json:"enums"`
	Statistics    map[string]int64       `json:"statistics"`
	Metadata      map[string]interface{} `json:"metadata"`
	EnrichedDocs  string                 `json:"enriched_docs,omitempty"` // From enricher
	OpenAPISpec   string                 `json:"openapi_spec,omitempty"`
	GeneratedAt   time.Time              `json:"generated_at"`
	SourceCommit  string                 `json:"source_commit"`
}

type ServiceDoc struct {
	Name        string      `json:"name"`
	FullName    string      `json:"full_name"`
	Description string      `json:"description"`
	Methods     []MethodDoc `json:"methods"`
	Visibility  string      `json:"visibility"`
}

type MethodDoc struct {
	Name            string   `json:"name"`
	FullName        string   `json:"full_name"`
	Description     string   `json:"description"`
	InputType       string   `json:"input_type"`
	OutputType      string   `json:"output_type"`
	ClientStreaming bool     `json:"client_streaming"`
	ServerStreaming bool     `json:"server_streaming"`
	HTTPMethods     []string `json:"http_methods,omitempty"`
}

type MessageDoc struct {
	Name        string     `json:"name"`
	FullName    string     `json:"full_name"`
	Description string     `json:"description"`
	Fields      []FieldDoc `json:"fields"`
}

type FieldDoc struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	TypeName    string `json:"type_name,omitempty"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

type EnumDoc struct {
	Name        string          `json:"name"`
	FullName    string          `json:"full_name"`
	Description string          `json:"description"`
	Values      []EnumValueDoc  `json:"values"`
}

type EnumValueDoc struct {
	Name        string `json:"name"`
	Number      int32  `json:"number"`
	Description string `json:"description"`
}

// =============================================================================
// Agent Types
// =============================================================================

// AgentRole represents the role of an agent in the multi-agent system
type AgentRole string

const (
	RoleArchitect AgentRole = "architect"
	RolePM        AgentRole = "pm"
	RoleSecurity  AgentRole = "security"
	RoleSRE       AgentRole = "sre"
	RoleQA        AgentRole = "qa"
	RoleCritic    AgentRole = "critic"
)

// AgentInput represents input to an agent
type AgentInput struct {
	Docs          *ConsolidatedDocs
	EnrichedCtx   *EnrichedContext
	PreviousDraft *HLDOutput
	Round         int
	Criticism     []Criticism
}

// AgentResponse represents an agent's response
type AgentResponse struct {
	Role        AgentRole              `json:"role"`
	Content     string                 `json:"content"`
	Diagrams    map[string]string      `json:"diagrams,omitempty"`
	SLO         map[string]SLO         `json:"slo,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Confidence  float64                `json:"confidence"`
	TokensUsed  int                    `json:"tokens_used"`
	GeneratedAt time.Time              `json:"generated_at"`
}

// SLO represents a Service Level Objective
type SLO struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Target      float64 `json:"target"`      // e.g., 0.9999 for 99.99%
	Metric      string  `json:"metric"`      // e.g., "availability", "latency_p95"
	Threshold   string  `json:"threshold"`   // e.g., "< 150ms", "99.99%"
}

// =============================================================================
// Criticism Types
// =============================================================================

// Criticism represents critique from the Critic Agent
type Criticism struct {
	Agent      string   `json:"agent"`
	Score      float64  `json:"score"`          // 0.0-1.0
	Confidence float64  `json:"confidence"`     // 0.0-1.0
	Issues     []Issue  `json:"issues"`
	Suggestion string   `json:"suggestion"`
	Fatal      bool     `json:"fatal,omitempty"`
}

// Issue represents a specific problem found by the Critic
type Issue struct {
	Type     IssueType     `json:"type"`
	Severity IssueSeverity `json:"severity"`
	Path     string        `json:"path"`        // e.g., "components[2].data_flow"
	Message  string        `json:"message"`
	Evidence string        `json:"evidence"`
	Fix      string        `json:"fix,omitempty"`
}

type IssueType string

const (
	IssueTypeMissingComponent  IssueType = "missing_component"
	IssueTypeSecurityGap       IssueType = "security_gap"
	IssueTypeIncorrectFlow     IssueType = "incorrect_flow"
	IssueTypeMissingSLO        IssueType = "missing_slo"
	IssueTypeComplianceGap     IssueType = "compliance_gap"
	IssueTypeHallucination     IssueType = "hallucination"
	IssueTypeInconsistency     IssueType = "inconsistency"
)

type IssueSeverity string

const (
	SeverityCritical IssueSeverity = "critical"
	SeverityHigh     IssueSeverity = "high"
	SeverityMedium   IssueSeverity = "medium"
	SeverityLow      IssueSeverity = "low"
)

// =============================================================================
// Context Types
// =============================================================================

// EnrichedContext represents enriched context from multiple sources
type EnrichedContext struct {
	Docs             *ConsolidatedDocs
	RAGContext       []RAGDocument
	GitHistory       string
	JIRATickets      []JIRATicket
	Ownership        *OwnershipInfo
	SLODashboard     map[string]interface{}
	SecurityPolicies []SecurityPolicy
	Sources          map[string]ContextSource
	GeneratedAt      time.Time
}

// RAGDocument represents a retrieved document from vector DB
type RAGDocument struct {
	Content    string                 `json:"content"`
	Source     string                 `json:"source"`
	Score      float64                `json:"score"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// JIRATicket represents a JIRA ticket
type JIRATicket struct {
	Key         string   `json:"key"`
	Summary     string   `json:"summary"`
	Description string   `json:"description"`
	Type        string   `json:"type"`
	Status      string   `json:"status"`
	Labels      []string `json:"labels"`
}

// OwnershipInfo represents code ownership information
type OwnershipInfo struct {
	Team    string   `json:"team"`
	Owners  []string `json:"owners"`
	Slack   string   `json:"slack,omitempty"`
	Email   string   `json:"email,omitempty"`
}

// SecurityPolicy represents a security policy
type SecurityPolicy struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Rules       []string `json:"rules"`
}

// ContextSource represents a context source configuration
type ContextSource struct {
	Type   string                 `json:"type"`
	Weight float64                `json:"weight"`
	Config map[string]interface{} `json:"config,omitempty"`
}

// =============================================================================
// Output Types
// =============================================================================

// HLDOutput represents the final HLD document
type HLDOutput struct {
	Metadata        Metadata            `json:"metadata"`
	BusinessContext string              `json:"business_context,omitempty"`
	Architecture    string              `json:"architecture"`
	Security        string              `json:"security,omitempty"`
	Observability   string              `json:"observability,omitempty"`
	Requirements    string              `json:"requirements,omitempty"`
	Diagrams        map[string]string   `json:"diagrams,omitempty"`
	SLO             map[string]SLO     `json:"slo,omitempty"`
	Warnings        []string            `json:"warnings,omitempty"`
	RefinementLog   string              `json:"refinement_log,omitempty"`
	Markdown        string              `json:"markdown"`
	Rendered        map[string]string   `json:"rendered,omitempty"`
	ConsensusScore  float64             `json:"consensus_score"`
	FinalScore      float64             `json:"final_score"`
	RoundsCompleted int                 `json:"rounds_completed"`
	GeneratedAt     time.Time           `json:"generated_at"`
}

// Metadata represents HLD metadata
type Metadata struct {
	Version        string    `json:"version"`
	GenerationMode string    `json:"generation_mode"`
	ModuleName     string    `json:"module_name"`
	Author         string    `json:"author"`
	GeneratedAt    time.Time `json:"generated_at"`
	SourceCommit   string    `json:"source_commit"`
}

// =============================================================================
// Generation Mode
// =============================================================================

// Mode represents the generation mode
type Mode string

const (
	ModeBasic         Mode = "basic"
	ModeAdvanced      Mode = "advanced"
	ModeBusiness      Mode = "business"
	ModeCompliance    Mode = "compliance"
	ModeUltraAdvanced Mode = "ultra_advanced"
	ModeCustom        Mode = "custom"
)

// GenerateInput represents input for HLD generation
type GenerateInput struct {
	Docs           *ConsolidatedDocs
	Mode           Mode
	StylePreset    string
	CustomPromptPath string
}
