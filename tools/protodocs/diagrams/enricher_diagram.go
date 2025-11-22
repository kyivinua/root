package diagrams

import (
	"strings"
	"time"
)

// generateEnricherDiagram creates the enricher orchestration diagram
func (g *DiagramGenerator) generateEnricherDiagram() GenerationResult {
	metadata := DiagramMetadata{
		Type:        DiagramTypeEnricher,
		Title:       "Enricher Orchestration Flow",
		Description: "LLM-based enrichment pipeline with policy enforcement, RAG, safety guards, and caching",
		Filename:    "enricher-orchestration.md",
		GeneratedAt: time.Now(),
		MermaidType: "flowchart TD",
	}

	var sb strings.Builder
	sb.WriteString("flowchart TD\n")

	// Define styles
	sb.WriteString("    classDef inputNode fill:#e1f5ff,stroke:#01579b,stroke-width:2px\n")
	sb.WriteString("    classDef processNode fill:#fff3e0,stroke:#e65100,stroke-width:2px\n")
	sb.WriteString("    classDef decisionNode fill:#f3e5f5,stroke:#4a148c,stroke-width:2px\n")
	sb.WriteString("    classDef safetyNode fill:#ffebee,stroke:#b71c1c,stroke-width:2px\n")
	sb.WriteString("    classDef successNode fill:#e8f5e9,stroke:#1b5e20,stroke-width:2px\n")
	sb.WriteString("    classDef cacheNode fill:#fff9c4,stroke:#f57f17,stroke-width:2px\n")
	sb.WriteString("\n")

	// Input
	sb.WriteString("    Start([Enrichment Start]):::inputNode\n")
	sb.WriteString("    Model[📦 API Doc Model<br/>Services, Messages, Enums]:::inputNode\n")
	sb.WriteString("    Tenant[👤 Tenant ID]:::inputNode\n")
	sb.WriteString("\n")

	// Target iteration
	sb.WriteString("    IterateTargets[🔄 Iterate Enrichment Targets<br/>Services → Methods → Fields]:::processNode\n")
	sb.WriteString("    Target[🎯 Current Target<br/>Service/Method/Message/Field]:::processNode\n")
	sb.WriteString("\n")

	// Policy check
	sb.WriteString("    PolicyEngine{🔒 Policy Engine<br/>Tenant Permissions}:::decisionNode\n")
	sb.WriteString("    PolicyDenied[❌ Denied<br/>Record in manifest]:::safetyNode\n")
	sb.WriteString("\n")

	// Cache check
	sb.WriteString("    CacheCheck{💾 Cache Check<br/>Ristretto Cache}:::decisionNode\n")
	sb.WriteString("    CacheHit[✅ Cache Hit<br/>Return cached result]:::cacheNode\n")
	sb.WriteString("\n")

	// Smart strategy
	sb.WriteString("    SmartStrategy{🧠 Smart Strategy<br/>Use RAG?}:::decisionNode\n")
	sb.WriteString("    RAGRetrieval[🔍 RAG Retrieval<br/>Query Weaviate<br/>Get top-K docs]:::processNode\n")
	sb.WriteString("\n")

	// Prompt building
	sb.WriteString("    BuildPrompt[📝 Build Prompt<br/>CoT Template Engine]:::processNode\n")
	sb.WriteString("    PromptContext[Context:<br/>- Current docs<br/>- RAG results<br/>- Target metadata]:::processNode\n")
	sb.WriteString("\n")

	// LLM call
	sb.WriteString("    LLMCall[🤖 LLM Client<br/>gollm library<br/>Anthropic/OpenAI/Ollama]:::processNode\n")
	sb.WriteString("    LLMResponse[Generated Content]:::processNode\n")
	sb.WriteString("\n")

	// Safety validation
	sb.WriteString("    SafetyGuard[🛡️ Safety Guard]:::safetyNode\n")
	sb.WriteString("    PIICheck{PII/PCI<br/>Detection?}:::safetyNode\n")
	sb.WriteString("    EntropyCheck{Semantic<br/>Entropy?}:::safetyNode\n")
	sb.WriteString("    JudgeCheck{LLM-as-Judge<br/>Faithfulness?}:::safetyNode\n")
	sb.WriteString("    SafetyFailed[⚠️ Safety Failed<br/>Use original docs]:::safetyNode\n")
	sb.WriteString("\n")

	// Recording
	sb.WriteString("    RecordMetrics[📊 Record Metrics<br/>Tokens, cost, duration]:::processNode\n")
	sb.WriteString("    RecordTrace[📋 Record Trace<br/>Full audit trail]:::processNode\n")
	sb.WriteString("    CacheSet[💾 Cache Set<br/>Store result]:::cacheNode\n")
	sb.WriteString("\n")

	// Loop control
	sb.WriteString("    MoreTargets{More<br/>Targets?}:::decisionNode\n")
	sb.WriteString("\n")

	// Outputs
	sb.WriteString("    Manifest[📊 Enrichment Manifest<br/>Statistics + Failures]:::successNode\n")
	sb.WriteString("    EnrichedModel[✨ Enriched Model<br/>Updated documentation]:::successNode\n")
	sb.WriteString("    End([Complete]):::successNode\n")
	sb.WriteString("\n")

	// Flow connections
	sb.WriteString("    Start --> Model\n")
	sb.WriteString("    Start --> Tenant\n")
	sb.WriteString("    Model --> IterateTargets\n")
	sb.WriteString("    Tenant --> IterateTargets\n")
	sb.WriteString("    IterateTargets --> Target\n")
	sb.WriteString("\n")

	// Policy branch
	sb.WriteString("    Target --> PolicyEngine\n")
	sb.WriteString("    PolicyEngine -->|Denied| PolicyDenied\n")
	sb.WriteString("    PolicyEngine -->|Allowed| CacheCheck\n")
	sb.WriteString("    PolicyDenied --> MoreTargets\n")
	sb.WriteString("\n")

	// Cache branch
	sb.WriteString("    CacheCheck -->|Hit| CacheHit\n")
	sb.WriteString("    CacheCheck -->|Miss| SmartStrategy\n")
	sb.WriteString("    CacheHit --> RecordMetrics\n")
	sb.WriteString("\n")

	// RAG strategy branch
	sb.WriteString("    SmartStrategy -->|Use RAG| RAGRetrieval\n")
	sb.WriteString("    SmartStrategy -->|No RAG| BuildPrompt\n")
	sb.WriteString("    RAGRetrieval --> PromptContext\n")
	sb.WriteString("    PromptContext --> BuildPrompt\n")
	sb.WriteString("\n")

	// LLM processing
	sb.WriteString("    BuildPrompt --> LLMCall\n")
	sb.WriteString("    LLMCall --> LLMResponse\n")
	sb.WriteString("    LLMResponse --> SafetyGuard\n")
	sb.WriteString("\n")

	// Safety checks
	sb.WriteString("    SafetyGuard --> PIICheck\n")
	sb.WriteString("    PIICheck -->|Detected| SafetyFailed\n")
	sb.WriteString("    PIICheck -->|Pass| EntropyCheck\n")
	sb.WriteString("    EntropyCheck -->|High| SafetyFailed\n")
	sb.WriteString("    EntropyCheck -->|Pass| JudgeCheck\n")
	sb.WriteString("    JudgeCheck -->|Failed| SafetyFailed\n")
	sb.WriteString("    JudgeCheck -->|Pass| RecordMetrics\n")
	sb.WriteString("    SafetyFailed --> RecordMetrics\n")
	sb.WriteString("\n")

	// Recording and caching
	sb.WriteString("    RecordMetrics --> RecordTrace\n")
	sb.WriteString("    RecordTrace --> CacheSet\n")
	sb.WriteString("    CacheSet --> MoreTargets\n")
	sb.WriteString("\n")

	// Loop control
	sb.WriteString("    MoreTargets -->|Yes| Target\n")
	sb.WriteString("    MoreTargets -->|No| Manifest\n")
	sb.WriteString("    Manifest --> EnrichedModel\n")
	sb.WriteString("    EnrichedModel --> End\n")

	return GenerationResult{
		Metadata: metadata,
		Content:  sb.String(),
	}
}
