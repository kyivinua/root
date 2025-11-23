package diagrams

import (
	"fmt"
	"strings"
)

// EnhancedMermaidGenerator generates comprehensive Mermaid diagrams for gRPC services
type EnhancedMermaidGenerator struct {
	config *DiagramConfig
}

// NewEnhancedMermaidGenerator creates a new enhanced Mermaid diagram generator
func NewEnhancedMermaidGenerator(config *DiagramConfig) *EnhancedMermaidGenerator {
	if config == nil {
		config = DefaultDiagramConfig()
	}
	return &EnhancedMermaidGenerator{
		config: config,
	}
}

// GenerateCompleteDiagramSet generates all diagram types for a service
func (g *EnhancedMermaidGenerator) GenerateCompleteDiagramSet(service *DocService, messages []DocMessage, enums []DocEnum) map[string]string {
	diagrams := make(map[string]string)

	// Core diagrams
	diagrams["architecture"] = g.GenerateEnhancedArchitectureDiagram(service, messages)
	diagrams["sequence"] = g.GenerateComprehensiveSequenceDiagram(service)
	diagrams["class"] = g.GenerateEnhancedClassDiagram(messages)
	diagrams["erd"] = g.GenerateERDiagram(messages)

	// Advanced diagrams
	diagrams["state_machine"] = g.GenerateStateMachineDiagram(service)
	diagrams["flowchart"] = g.GenerateMethodFlowchart(service)
	diagrams["mindmap"] = g.GenerateServiceMindmap(service, messages)
	diagrams["c4_context"] = g.GenerateC4ContextDiagram(service)
	diagrams["c4_container"] = g.GenerateC4ContainerDiagram(service)
	diagrams["data_flow"] = g.GenerateDataFlowDiagram(service, messages)

	return diagrams
}

// GenerateEnhancedArchitectureDiagram generates a comprehensive architecture diagram
func (g *EnhancedMermaidGenerator) GenerateEnhancedArchitectureDiagram(service *DocService, messages []DocMessage) string {
	var sb strings.Builder

	g.writeThemeConfig(&sb)
	sb.WriteString("graph TB\n")

	// Define styles
	sb.WriteString("    classDef service fill:#4CAF50,stroke:#2E7D32,stroke-width:3px,color:#fff\n")
	sb.WriteString("    classDef unary fill:#2196F3,stroke:#1565C0,stroke-width:2px,color:#fff\n")
	sb.WriteString("    classDef stream fill:#FF9800,stroke:#E65100,stroke-width:2px,color:#fff\n")
	sb.WriteString("    classDef message fill:#9C27B0,stroke:#4A148C,stroke-width:2px,color:#fff\n")
	sb.WriteString("    classDef enum fill:#00BCD4,stroke:#006064,stroke-width:2px,color:#fff\n\n")

	// Service node
	serviceID := sanitizeName(service.Name)
	emoji := g.getServiceEmoji()
	sb.WriteString(fmt.Sprintf("    %s[\"%s %s<br/>%d methods\"]:::service\n\n",
		serviceID, emoji, service.Name, len(service.Methods)))

	// Group methods by streaming type
	unaryMethods := []DocMethod{}
	streamingMethods := []DocMethod{}

	for _, method := range service.Methods {
		if method.ClientStreaming || method.ServerStreaming {
			streamingMethods = append(streamingMethods, method)
		} else {
			unaryMethods = append(unaryMethods, method)
		}
	}

	// Create method nodes
	for _, method := range unaryMethods {
		methodID := sanitizeName(method.Name)
		icon := "🔵"
		sb.WriteString(fmt.Sprintf("    %s[\"%s %s\"]:::unary\n", methodID, icon, method.Name))
		sb.WriteString(fmt.Sprintf("    %s --> %s\n", serviceID, methodID))
	}

	for _, method := range streamingMethods {
		methodID := sanitizeName(method.Name)
		icon := g.getStreamingIcon(method.ClientStreaming, method.ServerStreaming)
		sb.WriteString(fmt.Sprintf("    %s[\"%s %s\"]:::stream\n", methodID, icon, method.Name))
		sb.WriteString(fmt.Sprintf("    %s --> %s\n", serviceID, methodID))
	}

	// Add message connections
	messageMap := make(map[string]bool)
	for _, method := range service.Methods {
		inputMsg := getShortName(method.InputType)
		outputMsg := getShortName(method.OutputType)

		if !messageMap[inputMsg] {
			sb.WriteString(fmt.Sprintf("    %s[\"📥 %s\"]:::message\n", sanitizeName(inputMsg), inputMsg))
			messageMap[inputMsg] = true
		}

		if !messageMap[outputMsg] {
			sb.WriteString(fmt.Sprintf("    %s[\"📤 %s\"]:::message\n", sanitizeName(outputMsg), outputMsg))
			messageMap[outputMsg] = true
		}

		methodID := sanitizeName(method.Name)
		sb.WriteString(fmt.Sprintf("    %s -.->|request| %s\n", sanitizeName(inputMsg), methodID))
		sb.WriteString(fmt.Sprintf("    %s -.->|response| %s\n", methodID, sanitizeName(outputMsg)))
	}

	return sb.String()
}

// GenerateComprehensiveSequenceDiagram generates detailed sequence diagrams
func (g *EnhancedMermaidGenerator) GenerateComprehensiveSequenceDiagram(service *DocService) string {
	var sb strings.Builder

	g.writeThemeConfig(&sb)
	sb.WriteString("sequenceDiagram\n")
	sb.WriteString("    autonumber\n")
	sb.WriteString("    participant C as Client\n")
	sb.WriteString("    participant A as API Gateway\n")
	sb.WriteString(fmt.Sprintf("    participant S as %s\n", service.Name))
	sb.WriteString("    participant DB as Database\n\n")

	for _, method := range service.Methods {
		sb.WriteString(fmt.Sprintf("    %% %s\n", method.Name))

		if method.ClientStreaming && method.ServerStreaming {
			// Bidirectional streaming
			sb.WriteString(fmt.Sprintf("    C->>+S: %s (bidirectional stream)\n", method.Name))
			sb.WriteString("    loop Stream Exchange\n")
			sb.WriteString(fmt.Sprintf("        C->>S: %s\n", getShortName(method.InputType)))
			sb.WriteString("        S->>DB: Process\n")
			sb.WriteString("        DB-->>S: Data\n")
			sb.WriteString(fmt.Sprintf("        S->>C: %s\n", getShortName(method.OutputType)))
			sb.WriteString("    end\n")
			sb.WriteString("    S-->>-C: Stream Complete\n\n")
		} else if method.ClientStreaming {
			// Client streaming
			sb.WriteString(fmt.Sprintf("    C->>+S: %s (client stream)\n", method.Name))
			sb.WriteString("    loop Upload Stream\n")
			sb.WriteString(fmt.Sprintf("        C->>S: %s\n", getShortName(method.InputType)))
			sb.WriteString("    end\n")
			sb.WriteString("    S->>DB: Process Batch\n")
			sb.WriteString("    DB-->>S: Result\n")
			sb.WriteString(fmt.Sprintf("    S-->>-C: %s\n\n", getShortName(method.OutputType)))
		} else if method.ServerStreaming {
			// Server streaming
			sb.WriteString(fmt.Sprintf("    C->>+S: %s\n", method.Name))
			sb.WriteString(fmt.Sprintf("    Note right of C: %s\n", getShortName(method.InputType)))
			sb.WriteString("    S->>DB: Query\n")
			sb.WriteString("    DB-->>S: Results\n")
			sb.WriteString("    loop Download Stream\n")
			sb.WriteString(fmt.Sprintf("        S->>C: %s\n", getShortName(method.OutputType)))
			sb.WriteString("    end\n")
			sb.WriteString("    S-->>-C: Stream Complete\n\n")
		} else {
			// Unary
			sb.WriteString(fmt.Sprintf("    C->>A: %s Request\n", method.Name))
			sb.WriteString(fmt.Sprintf("    A->>+S: %s\n", method.Name))
			sb.WriteString(fmt.Sprintf("    Note right of S: %s\n", getShortName(method.InputType)))
			sb.WriteString("    S->>DB: Query/Update\n")
			sb.WriteString("    DB-->>S: Result\n")
			sb.WriteString(fmt.Sprintf("    S-->>-A: %s\n", getShortName(method.OutputType)))
			sb.WriteString("    A-->>C: Response\n\n")
		}
	}

	return sb.String()
}

// GenerateEnhancedClassDiagram generates a UML class diagram with relationships
func (g *EnhancedMermaidGenerator) GenerateEnhancedClassDiagram(messages []DocMessage) string {
	var sb strings.Builder

	g.writeThemeConfig(&sb)
	sb.WriteString("classDiagram\n")

	// Add messages as classes
	for _, message := range messages {
		className := sanitizeName(message.Name)
		sb.WriteString(fmt.Sprintf("    class %s {\n", className))

		// Add fields
		for _, field := range message.Fields {
			fieldType := field.Type
			if field.TypeName != "" {
				fieldType = getShortName(field.TypeName)
			}

			modifier := ""
			if field.Label == "repeated" {
				modifier = "[]"
			}

			sb.WriteString(fmt.Sprintf("        +%s %s%s\n", field.Name, fieldType, modifier))
		}

		sb.WriteString("    }\n\n")
	}

	// Add relationships
	for _, message := range messages {
		for _, field := range message.Fields {
			if field.TypeName != "" && isMessageType(field.TypeName) {
				fromClass := sanitizeName(message.Name)
				toClass := sanitizeName(getShortName(field.TypeName))

				if field.Label == "repeated" {
					sb.WriteString(fmt.Sprintf("    %s \"1\" --> \"*\" %s : %s\n",
						fromClass, toClass, field.Name))
				} else {
					sb.WriteString(fmt.Sprintf("    %s --> %s : %s\n",
						fromClass, toClass, field.Name))
				}
			}
		}
	}

	return sb.String()
}

// GenerateERDiagram generates an Entity-Relationship diagram
func (g *EnhancedMermaidGenerator) GenerateERDiagram(messages []DocMessage) string {
	var sb strings.Builder

	g.writeThemeConfig(&sb)
	sb.WriteString("erDiagram\n")

	// Add entities
	for _, message := range messages {
		entityName := sanitizeName(message.Name)

		for _, field := range message.Fields {
			fieldType := field.Type
			if field.TypeName != "" {
				fieldType = getShortName(field.TypeName)
			}

			sb.WriteString(fmt.Sprintf("    %s {\n", entityName))
			sb.WriteString(fmt.Sprintf("        %s %s\n", fieldType, field.Name))
			sb.WriteString("    }\n")
		}
	}

	// Add relationships
	for _, message := range messages {
		for _, field := range message.Fields {
			if field.TypeName != "" && isMessageType(field.TypeName) {
				fromEntity := sanitizeName(message.Name)
				toEntity := sanitizeName(getShortName(field.TypeName))

				if field.Label == "repeated" {
					sb.WriteString(fmt.Sprintf("    %s ||--o{ %s : \"%s\"\n",
						fromEntity, toEntity, field.Name))
				} else {
					sb.WriteString(fmt.Sprintf("    %s ||--|| %s : \"%s\"\n",
						fromEntity, toEntity, field.Name))
				}
			}
		}
	}

	return sb.String()
}

// GenerateStateMachineDiagram generates a state machine diagram for service lifecycle
func (g *EnhancedMermaidGenerator) GenerateStateMachineDiagram(service *DocService) string {
	var sb strings.Builder

	g.writeThemeConfig(&sb)
	sb.WriteString("stateDiagram-v2\n")
	sb.WriteString("    [*] --> Idle\n\n")

	for _, method := range service.Methods {
		if method.ClientStreaming || method.ServerStreaming {
			sb.WriteString(fmt.Sprintf("    Idle --> Streaming : %s\n", method.Name))
			sb.WriteString("    Streaming --> Complete : Success\n")
			sb.WriteString("    Streaming --> Error : Failure\n")
		} else {
			sb.WriteString(fmt.Sprintf("    Idle --> Processing : %s\n", method.Name))
			sb.WriteString("    Processing --> Complete : Success\n")
			sb.WriteString("    Processing --> Error : Failure\n")
		}
	}

	sb.WriteString("\n    Complete --> Idle\n")
	sb.WriteString("    Error --> Idle : Retry\n")
	sb.WriteString("    Error --> [*] : Fatal\n")

	return sb.String()
}

// GenerateMethodFlowchart generates a flowchart for method execution
func (g *EnhancedMermaidGenerator) GenerateMethodFlowchart(service *DocService) string {
	var sb strings.Builder

	g.writeThemeConfig(&sb)
	sb.WriteString("flowchart TD\n")
	sb.WriteString("    Start([Client Request]) --> Auth{Authenticated?}\n")
	sb.WriteString("    Auth -->|No| AuthError[Return 401]\n")
	sb.WriteString("    Auth -->|Yes| RateLimit{Rate Limit OK?}\n")
	sb.WriteString("    RateLimit -->|No| RateLimitError[Return 429]\n")
	sb.WriteString("    RateLimit -->|Yes| Validate{Valid Request?}\n")
	sb.WriteString("    Validate -->|No| ValidationError[Return 400]\n")
	sb.WriteString("    Validate -->|Yes| SelectMethod{Which Method?}\n\n")

	for i, method := range service.Methods {
		sb.WriteString(fmt.Sprintf("    SelectMethod -->|%s| Method%d[%s]\n",
			method.Name, i, method.Name))
		sb.WriteString(fmt.Sprintf("    Method%d --> Execute%d{Execute}\n", i, i))
		sb.WriteString(fmt.Sprintf("    Execute%d -->|Success| Success%d[Return %s]\n",
			i, i, getShortName(method.OutputType)))
		sb.WriteString(fmt.Sprintf("    Execute%d -->|Error| HandleError%d[Handle Error]\n", i, i))
		sb.WriteString(fmt.Sprintf("    Success%d --> End\n", i))
		sb.WriteString(fmt.Sprintf("    HandleError%d --> End\n", i))
	}

	sb.WriteString("    AuthError --> End([Response])\n")
	sb.WriteString("    RateLimitError --> End\n")
	sb.WriteString("    ValidationError --> End\n")

	return sb.String()
}

// GenerateServiceMindmap generates a mindmap visualization of the service
func (g *EnhancedMermaidGenerator) GenerateServiceMindmap(service *DocService, messages []DocMessage) string {
	var sb strings.Builder

	sb.WriteString("mindmap\n")
	sb.WriteString(fmt.Sprintf("  root((%s))\n", service.Name))

	// Methods branch
	sb.WriteString("    Methods\n")
	for _, method := range service.Methods {
		icon := g.getStreamingIcon(method.ClientStreaming, method.ServerStreaming)
		sb.WriteString(fmt.Sprintf("      %s %s\n", icon, method.Name))
	}

	// Messages branch
	sb.WriteString("    Messages\n")
	messageCount := len(messages)
	if messageCount > 10 {
		messageCount = 10 // Limit for readability
	}
	for i := 0; i < messageCount; i++ {
		sb.WriteString(fmt.Sprintf("      📋 %s\n", messages[i].Name))
	}

	// Streaming types branch
	sb.WriteString("    Streaming Types\n")
	sb.WriteString("      Unary\n")
	sb.WriteString("      Client Stream\n")
	sb.WriteString("      Server Stream\n")
	sb.WriteString("      Bidirectional\n")

	return sb.String()
}

// GenerateC4ContextDiagram generates a C4 context diagram
func (g *EnhancedMermaidGenerator) GenerateC4ContextDiagram(service *DocService) string {
	var sb strings.Builder

	sb.WriteString("C4Context\n")
	sb.WriteString(fmt.Sprintf("    title System Context diagram for %s\n\n", service.Name))

	sb.WriteString("    Person(user, \"User\", \"End user of the system\")\n")
	sb.WriteString("    Person(admin, \"Administrator\", \"System administrator\")\n\n")

	sb.WriteString(fmt.Sprintf("    System(%s, \"%s\", \"gRPC service providing core functionality\")\n\n",
		sanitizeName(service.Name), service.Name))

	sb.WriteString("    System_Ext(database, \"Database\", \"Stores application data\")\n")
	sb.WriteString("    System_Ext(cache, \"Cache\", \"Redis cache for performance\")\n")
	sb.WriteString("    System_Ext(queue, \"Message Queue\", \"Async processing queue\")\n\n")

	sb.WriteString(fmt.Sprintf("    Rel(user, %s, \"Uses\", \"gRPC/HTTP\")\n", sanitizeName(service.Name)))
	sb.WriteString(fmt.Sprintf("    Rel(admin, %s, \"Manages\", \"gRPC/HTTP\")\n", sanitizeName(service.Name)))
	sb.WriteString(fmt.Sprintf("    Rel(%s, database, \"Reads/Writes\", \"SQL\")\n", sanitizeName(service.Name)))
	sb.WriteString(fmt.Sprintf("    Rel(%s, cache, \"Caches\", \"Redis Protocol\")\n", sanitizeName(service.Name)))
	sb.WriteString(fmt.Sprintf("    Rel(%s, queue, \"Publishes\", \"AMQP\")\n", sanitizeName(service.Name)))

	return sb.String()
}

// GenerateC4ContainerDiagram generates a C4 container diagram
func (g *EnhancedMermaidGenerator) GenerateC4ContainerDiagram(service *DocService) string {
	var sb strings.Builder

	sb.WriteString("C4Container\n")
	sb.WriteString(fmt.Sprintf("    title Container diagram for %s\n\n", service.Name))

	sb.WriteString("    Person(user, \"User\", \"End user\")\n\n")

	sb.WriteString(fmt.Sprintf("    Container_Boundary(%s_boundary, \"%s System\") {\n",
		sanitizeName(service.Name), service.Name))
	sb.WriteString("        Container(api, \"API Gateway\", \"Envoy/NGINX\", \"Routes requests\")\n")
	sb.WriteString(fmt.Sprintf("        Container(service, \"%s\", \"Go/gRPC\", \"Core business logic\")\n", service.Name))
	sb.WriteString("        Container(worker, \"Background Worker\", \"Go\", \"Async processing\")\n")
	sb.WriteString("        ContainerDb(db, \"Database\", \"PostgreSQL\", \"Stores data\")\n")
	sb.WriteString("        ContainerDb(cache, \"Cache\", \"Redis\", \"Performance cache\")\n")
	sb.WriteString("    }\n\n")

	sb.WriteString("    Rel(user, api, \"Uses\", \"HTTPS/gRPC\")\n")
	sb.WriteString("    Rel(api, service, \"Routes to\", \"gRPC\")\n")
	sb.WriteString("    Rel(service, db, \"Reads/Writes\", \"SQL\")\n")
	sb.WriteString("    Rel(service, cache, \"Caches\", \"Redis\")\n")
	sb.WriteString("    Rel(service, worker, \"Delegates to\", \"Queue\")\n")
	sb.WriteString("    Rel(worker, db, \"Updates\", \"SQL\")\n")

	return sb.String()
}

// GenerateDataFlowDiagram generates a data flow diagram
func (g *EnhancedMermaidGenerator) GenerateDataFlowDiagram(service *DocService, messages []DocMessage) string {
	var sb strings.Builder

	g.writeThemeConfig(&sb)
	sb.WriteString("graph LR\n")

	// Define styles
	sb.WriteString("    classDef input fill:#4CAF50,stroke:#2E7D32,color:#fff\n")
	sb.WriteString("    classDef process fill:#2196F3,stroke:#1565C0,color:#fff\n")
	sb.WriteString("    classDef output fill:#FF5722,stroke:#BF360C,color:#fff\n")
	sb.WriteString("    classDef storage fill:#9C27B0,stroke:#4A148C,color:#fff\n\n")

	// Create data flow for each method
	for _, method := range service.Methods {
		inputID := sanitizeName(getShortName(method.InputType))
		methodID := sanitizeName(method.Name)
		outputID := sanitizeName(getShortName(method.OutputType))

		sb.WriteString(fmt.Sprintf("    %s[%s]:::input\n", inputID, getShortName(method.InputType)))
		sb.WriteString(fmt.Sprintf("    %s{{%s}}:::process\n", methodID, method.Name))
		sb.WriteString(fmt.Sprintf("    %s[%s]:::output\n", outputID, getShortName(method.OutputType)))

		sb.WriteString(fmt.Sprintf("    %s --> %s\n", inputID, methodID))
		sb.WriteString(fmt.Sprintf("    %s --> %s\n", methodID, outputID))
	}

	return sb.String()
}

// Helper methods

func (g *EnhancedMermaidGenerator) writeThemeConfig(sb *strings.Builder) {
	if g.config.Theme != "default" && g.config.Theme != "" {
		sb.WriteString(fmt.Sprintf("%%{init: {'theme':'%s'}}%%\n", g.config.Theme))
	}
}

func (g *EnhancedMermaidGenerator) getServiceEmoji() string {
	return "🔧"
}

func (g *EnhancedMermaidGenerator) getStreamingIcon(clientStreaming, serverStreaming bool) string {
	if clientStreaming && serverStreaming {
		return "↔️"
	} else if clientStreaming {
		return "↑"
	} else if serverStreaming {
		return "↓"
	}
	return "🔵"
}
