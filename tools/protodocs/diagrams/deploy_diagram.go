package diagrams

import (
	"strings"
	"time"
)

// generateDeployDiagram creates the deployment architecture diagram
func (g *DiagramGenerator) generateDeployDiagram() GenerationResult {
	metadata := DiagramMetadata{
		Type:        DiagramTypeDeploy,
		Title:       "Deployment Architecture",
		Description: "Runtime deployment showing CLI tools, configuration, inputs, outputs, and external services",
		Filename:    "deployment-architecture.md",
		GeneratedAt: time.Now(),
		MermaidType: "graph LR",
	}

	var sb strings.Builder
	sb.WriteString("graph LR\n")

	// Define styles
	sb.WriteString("    classDef cli fill:#fff9c4,stroke:#f57f17,stroke-width:3px\n")
	sb.WriteString("    classDef config fill:#e1f5ff,stroke:#01579b,stroke-width:2px\n")
	sb.WriteString("    classDef input fill:#f3e5f5,stroke:#4a148c,stroke-width:2px\n")
	sb.WriteString("    classDef output fill:#e8f5e9,stroke:#1b5e20,stroke-width:2px\n")
	sb.WriteString("    classDef external fill:#ffebee,stroke:#b71c1c,stroke-width:2px\n")
	sb.WriteString("    classDef artifact fill:#fff3e0,stroke:#e65100,stroke-width:2px\n")
	sb.WriteString("\n")

	// CLI tools
	sb.WriteString("    subgraph CLITools [CLI Tools]\n")
	sb.WriteString("        ProtoDocs[🖥️ proto-docs<br/>Main pipeline CLI]:::cli\n")
	sb.WriteString("        Enricher[🖥️ protodocs-enricher<br/>Standalone enricher]:::cli\n")
	sb.WriteString("        RuntimeService[🖥️ runtime-service<br/>ProtoContext service]:::cli\n")
	sb.WriteString("    end\n")
	sb.WriteString("\n")

	// Configuration files
	sb.WriteString("    subgraph Configuration [Configuration Files]\n")
	sb.WriteString("        ProtoDocsConfig[⚙️ proto-docs.config.yaml<br/>Pipeline settings]:::config\n")
	sb.WriteString("        EnricherConfig[⚙️ enricher.config.yaml<br/>LLM settings]:::config\n")
	sb.WriteString("        BufConfig[⚙️ buf.yaml<br/>Protobuf settings]:::config\n")
	sb.WriteString("        MkDocsConfig[⚙️ mkdocs.yml<br/>Site generation]:::config\n")
	sb.WriteString("        EnvVars[🔐 Environment Variables<br/>SLACK_WEBHOOK_URL<br/>LLM_API_KEY<br/>WEAVIATE_URL]:::config\n")
	sb.WriteString("    end\n")
	sb.WriteString("\n")

	// Input files
	sb.WriteString("    subgraph Inputs [Input Files]\n")
	sb.WriteString("        ProtoFiles[📄 .proto Files<br/>Protocol Buffer definitions]:::input\n")
	sb.WriteString("        GitRepo[📦 Git Repository<br/>Change detection]:::input\n")
	sb.WriteString("    end\n")
	sb.WriteString("\n")

	// Generated artifacts
	sb.WriteString("    subgraph Artifacts [Generated Artifacts]\n")
	sb.WriteString("        Descriptors[🗄️ api-docs/descriptors/<br/>image.bin]:::artifact\n")
	sb.WriteString("        DocModel[📦 api-docs/model/<br/>api-doc-model.json]:::artifact\n")
	sb.WriteString("        EnrichedModel[✨ api-docs/model/<br/>enriched.json]:::artifact\n")
	sb.WriteString("        Manifest[📊 api-docs/<br/>enrichment-manifest.json]:::artifact\n")
	sb.WriteString("        Traces[📋 api-docs/traces/<br/>enrichment-traces.jsonl]:::artifact\n")
	sb.WriteString("    end\n")
	sb.WriteString("\n")

	// Output files
	sb.WriteString("    subgraph Outputs [Output Files]\n")
	sb.WriteString("        Markdown[📝 api-docs/proto-docs/<br/>Markdown documentation]:::output\n")
	sb.WriteString("        OpenAPI[📋 api-docs/openapi/<br/>OpenAPI specs]:::output\n")
	sb.WriteString("        Diagrams[📊 api-docs/diagrams/<br/>Mermaid diagrams]:::output\n")
	sb.WriteString("        Site[🌐 api-docs/site/<br/>Static HTML site]:::output\n")
	sb.WriteString("    end\n")
	sb.WriteString("\n")

	// External services
	sb.WriteString("    subgraph ExternalServices [External Services]\n")
	sb.WriteString("        Buf[🔧 Buf CLI<br/>buf.build tooling]:::external\n")
	sb.WriteString("        Git[📦 Git<br/>Version control]:::external\n")
	sb.WriteString("        Anthropic[🤖 Anthropic<br/>Claude API]:::external\n")
	sb.WriteString("        OpenAI[🤖 OpenAI<br/>GPT API]:::external\n")
	sb.WriteString("        Ollama[🤖 Ollama<br/>Local LLM]:::external\n")
	sb.WriteString("        Weaviate[🗄️ Weaviate<br/>Vector database]:::external\n")
	sb.WriteString("        Slack[💬 Slack<br/>Notifications]:::external\n")
	sb.WriteString("        Prometheus[📈 Prometheus<br/>Metrics]:::external\n")
	sb.WriteString("    end\n")
	sb.WriteString("\n")

	// CLI connections to config
	sb.WriteString("    ProtoDocsConfig --> ProtoDocs\n")
	sb.WriteString("    EnricherConfig --> Enricher\n")
	sb.WriteString("    BufConfig --> ProtoDocs\n")
	sb.WriteString("    MkDocsConfig --> ProtoDocs\n")
	sb.WriteString("    EnvVars --> ProtoDocs\n")
	sb.WriteString("    EnvVars --> Enricher\n")
	sb.WriteString("\n")

	// CLI connections to inputs
	sb.WriteString("    ProtoFiles --> ProtoDocs\n")
	sb.WriteString("    GitRepo --> ProtoDocs\n")
	sb.WriteString("\n")

	// CLI to external services
	sb.WriteString("    ProtoDocs --> Buf\n")
	sb.WriteString("    ProtoDocs --> Git\n")
	sb.WriteString("    ProtoDocs --> Slack\n")
	sb.WriteString("    Enricher --> Anthropic\n")
	sb.WriteString("    Enricher --> OpenAI\n")
	sb.WriteString("    Enricher --> Ollama\n")
	sb.WriteString("    Enricher --> Weaviate\n")
	sb.WriteString("    Enricher --> Prometheus\n")
	sb.WriteString("\n")

	// CLI to artifacts
	sb.WriteString("    ProtoDocs --> Descriptors\n")
	sb.WriteString("    ProtoDocs --> DocModel\n")
	sb.WriteString("    Enricher --> EnrichedModel\n")
	sb.WriteString("    Enricher --> Manifest\n")
	sb.WriteString("    Enricher --> Traces\n")
	sb.WriteString("\n")

	// Artifacts to outputs
	sb.WriteString("    DocModel --> ProtoDocs\n")
	sb.WriteString("    EnrichedModel --> ProtoDocs\n")
	sb.WriteString("    ProtoDocs --> Markdown\n")
	sb.WriteString("    ProtoDocs --> OpenAPI\n")
	sb.WriteString("    ProtoDocs --> Diagrams\n")
	sb.WriteString("    ProtoDocs --> Site\n")
	sb.WriteString("\n")

	// Enricher standalone mode
	sb.WriteString("    DocModel -.->|standalone mode| Enricher\n")

	return GenerationResult{
		Metadata: metadata,
		Content:  sb.String(),
	}
}
