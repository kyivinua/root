package pipeline

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/kyivinua/docgen-tool/tools/protoctx"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// ApiDocModel represents the complete documentation model for the API.
type ApiDocModel struct {
	Modules      []DocModule       `json:"modules"`
	GeneratedAt  time.Time         `json:"generated_at"`
	SourceCommit string            `json:"source_commit"`
	Tools        map[string]string `json:"tools,omitempty"`
	Statistics   map[string]int64  `json:"statistics,omitempty"`
}

// DocModule represents a documentation module (usually corresponds to a package).
type DocModule struct {
	Name        string       `json:"name"`
	Version     string       `json:"version,omitempty"`
	Package     string       `json:"package"`
	Summary     string       `json:"summary,omitempty"`
	Description string       `json:"description,omitempty"`
	Services    []DocService `json:"services"`
	Messages    []DocMessage `json:"messages"`
	Enums       []DocEnum    `json:"enums,omitempty"`
	Visibility  string       `json:"visibility,omitempty"`
}

// DocService represents a gRPC service.
type DocService struct {
	FQN         string      `json:"fqn"`
	Name        string      `json:"name"`
	Package     string      `json:"package"`
	File        string      `json:"file"`
	Visibility  string      `json:"visibility,omitempty"`
	Summary     string      `json:"summary,omitempty"`
	Description string      `json:"description,omitempty"`
	Methods     []DocMethod `json:"methods"`
}

// DocMethod represents a service method.
type DocMethod struct {
	FQN           string   `json:"fqn"`
	Name          string   `json:"name"`
	InputType     string   `json:"input_type"`
	OutputType    string   `json:"output_type"`
	HTTPMethods   []string `json:"http_methods,omitempty"`
	HTTPPaths     []string `json:"http_paths,omitempty"`
	StreamingMode string   `json:"streaming_mode,omitempty"`
	Visibility    string   `json:"visibility,omitempty"`
	Deprecated    bool     `json:"deprecated"`
	Summary       string   `json:"summary,omitempty"`
	Description   string   `json:"description,omitempty"`
}

// DocMessage represents a Protobuf message.
type DocMessage struct {
	FQN         string     `json:"fqn"`
	Name        string     `json:"name"`
	Package     string     `json:"package"`
	File        string     `json:"file"`
	Summary     string     `json:"summary,omitempty"`
	Description string     `json:"description,omitempty"`
	Fields      []DocField `json:"fields"`
	Deprecated  bool       `json:"deprecated"`
}

// DocField represents a message field.
type DocField struct {
	Name        string `json:"name"`
	Number      int32  `json:"number"`
	Label       string `json:"label"` // "optional", "required", "repeated"
	Type        string `json:"type"`
	JSONName    string `json:"json_name,omitempty"`
	Oneof       string `json:"oneof,omitempty"`
	Required    bool   `json:"required"`
	Deprecated  bool   `json:"deprecated"`
	Summary     string `json:"summary,omitempty"`
	Description string `json:"description,omitempty"`
}

// DocEnum represents a Protobuf enum.
type DocEnum struct {
	FQN         string         `json:"fqn"`
	Name        string         `json:"name"`
	Package     string         `json:"package"`
	File        string         `json:"file"`
	Summary     string         `json:"summary,omitempty"`
	Description string         `json:"description,omitempty"`
	Values      []DocEnumValue `json:"values"`
	Deprecated  bool           `json:"deprecated"`
}

// DocEnumValue represents an enum value.
type DocEnumValue struct {
	Name        string `json:"name"`
	Number      int32  `json:"number"`
	Summary     string `json:"summary,omitempty"`
	Description string `json:"description,omitempty"`
	Deprecated  bool   `json:"deprecated"`
}

// BuildApiDocModel builds the documentation model from a ProtoContext.
func BuildApiDocModel(ctx *protoctx.Context, sourceCommit string) (*ApiDocModel, error) {
	model := &ApiDocModel{
		GeneratedAt:  time.Now().UTC(),
		SourceCommit: sourceCommit,
		Tools:        make(map[string]string),
		Statistics:   make(map[string]int64),
		Modules:      []DocModule{},
	}

	// Group by package
	packageMap := make(map[string]*DocModule)

	// Process services
	services := ctx.ListServices()
	for _, sd := range services {
		pkg := string(sd.ParentFile().Package())
		module := getOrCreateModule(packageMap, pkg)

		docService := buildDocService(sd, ctx)
		module.Services = append(module.Services, docService)
	}

	// Process messages
	messages := ctx.ListMessages()
	for _, md := range messages {
		pkg := string(md.ParentFile().Package())
		module := getOrCreateModule(packageMap, pkg)

		docMessage := buildDocMessage(md, ctx)
		module.Messages = append(module.Messages, docMessage)
	}

	// Process enums
	enums := ctx.ListEnums()
	for _, ed := range enums {
		pkg := string(ed.ParentFile().Package())
		module := getOrCreateModule(packageMap, pkg)

		docEnum := buildDocEnum(ed, ctx)
		module.Enums = append(module.Enums, docEnum)
	}

	// Convert map to slice
	for _, module := range packageMap {
		model.Modules = append(model.Modules, *module)
	}

	// Compute statistics
	model.Statistics["total_modules"] = int64(len(model.Modules))
	model.Statistics["total_services"] = int64(len(services))
	model.Statistics["total_messages"] = int64(len(messages))
	model.Statistics["total_enums"] = int64(len(enums))

	return model, nil
}

func getOrCreateModule(packageMap map[string]*DocModule, pkg string) *DocModule {
	if module, exists := packageMap[pkg]; exists {
		return module
	}

	module := &DocModule{
		Name:     pkg,
		Package:  pkg,
		Services: []DocService{},
		Messages: []DocMessage{},
		Enums:    []DocEnum{},
	}
	packageMap[pkg] = module
	return module
}

func buildDocService(sd protoreflect.ServiceDescriptor, ctx *protoctx.Context) DocService {
	comments, _ := ctx.CommentsIndex().GetByDescriptor(sd)

	svc := DocService{
		FQN:         string(sd.FullName()),
		Name:        string(sd.Name()),
		Package:     string(sd.ParentFile().Package()),
		File:        sd.ParentFile().Path(),
		Summary:     comments.Summary(),
		Description: comments.Description(),
		Methods:     []DocMethod{},
	}

	for i := 0; i < sd.Methods().Len(); i++ {
		md := sd.Methods().Get(i)
		method := buildDocMethod(md, ctx)
		svc.Methods = append(svc.Methods, method)
	}

	return svc
}

func buildDocMethod(md protoreflect.MethodDescriptor, ctx *protoctx.Context) DocMethod {
	comments, _ := ctx.CommentsIndex().GetByDescriptor(md)

	streamingMode := "unary"
	if md.IsStreamingClient() && md.IsStreamingServer() {
		streamingMode = "bidi_streaming"
	} else if md.IsStreamingClient() {
		streamingMode = "client_streaming"
	} else if md.IsStreamingServer() {
		streamingMode = "server_streaming"
	}

	method := DocMethod{
		FQN:           string(md.FullName()),
		Name:          string(md.Name()),
		InputType:     string(md.Input().FullName()),
		OutputType:    string(md.Output().FullName()),
		StreamingMode: streamingMode,
		Summary:       comments.Summary(),
		Description:   comments.Description(),
		HTTPMethods:   []string{},
		HTTPPaths:     []string{},
	}

	// TODO: Extract HTTP bindings from google.api.http options
	// This requires parsing method options

	return method
}

func buildDocMessage(md protoreflect.MessageDescriptor, ctx *protoctx.Context) DocMessage {
	comments, _ := ctx.CommentsIndex().GetByDescriptor(md)

	msg := DocMessage{
		FQN:         string(md.FullName()),
		Name:        string(md.Name()),
		Package:     string(md.ParentFile().Package()),
		File:        md.ParentFile().Path(),
		Summary:     comments.Summary(),
		Description: comments.Description(),
		Fields:      []DocField{},
	}

	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		field := buildDocField(fd, ctx)
		msg.Fields = append(msg.Fields, field)
	}

	return msg
}

func buildDocField(fd protoreflect.FieldDescriptor, ctx *protoctx.Context) DocField {
	comments, _ := ctx.CommentsIndex().GetByDescriptor(fd)

	label := "optional"
	if fd.Cardinality() == protoreflect.Repeated {
		label = "repeated"
	} else if fd.Cardinality() == protoreflect.Required {
		label = "required"
	}

	typeName := fd.Kind().String()
	if fd.Kind() == protoreflect.MessageKind || fd.Kind() == protoreflect.EnumKind {
		typeName = string(fd.Message().FullName())
		if fd.Kind() == protoreflect.EnumKind {
			typeName = string(fd.Enum().FullName())
		}
	}

	field := DocField{
		Name:        string(fd.Name()),
		Number:      int32(fd.Number()),
		Label:       label,
		Type:        typeName,
		JSONName:    fd.JSONName(),
		Summary:     comments.Summary(),
		Description: comments.Description(),
	}

	if fd.ContainingOneof() != nil {
		field.Oneof = string(fd.ContainingOneof().Name())
	}

	return field
}

func buildDocEnum(ed protoreflect.EnumDescriptor, ctx *protoctx.Context) DocEnum {
	comments, _ := ctx.CommentsIndex().GetByDescriptor(ed)

	enum := DocEnum{
		FQN:         string(ed.FullName()),
		Name:        string(ed.Name()),
		Package:     string(ed.ParentFile().Package()),
		File:        ed.ParentFile().Path(),
		Summary:     comments.Summary(),
		Description: comments.Description(),
		Values:      []DocEnumValue{},
	}

	for i := 0; i < ed.Values().Len(); i++ {
		vd := ed.Values().Get(i)
		valComments, _ := ctx.CommentsIndex().GetByDescriptor(vd)

		value := DocEnumValue{
			Name:        string(vd.Name()),
			Number:      int32(vd.Number()),
			Summary:     valComments.Summary(),
			Description: valComments.Description(),
		}
		enum.Values = append(enum.Values, value)
	}

	return enum
}

// SaveToFile saves the ApiDocModel to a JSON file.
func (m *ApiDocModel) SaveToFile(path string) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal model: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

// LoadFromFile loads an ApiDocModel from a JSON file.
func LoadApiDocModelFromFile(path string) (*ApiDocModel, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var model ApiDocModel
	if err := json.Unmarshal(data, &model); err != nil {
		return nil, fmt.Errorf("unmarshal model: %w", err)
	}

	return &model, nil
}
