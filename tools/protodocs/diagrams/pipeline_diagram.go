package diagrams

import (
	"strings"
	"time"
)

// generatePipelineDiagram creates the pipeline architecture diagram
func (g *DiagramGenerator) generatePipelineDiagram() GenerationResult {
	metadata := DiagramMetadata{
		Type:        DiagramTypePipeline,
		Title:       "ProtoDocs Pipeline Architecture",
		Description: "End-to-end documentation generation pipeline with all stages and decision points",
		Filename:    "pipeline-architecture.md",
		GeneratedAt: time.Now(),
		MermaidType: "flowchart LR",
	}

	var sb strings.Builder
	sb.WriteString("flowchart LR\n")

	// Define styles
	sb.WriteString("    classDef inputNode fill:#e1f5ff,stroke:#01579b,stroke-width:2px\n")
	sb.WriteString("    classDef processNode fill:#fff3e0,stroke:#e65100,stroke-width:2px\n")
	sb.WriteString("    classDef decisionNode fill:#f3e5f5,stroke:#4a148c,stroke-width:2px\n")
	sb.WriteString("    classDef outputNode fill:#e8f5e9,stroke:#1b5e20,stroke-width:2px\n")
	sb.WriteString("    classDef errorNode fill:#ffebee,stroke:#b71c1c,stroke-width:2px\n")
	sb.WriteString("\n")

	// Input
	sb.WriteString("    Start([Start Pipeline]):::inputNode\n")
	sb.WriteString("    ProtoFiles[📄 .proto Files]:::inputNode\n")
	sb.WriteString("    Config[⚙️ Configuration YAML]:::inputNode\n")
	sb.WriteString("\n")

	// Pipeline stages
	sb.WriteString("    Discovery[🔍 Discovery Stage<br/>Find changed .proto files]:::processNode\n")
	sb.WriteString("    Lint[✓ Lint Stage<br/>buf lint + comments check]:::processNode\n")
	sb.WriteString("    Breaking[⚠️ Breaking Check<br/>buf breaking vs main]:::processNode\n")
	sb.WriteString("    Build[🔨 Descriptor Build<br/>buf build → image.bin]:::processNode\n")
	sb.WriteString("    LoadContext[📥 Load ProtoContext<br/>Parse descriptors + comments]:::processNode\n")
	sb.WriteString("    BuildModel[🏗️ Build Doc Model<br/>Create ApiDocModel JSON]:::processNode\n")
	sb.WriteString("\n")

	// Conditional enrichment
	sb.WriteString("    EnrichDecision{Enrichment<br/>Enabled?}:::decisionNode\n")
	sb.WriteString("    Enrich[🤖 LLM Enrichment<br/>Policy + RAG + Safety]:::processNode\n")
	sb.WriteString("    Manifest[📊 Enrichment Manifest<br/>Statistics + Traces]:::outputNode\n")
	sb.WriteString("\n")

	// Documentation generation
	sb.WriteString("    GenDocs[📝 Generate Docs<br/>Markdown/HTML]:::processNode\n")
	sb.WriteString("    GenOpenAPI[📋 Generate OpenAPI<br/>Swagger/OpenAPI v3]:::processNode\n")
	sb.WriteString("    BuildSite[🌐 Build Site<br/>mkdocs/docusaurus]:::processNode\n")
	sb.WriteString("\n")

	// Notifications
	sb.WriteString("    NotifyStart[📢 Notify Start<br/>Slack notification]:::processNode\n")
	sb.WriteString("    NotifyComplete[📢 Notify Complete<br/>Slack + Release Notes]:::processNode\n")
	sb.WriteString("    NotifyFail[📢 Notify Failure<br/>Error details]:::errorNode\n")
	sb.WriteString("\n")

	// Breaking changes notification
	sb.WriteString("    BreakingDetected{Breaking<br/>Changes?}:::decisionNode\n")
	sb.WriteString("    NotifyBreaking[⚠️ Breaking Changes Alert]:::errorNode\n")
	sb.WriteString("\n")

	// Outputs
	sb.WriteString("    Descriptors[(🗄️ Descriptors<br/>image.bin)]:::outputNode\n")
	sb.WriteString("    DocModel[(📦 API Doc Model<br/>api-doc-model.json)]:::outputNode\n")
	sb.WriteString("    EnrichedModel[(✨ Enriched Model<br/>enriched.json)]:::outputNode\n")
	sb.WriteString("    Docs[(📚 Documentation<br/>Markdown/HTML)]:::outputNode\n")
	sb.WriteString("    OpenAPI[(📋 OpenAPI Specs<br/>swagger.json)]:::outputNode\n")
	sb.WriteString("    Site[(🌐 Static Site<br/>HTML)]:::outputNode\n")
	sb.WriteString("    End([✅ Complete]):::outputNode\n")
	sb.WriteString("    Error([❌ Failed]):::errorNode\n")
	sb.WriteString("\n")

	// Main flow
	sb.WriteString("    Start --> NotifyStart\n")
	sb.WriteString("    NotifyStart --> Discovery\n")
	sb.WriteString("    ProtoFiles --> Discovery\n")
	sb.WriteString("    Config --> Discovery\n")
	sb.WriteString("    Discovery --> Lint\n")
	sb.WriteString("    Lint --> Breaking\n")
	sb.WriteString("    Breaking --> BreakingDetected\n")
	sb.WriteString("    BreakingDetected -->|Yes| NotifyBreaking\n")
	sb.WriteString("    BreakingDetected -->|Continue| Build\n")
	sb.WriteString("    NotifyBreaking -.->|Continue| Build\n")
	sb.WriteString("    Build --> Descriptors\n")
	sb.WriteString("    Descriptors --> LoadContext\n")
	sb.WriteString("    LoadContext --> BuildModel\n")
	sb.WriteString("    BuildModel --> DocModel\n")
	sb.WriteString("    DocModel --> EnrichDecision\n")
	sb.WriteString("\n")

	// Enrichment branch
	sb.WriteString("    EnrichDecision -->|Yes| Enrich\n")
	sb.WriteString("    EnrichDecision -->|No| GenDocs\n")
	sb.WriteString("    Enrich --> EnrichedModel\n")
	sb.WriteString("    Enrich --> Manifest\n")
	sb.WriteString("    EnrichedModel --> GenDocs\n")
	sb.WriteString("\n")

	// Documentation generation
	sb.WriteString("    GenDocs --> Docs\n")
	sb.WriteString("    GenDocs --> GenOpenAPI\n")
	sb.WriteString("    GenOpenAPI --> OpenAPI\n")
	sb.WriteString("    Docs --> BuildSite\n")
	sb.WriteString("    BuildSite --> Site\n")
	sb.WriteString("    Site --> NotifyComplete\n")
	sb.WriteString("    NotifyComplete --> End\n")
	sb.WriteString("\n")

	// Error handling
	sb.WriteString("    Lint -.->|Errors| NotifyFail\n")
	sb.WriteString("    Breaking -.->|Errors| NotifyFail\n")
	sb.WriteString("    Build -.->|Errors| NotifyFail\n")
	sb.WriteString("    Enrich -.->|Errors| NotifyFail\n")
	sb.WriteString("    GenDocs -.->|Errors| NotifyFail\n")
	sb.WriteString("    NotifyFail --> Error\n")

	return GenerationResult{
		Metadata: metadata,
		Content:  sb.String(),
	}
}
