package hldgen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPromptLoader(t *testing.T) {
	tests := []struct {
		name         string
		basePath     string
		version      string
		wantBasePath string
		wantVersion  string
	}{
		{
			name:         "with explicit values",
			basePath:     "custom/prompts",
			version:      "v2",
			wantBasePath: "custom/prompts",
			wantVersion:  "v2",
		},
		{
			name:         "with defaults",
			basePath:     "",
			version:      "",
			wantBasePath: "prompts",
			wantVersion:  "v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewPromptLoader(tt.basePath, tt.version)
			assert.NotNil(t, loader)
			assert.Equal(t, tt.wantBasePath, loader.basePath)
			assert.Equal(t, tt.wantVersion, loader.version)
			assert.NotNil(t, loader.cache)
		})
	}
}

func TestPromptLoader_Load(t *testing.T) {
	// Create test directory
	tmpDir := t.TempDir()
	v1Dir := filepath.Join(tmpDir, "v1")
	require.NoError(t, os.MkdirAll(v1Dir, 0755))

	// Create test prompt file
	testPrompt := `<?xml version="1.0" encoding="UTF-8"?>
<prompt version="1.0" role="architect">
  <system>You are an architect.</system>
  <context>Module: {{.ModuleName}}</context>
  <thinking>Think carefully</thinking>
  <instructions>Generate design</instructions>
</prompt>`
	require.NoError(t, os.WriteFile(filepath.Join(v1Dir, "architect.xml"), []byte(testPrompt), 0644))

	loader := NewPromptLoader(tmpDir, "v1")

	t.Run("load valid prompt", func(t *testing.T) {
		pt, err := loader.Load(RoleArchitect)
		require.NoError(t, err)
		assert.Equal(t, "1.0", pt.Version)
		assert.Equal(t, "architect", pt.Role)
		assert.Contains(t, pt.System, "architect")
		assert.Contains(t, pt.Context, "ModuleName")
	})

	t.Run("load from cache", func(t *testing.T) {
		// Load again - should use cache
		pt, err := loader.Load(RoleArchitect)
		require.NoError(t, err)
		assert.NotNil(t, pt)

		// Verify it's from cache
		cached, ok := loader.cache[string(RoleArchitect)]
		assert.True(t, ok)
		assert.Equal(t, pt, cached)
	})

	t.Run("load non-existent role", func(t *testing.T) {
		_, err := loader.Load(RolePM)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read prompt template")
	})
}

func TestPromptLoader_Render(t *testing.T) {
	// Create test directory
	tmpDir := t.TempDir()
	v1Dir := filepath.Join(tmpDir, "v1")
	require.NoError(t, os.MkdirAll(v1Dir, 0755))

	// Create test prompt with templates
	testPrompt := `<?xml version="1.0" encoding="UTF-8"?>
<prompt version="1.0" role="architect">
  <system>You are an architect for {{.ModuleName}}.</system>
  <context>
**Module**: {{.ModuleName}}
**Services**: {{.ServicesCount}}
{{range .Services}}
- {{.Name}} ({{.MethodCount}} methods)
{{end}}
  </context>
  <thinking>Analyze the system</thinking>
  <instructions>Generate comprehensive design</instructions>
</prompt>`
	require.NoError(t, os.WriteFile(filepath.Join(v1Dir, "architect.xml"), []byte(testPrompt), 0644))

	loader := NewPromptLoader(tmpDir, "v1")

	t.Run("render with data", func(t *testing.T) {
		data := PromptData{
			ModuleName:    "user-service",
			ServicesCount: 2,
			Services: []ServiceTemplateData{
				{Name: "UserService", MethodCount: 5},
				{Name: "AuthService", MethodCount: 3},
			},
		}

		rendered, err := loader.Render(RoleArchitect, data)
		require.NoError(t, err)

		// Check that template variables were replaced
		assert.Contains(t, rendered, "<system>")
		assert.Contains(t, rendered, "user-service")
		assert.Contains(t, rendered, "UserService")
		assert.Contains(t, rendered, "AuthService")
		assert.Contains(t, rendered, "5 methods")
		assert.Contains(t, rendered, "3 methods")
		assert.Contains(t, rendered, "<thinking>")
		assert.Contains(t, rendered, "<instructions>")
	})

	t.Run("render with empty data", func(t *testing.T) {
		data := PromptData{
			ModuleName:    "empty-service",
			ServicesCount: 0,
		}

		rendered, err := loader.Render(RoleArchitect, data)
		require.NoError(t, err)
		assert.Contains(t, rendered, "empty-service")
	})
}

func TestConvertAgentInputToPromptData(t *testing.T) {
	input := &AgentInput{
		Docs: &ConsolidatedDocs{
			ModuleName:   "payment-service",
			SourceCommit: "abc123",
			Services: []ServiceDoc{
				{
					Name:        "PaymentService",
					Description: "Handles payments",
					Methods: []MethodDoc{
						{
							Name:        "ProcessPayment",
							Description: "Process a payment",
							InputType:   "PaymentRequest",
							OutputType:  "PaymentResponse",
						},
					},
				},
			},
			Messages: []MessageDoc{
				{Name: "PaymentRequest"},
				{Name: "PaymentResponse"},
			},
		},
		EnrichedCtx: &EnrichedContext{
			RAGContext: []RAGDocument{
				{Source: "payment-docs.md", Score: 0.95},
			},
		},
		PreviousDraft: &HLDOutput{
			Architecture: "Previous design",
		},
		Criticism: []Criticism{
			{
				Agent:      "security",
				Score:      0.80,
				Issues:     []Issue{{Message: "Missing auth"}},
				Suggestion: "Add OAuth2",
			},
		},
		Round: 2,
	}

	data := ConvertAgentInputToPromptData(input)

	// Verify basic fields
	assert.Equal(t, "payment-service", data.ModuleName)
	assert.Equal(t, "abc123", data.SourceCommit)
	assert.Equal(t, 1, data.ServicesCount)
	assert.Equal(t, 2, data.MessagesCount)
	assert.Equal(t, 1, data.MethodsCount)

	// Verify services
	require.Len(t, data.Services, 1)
	assert.Equal(t, "PaymentService", data.Services[0].Name)
	assert.Equal(t, 1, data.Services[0].MethodCount)
	require.Len(t, data.Services[0].Methods, 1)
	assert.Equal(t, "ProcessPayment", data.Services[0].Methods[0].Name)

	// Verify RAG context
	require.Len(t, data.RAGContext, 1)
	assert.Equal(t, "payment-docs.md", data.RAGContext[0].Source)
	assert.Equal(t, 0.95, data.RAGContext[0].Score)

	// Verify previous draft
	assert.Equal(t, "Previous design", data.PreviousDraft)

	// Verify criticism
	require.Len(t, data.Criticism, 1)
	assert.Equal(t, 0.80, data.Criticism[0].Score)
	assert.Equal(t, 1, data.Criticism[0].IssuesCount)
	assert.Equal(t, "Add OAuth2", data.Criticism[0].Suggestion)
}

func TestPromptLoader_Integration(t *testing.T) {
	// This test verifies the full workflow with real prompt files
	loader := NewPromptLoader("../../../prompts", "v1")

	// Test loading each role's prompt
	roles := []AgentRole{
		RoleArchitect,
		RolePM,
		RoleSecurity,
		RoleSRE,
		RoleQA,
	}

	for _, role := range roles {
		t.Run(string(role), func(t *testing.T) {
			// Load the prompt template
			pt, err := loader.Load(role)
			if err != nil {
				t.Skipf("Prompt file not found for role %s: %v", role, err)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, pt.System, "System section should not be empty")
			assert.NotEmpty(t, pt.Context, "Context section should not be empty")
			assert.NotEmpty(t, pt.Instructions, "Instructions section should not be empty")

			// Test rendering with sample data
			data := PromptData{
				ModuleName:    "test-service",
				ServicesCount: 1,
				MessagesCount: 2,
				MethodsCount:  3,
				Services: []ServiceTemplateData{
					{
						Name:        "TestService",
						Description: "A test service",
						MethodCount: 3,
						Methods: []MethodTemplateData{
							{Name: "GetUser", Description: "Get user by ID"},
							{Name: "CreateUser", Description: "Create new user"},
							{Name: "UpdateUser", Description: "Update user"},
						},
					},
				},
			}

			rendered, err := loader.Render(role, data)
			require.NoError(t, err)
			assert.Contains(t, rendered, "test-service")
			assert.Contains(t, rendered, "<system>")
			assert.Contains(t, rendered, "<context>")
			assert.Contains(t, rendered, "<instructions>")
		})
	}
}
