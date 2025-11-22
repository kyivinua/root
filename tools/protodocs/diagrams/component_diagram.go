package diagrams

import (
	"strings"
	"time"
)

// generateComponentDiagram creates the component interaction diagram
func (g *DiagramGenerator) generateComponentDiagram() GenerationResult {
	metadata := DiagramMetadata{
		Type:        DiagramTypeComponent,
		Title:       "Component Interaction Diagram",
		Description: "Major system components and their dependencies with external integrations",
		Filename:    "component-interaction.md",
		GeneratedAt: time.Now(),
		MermaidType: "graph TB",
	}

	var sb strings.Builder
	sb.WriteString("graph TB\n")

	// Define styles
	sb.WriteString("    classDef coreModule fill:#e1f5ff,stroke:#01579b,stroke-width:3px\n")
	sb.WriteString("    classDef adapter fill:#fff3e0,stroke:#e65100,stroke-width:2px\n")
	sb.WriteString("    classDef external fill:#f3e5f5,stroke:#4a148c,stroke-width:2px\n")
	sb.WriteString("    classDef storage fill:#e8f5e9,stroke:#1b5e20,stroke-width:2px\n")
	sb.WriteString("    classDef cli fill:#fff9c4,stroke:#f57f17,stroke-width:2px\n")
	sb.WriteString("\n")

	// CLI Layer
	sb.WriteString("    CLI[🖥️ CLI Commands<br/>proto-docs<br/>protodocs-enricher<br/>runtime-service]:::cli\n")
	sb.WriteString("\n")

	// Core modules
	sb.WriteString("    Pipeline[🔄 Pipeline Orchestrator<br/>Coordinates all stages]:::coreModule\n")
	sb.WriteString("    ProtoContext[📚 ProtoContext<br/>Descriptor handling<br/>Reflection + JSON]:::coreModule\n")
	sb.WriteString("    Enricher[🤖 Enricher<br/>LLM enrichment engine]:::coreModule\n")
	sb.WriteString("    NotificationMgr[📢 Notification Manager<br/>Event coordination]:::coreModule\n")
	sb.WriteString("    DiagramGen[📊 Diagram Generator<br/>Mermaid visualization]:::coreModule\n")
	sb.WriteString("\n")

	// Enricher components
	sb.WriteString("    PolicyEngine[🔒 Policy Engine<br/>Tenant permissions]:::adapter\n")
	sb.WriteString("    LLMClient[🤖 LLM Client<br/>gollm adapters]:::adapter\n")
	sb.WriteString("    RAGRetriever[🔍 RAG Retriever<br/>Weaviate + langchain]:::adapter\n")
	sb.WriteString("    Cache[💾 Cache<br/>Ristretto in-memory]:::adapter\n")
	sb.WriteString("    SafetyGuard[🛡️ Safety Guard<br/>Entropy + Judge]:::adapter\n")
	sb.WriteString("    Metrics[📊 Metrics<br/>Prometheus]:::adapter\n")
	sb.WriteString("    TraceSink[📋 Trace Sink<br/>Audit trail]:::adapter\n")
	sb.WriteString("    TemplateEngine[📝 Template Engine<br/>CoT prompts]:::adapter\n")
	sb.WriteString("    SmartStrategy[🧠 Smart Strategy<br/>Adaptive RAG]:::adapter\n")
	sb.WriteString("\n")

	// External systems
	sb.WriteString("    Buf[🔧 Buf CLI<br/>Lint, breaking, build]:::external\n")
	sb.WriteString("    Git[📦 Git<br/>Change detection]:::external\n")
	sb.WriteString("    Weaviate[🗄️ Weaviate<br/>Vector database]:::external\n")
	sb.WriteString("    Slack[💬 Slack<br/>Notifications]:::external\n")
	sb.WriteString("    LLMProviders[☁️ LLM Providers<br/>Anthropic/OpenAI/Ollama]:::external\n")
	sb.WriteString("    PrometheusServer[📈 Prometheus<br/>Metrics server]:::external\n")
	sb.WriteString("\n")

	// Storage
	sb.WriteString("    ProtoFiles[(📄 .proto Files)]:::storage\n")
	sb.WriteString("    Config[(⚙️ Configuration<br/>YAML)]:::storage\n")
	sb.WriteString("    Descriptors[(🗄️ Descriptors<br/>image.bin)]:::storage\n")
	sb.WriteString("    DocModel[(📦 API Doc Model<br/>JSON)]:::storage\n")
	sb.WriteString("    EnrichedModel[(✨ Enriched Model<br/>JSON)]:::storage\n")
	sb.WriteString("    Manifest[(📊 Manifest<br/>JSON)]:::storage\n")
	sb.WriteString("    Docs[(📚 Documentation<br/>Markdown/HTML)]:::storage\n")
	sb.WriteString("    Diagrams[(📊 Diagrams<br/>Mermaid)]:::storage\n")
	sb.WriteString("\n")

	// CLI connections
	sb.WriteString("    CLI --> Pipeline\n")
	sb.WriteString("    CLI --> Enricher\n")
	sb.WriteString("    CLI --> ProtoContext\n")
	sb.WriteString("\n")

	// Pipeline connections
	sb.WriteString("    Pipeline --> Config\n")
	sb.WriteString("    Pipeline --> Buf\n")
	sb.WriteString("    Pipeline --> Git\n")
	sb.WriteString("    Pipeline --> ProtoContext\n")
	sb.WriteString("    Pipeline --> Enricher\n")
	sb.WriteString("    Pipeline --> NotificationMgr\n")
	sb.WriteString("    Pipeline --> DiagramGen\n")
	sb.WriteString("\n")

	// ProtoContext connections
	sb.WriteString("    ProtoFiles --> Buf\n")
	sb.WriteString("    Buf --> Descriptors\n")
	sb.WriteString("    Descriptors --> ProtoContext\n")
	sb.WriteString("    ProtoContext --> DocModel\n")
	sb.WriteString("\n")

	// Enricher connections
	sb.WriteString("    DocModel --> Enricher\n")
	sb.WriteString("    Enricher --> PolicyEngine\n")
	sb.WriteString("    Enricher --> LLMClient\n")
	sb.WriteString("    Enricher --> RAGRetriever\n")
	sb.WriteString("    Enricher --> Cache\n")
	sb.WriteString("    Enricher --> SafetyGuard\n")
	sb.WriteString("    Enricher --> Metrics\n")
	sb.WriteString("    Enricher --> TraceSink\n")
	sb.WriteString("    Enricher --> TemplateEngine\n")
	sb.WriteString("    Enricher --> SmartStrategy\n")
	sb.WriteString("    Enricher --> EnrichedModel\n")
	sb.WriteString("    Enricher --> Manifest\n")
	sb.WriteString("\n")

	// External integrations
	sb.WriteString("    LLMClient --> LLMProviders\n")
	sb.WriteString("    RAGRetriever --> Weaviate\n")
	sb.WriteString("    NotificationMgr --> Slack\n")
	sb.WriteString("    Metrics --> PrometheusServer\n")
	sb.WriteString("\n")

	// Diagram generation
	sb.WriteString("    DiagramGen --> DocModel\n")
	sb.WriteString("    DiagramGen --> EnrichedModel\n")
	sb.WriteString("    DiagramGen --> Diagrams\n")
	sb.WriteString("\n")

	// Documentation output
	sb.WriteString("    DocModel --> Docs\n")
	sb.WriteString("    EnrichedModel --> Docs\n")

	return GenerationResult{
		Metadata: metadata,
		Content:  sb.String(),
	}
}
