package diagrams

import (
	"fmt"
	"strings"
	"time"
)

// generateDataModelDiagram creates an ER diagram of the API documentation model structure
func (g *DiagramGenerator) generateDataModelDiagram(model *ApiDocModel) GenerationResult {
	metadata := DiagramMetadata{
		Type:        DiagramTypeDataModel,
		Title:       "API Documentation Data Model",
		Description: fmt.Sprintf("Entity-relationship diagram showing the structure of the documentation model (%d modules)", len(model.Modules)),
		Filename:    "data-model-structure.md",
		GeneratedAt: time.Now(),
		MermaidType: "erDiagram",
	}

	var sb strings.Builder
	sb.WriteString("erDiagram\n")

	// ApiDocModel
	sb.WriteString("    ApiDocModel ||--o{ DocModule : contains\n")
	sb.WriteString("    ApiDocModel {\n")
	sb.WriteString("        time GeneratedAt\n")
	sb.WriteString("        string SourceCommit\n")
	sb.WriteString("        map Statistics\n")
	sb.WriteString("        map Tools\n")
	sb.WriteString("    }\n")
	sb.WriteString("\n")

	// DocModule
	sb.WriteString("    DocModule ||--o{ DocService : contains\n")
	sb.WriteString("    DocModule ||--o{ DocMessage : contains\n")
	sb.WriteString("    DocModule ||--o{ DocEnum : contains\n")
	sb.WriteString("    DocModule {\n")
	sb.WriteString("        string Name\n")
	sb.WriteString("        string Package\n")
	sb.WriteString("        string Description\n")
	sb.WriteString("        string FilePath\n")
	sb.WriteString("    }\n")
	sb.WriteString("\n")

	// DocService
	sb.WriteString("    DocService ||--o{ DocMethod : contains\n")
	sb.WriteString("    DocService {\n")
	sb.WriteString("        string Name\n")
	sb.WriteString("        string FullName\n")
	sb.WriteString("        string Description\n")
	sb.WriteString("        string Visibility\n")
	sb.WriteString("    }\n")
	sb.WriteString("\n")

	// DocMethod
	sb.WriteString("    DocMethod }o--|| DocMessage : \"input type\"\n")
	sb.WriteString("    DocMethod }o--|| DocMessage : \"output type\"\n")
	sb.WriteString("    DocMethod ||--o{ HTTPMethodInfo : \"has\"\n")
	sb.WriteString("    DocMethod {\n")
	sb.WriteString("        string Name\n")
	sb.WriteString("        string FullName\n")
	sb.WriteString("        string Description\n")
	sb.WriteString("        string InputType\n")
	sb.WriteString("        string OutputType\n")
	sb.WriteString("        bool ClientStreaming\n")
	sb.WriteString("        bool ServerStreaming\n")
	sb.WriteString("        string Visibility\n")
	sb.WriteString("    }\n")
	sb.WriteString("\n")

	// HTTPMethodInfo
	sb.WriteString("    HTTPMethodInfo {\n")
	sb.WriteString("        string Method\n")
	sb.WriteString("        string Path\n")
	sb.WriteString("    }\n")
	sb.WriteString("\n")

	// DocMessage
	sb.WriteString("    DocMessage ||--o{ DocField : contains\n")
	sb.WriteString("    DocMessage {\n")
	sb.WriteString("        string Name\n")
	sb.WriteString("        string FullName\n")
	sb.WriteString("        string Description\n")
	sb.WriteString("        string Visibility\n")
	sb.WriteString("    }\n")
	sb.WriteString("\n")

	// DocField
	sb.WriteString("    DocField }o--o| DocMessage : \"references (type)\"\n")
	sb.WriteString("    DocField }o--o| DocEnum : \"references (type)\"\n")
	sb.WriteString("    DocField {\n")
	sb.WriteString("        string Name\n")
	sb.WriteString("        int Number\n")
	sb.WriteString("        string Type\n")
	sb.WriteString("        string TypeName\n")
	sb.WriteString("        string Label\n")
	sb.WriteString("        string Description\n")
	sb.WriteString("        string OneofGroup\n")
	sb.WriteString("    }\n")
	sb.WriteString("\n")

	// DocEnum
	sb.WriteString("    DocEnum ||--o{ DocEnumValue : contains\n")
	sb.WriteString("    DocEnum {\n")
	sb.WriteString("        string Name\n")
	sb.WriteString("        string FullName\n")
	sb.WriteString("        string Description\n")
	sb.WriteString("        string Visibility\n")
	sb.WriteString("    }\n")
	sb.WriteString("\n")

	// DocEnumValue
	sb.WriteString("    DocEnumValue {\n")
	sb.WriteString("        string Name\n")
	sb.WriteString("        int Number\n")
	sb.WriteString("        string Description\n")
	sb.WriteString("    }\n")

	return GenerationResult{
		Metadata: metadata,
		Content:  sb.String(),
	}
}
