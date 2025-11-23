package diagrams

import (
	"encoding/xml"
	"fmt"
	"strings"
)

// GraphML represents the root GraphML document
type GraphML struct {
	XMLName xml.Name `xml:"graphml"`
	XMLNS   string   `xml:"xmlns,attr"`
	XMLNSXSI string  `xml:"xmlns:xsi,attr"`
	XSISchemaLocation string `xml:"xsi:schemaLocation,attr"`
	Keys    []Key    `xml:"key"`
	Graph   *Graph   `xml:"graph"`
}

// Key represents a GraphML key definition
type Key struct {
	ID       string `xml:"id,attr"`
	For      string `xml:"for,attr"` // node, edge, graph
	AttrName string `xml:"attr.name,attr"`
	AttrType string `xml:"attr.type,attr"`
}

// Graph represents a graph in GraphML
type Graph struct {
	ID          string `xml:"id,attr"`
	EdgeDefault string `xml:"edgedefault,attr"` // directed or undirected
	Nodes       []Node `xml:"node"`
	Edges       []Edge `xml:"edge"`
}

// Node represents a node in the graph
type Node struct {
	ID   string `xml:"id,attr"`
	Data []Data `xml:"data"`
}

// Edge represents an edge in the graph
type Edge struct {
	ID     string `xml:"id,attr,omitempty"`
	Source string `xml:"source,attr"`
	Target string `xml:"target,attr"`
	Data   []Data `xml:"data"`
}

// Data represents node/edge data
type Data struct {
	Key   string `xml:"key,attr"`
	Value string `xml:",chardata"`
}

// ServiceGraphMLGenerator generates GraphML from service documentation
type ServiceGraphMLGenerator struct {
	serviceCount int
	methodCount  int
	messageCount int
	edgeCount    int
}

// NewServiceGraphMLGenerator creates a new GraphML generator
func NewServiceGraphMLGenerator() *ServiceGraphMLGenerator {
	return &ServiceGraphMLGenerator{}
}

// GenerateServiceGraphML generates a GraphML representation of a gRPC service
func (g *ServiceGraphMLGenerator) GenerateServiceGraphML(service *DocService, messages []DocMessage) (*GraphML, error) {
	g.serviceCount = 0
	g.methodCount = 0
	g.messageCount = 0
	g.edgeCount = 0

	graphml := &GraphML{
		XMLNS:             "http://graphml.graphdrawing.org/xmlns",
		XMLNSXSI:          "http://www.w3.org/2001/XMLSchema-instance",
		XSISchemaLocation: "http://graphml.graphdrawing.org/xmlns http://graphml.graphdrawing.org/xmlns/1.0/graphml.xsd",
		Keys:              g.defineKeys(),
		Graph: &Graph{
			ID:          service.Name,
			EdgeDefault: "directed",
			Nodes:       []Node{},
			Edges:       []Edge{},
		},
	}

	// Add service node
	serviceNode := g.createServiceNode(service)
	graphml.Graph.Nodes = append(graphml.Graph.Nodes, serviceNode)

	// Track messages to avoid duplicates
	addedMessages := make(map[string]bool)

	// Build message lookup map
	messageMap := make(map[string]*DocMessage)
	for i := range messages {
		messageMap[messages[i].FullName] = &messages[i]
		messageMap[messages[i].Name] = &messages[i]
	}

	// Add method nodes and connections
	for _, method := range service.Methods {
		methodNode := g.createMethodNode(&method)
		graphml.Graph.Nodes = append(graphml.Graph.Nodes, methodNode)

		// Connect service to method
		edge := g.createEdge(serviceNode.ID, methodNode.ID, "contains", "")
		graphml.Graph.Edges = append(graphml.Graph.Edges, edge)

		// Add input message if not already added
		if !addedMessages[method.InputType] {
			inputMsg := messageMap[method.InputType]
			inputNode := g.createMessageNode(method.InputType, inputMsg)
			graphml.Graph.Nodes = append(graphml.Graph.Nodes, inputNode)
			addedMessages[method.InputType] = true
		}

		// Add output message if not already added
		if !addedMessages[method.OutputType] {
			outputMsg := messageMap[method.OutputType]
			outputNode := g.createMessageNode(method.OutputType, outputMsg)
			graphml.Graph.Nodes = append(graphml.Graph.Nodes, outputNode)
			addedMessages[method.OutputType] = true
		}

		// Connect method to input/output messages
		inputEdge := g.createEdge(getNodeID(method.InputType, "message"), methodNode.ID, "input", "")
		graphml.Graph.Edges = append(graphml.Graph.Edges, inputEdge)

		outputEdge := g.createEdge(methodNode.ID, getNodeID(method.OutputType, "message"), "output", "")
		graphml.Graph.Edges = append(graphml.Graph.Edges, outputEdge)
	}

	// Add message relationships (field dependencies)
	for _, message := range messages {
		if !addedMessages[message.FullName] && !addedMessages[message.Name] {
			continue
		}
		for _, field := range message.Fields {
			// If field type is another message, create a relationship
			fieldType := field.TypeName
			if fieldType == "" {
				fieldType = field.Type
			}
			if isMessageType(fieldType) && (addedMessages[fieldType] || addedMessages[getShortName(fieldType)]) {
				fromID := getNodeID(message.Name, "message")
				toID := getNodeID(getShortName(fieldType), "message")
				edge := g.createEdge(fromID, toID, "references", field.Name)
				graphml.Graph.Edges = append(graphml.Graph.Edges, edge)
			}
		}
	}

	return graphml, nil
}

// defineKeys defines the GraphML keys
func (g *ServiceGraphMLGenerator) defineKeys() []Key {
	return []Key{
		{ID: "d0", For: "node", AttrName: "name", AttrType: "string"},
		{ID: "d1", For: "node", AttrName: "type", AttrType: "string"},
		{ID: "d2", For: "node", AttrName: "description", AttrType: "string"},
		{ID: "d3", For: "node", AttrName: "package", AttrType: "string"},
		{ID: "d4", For: "node", AttrName: "streaming", AttrType: "string"},
		{ID: "d5", For: "node", AttrName: "fields", AttrType: "string"},
		{ID: "d6", For: "edge", AttrName: "relationship", AttrType: "string"},
		{ID: "d7", For: "edge", AttrName: "label", AttrType: "string"},
		{ID: "d8", For: "graph", AttrName: "description", AttrType: "string"},
	}
}

// createServiceNode creates a node for the service
func (g *ServiceGraphMLGenerator) createServiceNode(service *DocService) Node {
	g.serviceCount++

	// Extract package from FullName
	pkg := ""
	if parts := strings.Split(service.FullName, "."); len(parts) > 1 {
		pkg = strings.Join(parts[:len(parts)-1], ".")
	}

	return Node{
		ID: fmt.Sprintf("service_%d", g.serviceCount),
		Data: []Data{
			{Key: "d0", Value: service.Name},
			{Key: "d1", Value: "service"},
			{Key: "d2", Value: service.Description},
			{Key: "d3", Value: pkg},
		},
	}
}

// createMethodNode creates a node for a method
func (g *ServiceGraphMLGenerator) createMethodNode(method *DocMethod) Node {
	g.methodCount++

	streaming := "unary"
	if method.ClientStreaming && method.ServerStreaming {
		streaming = "bidirectional"
	} else if method.ClientStreaming {
		streaming = "client_stream"
	} else if method.ServerStreaming {
		streaming = "server_stream"
	}

	return Node{
		ID: fmt.Sprintf("method_%d", g.methodCount),
		Data: []Data{
			{Key: "d0", Value: method.Name},
			{Key: "d1", Value: "method"},
			{Key: "d2", Value: method.Description},
			{Key: "d4", Value: streaming},
		},
	}
}

// createMessageNode creates a node for a message
func (g *ServiceGraphMLGenerator) createMessageNode(fullName string, message *DocMessage) Node {
	g.messageCount++

	description := ""
	fields := ""
	if message != nil {
		description = message.Description
		fieldNames := make([]string, len(message.Fields))
		for i, field := range message.Fields {
			fieldNames[i] = fmt.Sprintf("%s: %s", field.Name, field.Type)
		}
		fields = strings.Join(fieldNames, ", ")
	}

	return Node{
		ID: getNodeID(fullName, "message"),
		Data: []Data{
			{Key: "d0", Value: getShortName(fullName)},
			{Key: "d1", Value: "message"},
			{Key: "d2", Value: description},
			{Key: "d5", Value: fields},
		},
	}
}

// createEdge creates an edge between two nodes
func (g *ServiceGraphMLGenerator) createEdge(source, target, relationship, label string) Edge {
	g.edgeCount++
	return Edge{
		ID:     fmt.Sprintf("e%d", g.edgeCount),
		Source: source,
		Target: target,
		Data: []Data{
			{Key: "d6", Value: relationship},
			{Key: "d7", Value: label},
		},
	}
}

// Helper functions
func getNodeID(fullName, nodeType string) string {
	// Create a unique but readable ID
	shortName := getShortName(fullName)
	sanitized := sanitizeName(shortName)
	return fmt.Sprintf("%s_%s", nodeType, sanitized)
}

func isMessageType(typeName string) bool {
	// Check if type is a message (not a primitive type)
	primitives := map[string]bool{
		"string": true, "int32": true, "int64": true, "uint32": true, "uint64": true,
		"sint32": true, "sint64": true, "fixed32": true, "fixed64": true,
		"sfixed32": true, "sfixed64": true, "bool": true, "bytes": true,
		"float": true, "double": true,
	}
	shortType := getShortName(typeName)
	return !primitives[shortType]
}

// ToXML converts the GraphML to XML string
func (g *GraphML) ToXML() (string, error) {
	data, err := xml.MarshalIndent(g, "", "  ")
	if err != nil {
		return "", err
	}
	return xml.Header + string(data), nil
}
