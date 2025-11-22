package diagrams

import (
	"strings"
	"time"
)

// generateTransformDiagram creates the Proto → Documentation transformation diagram
func (g *DiagramGenerator) generateTransformDiagram() GenerationResult {
	metadata := DiagramMetadata{
		Type:        DiagramTypeTransform,
		Title:       "Protocol Buffer to Documentation Transformation",
		Description: "How .proto constructs map to documentation model structures",
		Filename:    "proto-transform.md",
		GeneratedAt: time.Now(),
		MermaidType: "flowchart LR",
	}

	var sb strings.Builder
	sb.WriteString("flowchart LR\n")

	// Define styles
	sb.WriteString("    classDef protoNode fill:#e1f5ff,stroke:#01579b,stroke-width:2px\n")
	sb.WriteString("    classDef intermediate fill:#fff3e0,stroke:#e65100,stroke-width:2px\n")
	sb.WriteString("    classDef docNode fill:#e8f5e9,stroke:#1b5e20,stroke-width:2px\n")
	sb.WriteString("    classDef metadata fill:#f3e5f5,stroke:#4a148c,stroke-width:2px\n")
	sb.WriteString("\n")

	// Proto structures
	sb.WriteString("    subgraph ProtoDefinitions [Protocol Buffer Definitions]\n")
	sb.WriteString("        ProtoFile[📄 .proto File]:::protoNode\n")
	sb.WriteString("        Package[package statement]:::protoNode\n")
	sb.WriteString("        Service[service definition]:::protoNode\n")
	sb.WriteString("        RPC[rpc method]:::protoNode\n")
	sb.WriteString("        Message[message definition]:::protoNode\n")
	sb.WriteString("        Field[field definition]:::protoNode\n")
	sb.WriteString("        Enum[enum definition]:::protoNode\n")
	sb.WriteString("        EnumValue[enum value]:::protoNode\n")
	sb.WriteString("        Comments[// Comments]:::metadata\n")
	sb.WriteString("        Options[options & annotations]:::metadata\n")
	sb.WriteString("    end\n")
	sb.WriteString("\n")

	// Intermediate layer
	sb.WriteString("    subgraph DescriptorLayer [Descriptor Layer]\n")
	sb.WriteString("        FDS[FileDescriptorSet<br/>image.bin]:::intermediate\n")
	sb.WriteString("        FileDesc[FileDescriptor]:::intermediate\n")
	sb.WriteString("        ServiceDesc[ServiceDescriptor]:::intermediate\n")
	sb.WriteString("        MethodDesc[MethodDescriptor]:::intermediate\n")
	sb.WriteString("        MessageDesc[MessageDescriptor]:::intermediate\n")
	sb.WriteString("        FieldDesc[FieldDescriptor]:::intermediate\n")
	sb.WriteString("        EnumDesc[EnumDescriptor]:::intermediate\n")
	sb.WriteString("        SourceCodeInfo[SourceCodeInfo<br/>comment mapping]:::intermediate\n")
	sb.WriteString("    end\n")
	sb.WriteString("\n")

	// ProtoContext layer
	sb.WriteString("    subgraph ContextLayer [ProtoContext Layer]\n")
	sb.WriteString("        PC[ProtoContext]:::intermediate\n")
	sb.WriteString("        FileIndex[FileIndex<br/>fast lookup]:::intermediate\n")
	sb.WriteString("        CommentIndex[CommentsIndex<br/>extracted comments]:::intermediate\n")
	sb.WriteString("        Registry[protoregistry.Files<br/>reflection]:::intermediate\n")
	sb.WriteString("    end\n")
	sb.WriteString("\n")

	// Documentation model
	sb.WriteString("    subgraph DocumentationModel [API Documentation Model]\n")
	sb.WriteString("        ApiDocModel[📦 ApiDocModel]:::docNode\n")
	sb.WriteString("        DocModule[DocModule<br/>package grouping]:::docNode\n")
	sb.WriteString("        DocService[DocService]:::docNode\n")
	sb.WriteString("        DocMethod[DocMethod<br/>+ HTTP mapping]:::docNode\n")
	sb.WriteString("        DocMessage[DocMessage]:::docNode\n")
	sb.WriteString("        DocField[DocField<br/>+ oneof groups]:::docNode\n")
	sb.WriteString("        DocEnum[DocEnum]:::docNode\n")
	sb.WriteString("        DocEnumValue[DocEnumValue]:::docNode\n")
	sb.WriteString("        Statistics[Statistics<br/>aggregations]:::metadata\n")
	sb.WriteString("    end\n")
	sb.WriteString("\n")

	// Proto → Descriptor transformations
	sb.WriteString("    ProtoFile -->|buf build| FDS\n")
	sb.WriteString("    Package -->|compiled| FileDesc\n")
	sb.WriteString("    Service -->|compiled| ServiceDesc\n")
	sb.WriteString("    RPC -->|compiled| MethodDesc\n")
	sb.WriteString("    Message -->|compiled| MessageDesc\n")
	sb.WriteString("    Field -->|compiled| FieldDesc\n")
	sb.WriteString("    Enum -->|compiled| EnumDesc\n")
	sb.WriteString("    Comments -->|extracted| SourceCodeInfo\n")
	sb.WriteString("    Options -->|compiled| MethodDesc\n")
	sb.WriteString("\n")

	// Descriptor → Context transformations
	sb.WriteString("    FDS -->|LoadFromDescriptorSetFile| PC\n")
	sb.WriteString("    FileDesc -->|indexed| FileIndex\n")
	sb.WriteString("    SourceCodeInfo -->|processed| CommentIndex\n")
	sb.WriteString("    ServiceDesc -->|registered| Registry\n")
	sb.WriteString("\n")

	// Context → Doc Model transformations
	sb.WriteString("    PC -->|BuildApiDocModel| ApiDocModel\n")
	sb.WriteString("    FileIndex -->|grouped by package| DocModule\n")
	sb.WriteString("    ServiceDesc -->|mapped| DocService\n")
	sb.WriteString("    MethodDesc -->|mapped| DocMethod\n")
	sb.WriteString("    MessageDesc -->|mapped| DocMessage\n")
	sb.WriteString("    FieldDesc -->|mapped| DocField\n")
	sb.WriteString("    EnumDesc -->|mapped| DocEnum\n")
	sb.WriteString("    EnumValue -->|mapped| DocEnumValue\n")
	sb.WriteString("    CommentIndex -->|attached| DocService\n")
	sb.WriteString("    CommentIndex -->|attached| DocMessage\n")
	sb.WriteString("    CommentIndex -->|attached| DocField\n")
	sb.WriteString("    Registry -->|calculated| Statistics\n")
	sb.WriteString("\n")

	// Document model relationships
	sb.WriteString("    ApiDocModel --> DocModule\n")
	sb.WriteString("    ApiDocModel --> Statistics\n")
	sb.WriteString("    DocModule --> DocService\n")
	sb.WriteString("    DocModule --> DocMessage\n")
	sb.WriteString("    DocModule --> DocEnum\n")
	sb.WriteString("    DocService --> DocMethod\n")
	sb.WriteString("    DocMethod -.->|references| DocMessage\n")
	sb.WriteString("    DocMessage --> DocField\n")
	sb.WriteString("    DocField -.->|references| DocMessage\n")
	sb.WriteString("    DocField -.->|references| DocEnum\n")
	sb.WriteString("    DocEnum --> DocEnumValue\n")

	return GenerationResult{
		Metadata: metadata,
		Content:  sb.String(),
	}
}
