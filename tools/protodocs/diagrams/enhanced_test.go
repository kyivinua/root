package diagrams

import (
	"testing"
)

// TestGraphMLGeneration tests GraphML generation
func TestGraphMLGeneration(t *testing.T) {
	// Create test service
	service := &DocService{
		Name:        "UserService",
		FullName:    "users.v1.UserService",
		Description: "User management service",
		Methods: []DocMethod{
			{
				Name:            "CreateUser",
				FullName:        "users.v1.UserService.CreateUser",
				Description:     "Creates a new user",
				InputType:       "users.v1.CreateUserRequest",
				OutputType:      "users.v1.CreateUserResponse",
				ClientStreaming: false,
				ServerStreaming: false,
			},
			{
				Name:            "ListUsers",
				FullName:        "users.v1.UserService.ListUsers",
				Description:     "Lists users",
				InputType:       "users.v1.ListUsersRequest",
				OutputType:      "users.v1.User",
				ClientStreaming: false,
				ServerStreaming: true,
			},
		},
	}

	// Create test messages
	messages := []DocMessage{
		{
			Name:        "CreateUserRequest",
			FullName:    "users.v1.CreateUserRequest",
			Description: "Request to create a user",
			Fields: []DocField{
				{Name: "username", Type: "string", Label: "optional"},
				{Name: "email", Type: "string", Label: "optional"},
			},
		},
		{
			Name:        "CreateUserResponse",
			FullName:    "users.v1.CreateUserResponse",
			Description: "Response with created user",
			Fields: []DocField{
				{Name: "user", Type: "User", TypeName: "users.v1.User", Label: "optional"},
			},
		},
		{
			Name:        "User",
			FullName:    "users.v1.User",
			Description: "User model",
			Fields: []DocField{
				{Name: "id", Type: "string", Label: "optional"},
				{Name: "username", Type: "string", Label: "optional"},
				{Name: "email", Type: "string", Label: "optional"},
			},
		},
	}

	// Test GraphML generation
	generator := NewServiceGraphMLGenerator()
	graphml, err := generator.GenerateServiceGraphML(service, messages)
	if err != nil {
		t.Fatalf("Failed to generate GraphML: %v", err)
	}

	if graphml == nil {
		t.Fatal("GraphML is nil")
	}

	// Convert to XML
	xml, err := graphml.ToXML()
	if err != nil {
		t.Fatalf("Failed to convert GraphML to XML: %v", err)
	}

	if len(xml) == 0 {
		t.Fatal("Generated XML is empty")
	}

	t.Logf("Generated GraphML (%d bytes):\n%s", len(xml), xml)

	// Validate XML structure
	if !containsString(xml, "<graphml") {
		t.Error("XML missing graphml root element")
	}

	if !containsString(xml, "<node") {
		t.Error("XML missing node elements")
	}

	if !containsString(xml, "<edge") {
		t.Error("XML missing edge elements")
	}

	if !containsString(xml, "UserService") {
		t.Error("XML missing service name")
	}
}

// TestEnhancedMermaidDiagrams tests enhanced Mermaid diagram generation
func TestEnhancedMermaidDiagrams(t *testing.T) {
	// Create test data
	service := &DocService{
		Name:        "PaymentService",
		FullName:    "payments.v1.PaymentService",
		Description: "Payment processing service",
		Methods: []DocMethod{
			{
				Name:            "ProcessPayment",
				FullName:        "payments.v1.PaymentService.ProcessPayment",
				Description:     "Process a payment",
				InputType:       "payments.v1.PaymentRequest",
				OutputType:      "payments.v1.PaymentResponse",
				ClientStreaming: false,
				ServerStreaming: false,
			},
			{
				Name:            "StreamTransactions",
				FullName:        "payments.v1.PaymentService.StreamTransactions",
				Description:     "Stream transactions",
				InputType:       "payments.v1.StreamRequest",
				OutputType:      "payments.v1.Transaction",
				ClientStreaming: false,
				ServerStreaming: true,
			},
		},
	}

	messages := []DocMessage{
		{
			Name:        "PaymentRequest",
			FullName:    "payments.v1.PaymentRequest",
			Description: "Payment request",
			Fields: []DocField{
				{Name: "amount", Type: "double", Label: "optional"},
				{Name: "currency", Type: "string", Label: "optional"},
			},
		},
		{
			Name:        "PaymentResponse",
			FullName:    "payments.v1.PaymentResponse",
			Description: "Payment response",
			Fields: []DocField{
				{Name: "transaction_id", Type: "string", Label: "optional"},
				{Name: "status", Type: "string", Label: "optional"},
			},
		},
	}

	config := DefaultDiagramConfig()
	generator := NewEnhancedMermaidGenerator(config)

	// Test architecture diagram
	t.Run("ArchitectureDiagram", func(t *testing.T) {
		diagram := generator.GenerateEnhancedArchitectureDiagram(service, messages)
		if len(diagram) == 0 {
			t.Fatal("Architecture diagram is empty")
		}
		if !containsString(diagram, "graph TB") {
			t.Error("Architecture diagram missing graph declaration")
		}
		if !containsString(diagram, "PaymentService") {
			t.Error("Architecture diagram missing service name")
		}
		t.Logf("Architecture Diagram:\n%s", diagram)
	})

	// Test sequence diagram
	t.Run("SequenceDiagram", func(t *testing.T) {
		diagram := generator.GenerateComprehensiveSequenceDiagram(service)
		if len(diagram) == 0 {
			t.Fatal("Sequence diagram is empty")
		}
		if !containsString(diagram, "sequenceDiagram") {
			t.Error("Sequence diagram missing declaration")
		}
		if !containsString(diagram, "ProcessPayment") {
			t.Error("Sequence diagram missing method name")
		}
		t.Logf("Sequence Diagram:\n%s", diagram)
	})

	// Test class diagram
	t.Run("ClassDiagram", func(t *testing.T) {
		diagram := generator.GenerateEnhancedClassDiagram(messages)
		if len(diagram) == 0 {
			t.Fatal("Class diagram is empty")
		}
		if !containsString(diagram, "classDiagram") {
			t.Error("Class diagram missing declaration")
		}
		t.Logf("Class Diagram:\n%s", diagram)
	})

	// Test ER diagram
	t.Run("ERDiagram", func(t *testing.T) {
		diagram := generator.GenerateERDiagram(messages)
		if len(diagram) == 0 {
			t.Fatal("ER diagram is empty")
		}
		if !containsString(diagram, "erDiagram") {
			t.Error("ER diagram missing declaration")
		}
		t.Logf("ER Diagram:\n%s", diagram)
	})

	// Test state machine diagram
	t.Run("StateMachineDiagram", func(t *testing.T) {
		diagram := generator.GenerateStateMachineDiagram(service)
		if len(diagram) == 0 {
			t.Fatal("State machine diagram is empty")
		}
		if !containsString(diagram, "stateDiagram") {
			t.Error("State machine diagram missing declaration")
		}
		t.Logf("State Machine Diagram:\n%s", diagram)
	})

	// Test flowchart
	t.Run("Flowchart", func(t *testing.T) {
		diagram := generator.GenerateMethodFlowchart(service)
		if len(diagram) == 0 {
			t.Fatal("Flowchart is empty")
		}
		if !containsString(diagram, "flowchart") {
			t.Error("Flowchart missing declaration")
		}
		t.Logf("Flowchart:\n%s", diagram)
	})

	// Test mindmap
	t.Run("Mindmap", func(t *testing.T) {
		diagram := generator.GenerateServiceMindmap(service, messages)
		if len(diagram) == 0 {
			t.Fatal("Mindmap is empty")
		}
		if !containsString(diagram, "mindmap") {
			t.Error("Mindmap missing declaration")
		}
		t.Logf("Mindmap:\n%s", diagram)
	})

	// Test C4 context diagram
	t.Run("C4ContextDiagram", func(t *testing.T) {
		diagram := generator.GenerateC4ContextDiagram(service)
		if len(diagram) == 0 {
			t.Fatal("C4 context diagram is empty")
		}
		if !containsString(diagram, "C4Context") {
			t.Error("C4 context diagram missing declaration")
		}
		t.Logf("C4 Context Diagram:\n%s", diagram)
	})

	// Test C4 container diagram
	t.Run("C4ContainerDiagram", func(t *testing.T) {
		diagram := generator.GenerateC4ContainerDiagram(service)
		if len(diagram) == 0 {
			t.Fatal("C4 container diagram is empty")
		}
		if !containsString(diagram, "C4Container") {
			t.Error("C4 container diagram missing declaration")
		}
		t.Logf("C4 Container Diagram:\n%s", diagram)
	})

	// Test data flow diagram
	t.Run("DataFlowDiagram", func(t *testing.T) {
		diagram := generator.GenerateDataFlowDiagram(service, messages)
		if len(diagram) == 0 {
			t.Fatal("Data flow diagram is empty")
		}
		if !containsString(diagram, "graph LR") {
			t.Error("Data flow diagram missing graph declaration")
		}
		t.Logf("Data Flow Diagram:\n%s", diagram)
	})
}

// TestCompleteDiagramSet tests generation of all diagram types
func TestCompleteDiagramSet(t *testing.T) {
	service := &DocService{
		Name:        "OrderService",
		FullName:    "orders.v1.OrderService",
		Description: "Order management service",
		Methods: []DocMethod{
			{
				Name:            "CreateOrder",
				FullName:        "orders.v1.OrderService.CreateOrder",
				Description:     "Create new order",
				InputType:       "orders.v1.CreateOrderRequest",
				OutputType:      "orders.v1.Order",
				ClientStreaming: false,
				ServerStreaming: false,
			},
		},
	}

	messages := []DocMessage{
		{
			Name:     "Order",
			FullName: "orders.v1.Order",
			Fields: []DocField{
				{Name: "id", Type: "string"},
				{Name: "status", Type: "string"},
			},
		},
	}

	enums := []DocEnum{
		{
			Name:     "OrderStatus",
			FullName: "orders.v1.OrderStatus",
			Values: []DocEnumValue{
				{Name: "PENDING", Number: 0},
				{Name: "PROCESSING", Number: 1},
				{Name: "COMPLETED", Number: 2},
			},
		},
	}

	config := DefaultDiagramConfig()
	generator := NewEnhancedMermaidGenerator(config)

	diagrams := generator.GenerateCompleteDiagramSet(service, messages, enums)

	// Verify all diagram types were generated
	expectedDiagrams := []string{
		"architecture", "sequence", "class", "erd",
		"state_machine", "flowchart", "mindmap",
		"c4_context", "c4_container", "data_flow",
	}

	for _, diagramType := range expectedDiagrams {
		diagram, exists := diagrams[diagramType]
		if !exists {
			t.Errorf("Missing diagram type: %s", diagramType)
			continue
		}

		if len(diagram) == 0 {
			t.Errorf("Diagram %s is empty", diagramType)
		}

		t.Logf("Generated %s diagram (%d bytes)", diagramType, len(diagram))
	}

	t.Logf("Generated %d diagram types", len(diagrams))
}

// Helper function
func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) >= len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
