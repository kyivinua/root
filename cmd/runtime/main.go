package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/kyivinua/docgen-tool/tools/protoctx"
)

var (
	rs *RuntimeState
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Get descriptor path from environment or use default
	imagePath := os.Getenv("PROTO_DESCRIPTOR_PATH")
	if imagePath == "" {
		imagePath = "api-docs/descriptors/image.bin"
	}

	log.Printf("Starting runtime service...")
	log.Printf("Descriptor path: %s", imagePath)

	// Initialize runtime state
	rs = NewRuntimeState()
	go rs.Init(imagePath)

	// Setup HTTP routes
	http.HandleFunc("/debug/readyz", handleReadyz)
	http.HandleFunc("/debug/schema/service", handleServiceSchema)
	http.HandleFunc("/proto/decode", handleDecode)
	http.HandleFunc("/proto/encode", handleEncode)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// handleReadyz returns the readiness status of the service.
//
// GET /debug/readyz
//
// Returns:
//   - 200 OK if context is ready
//   - 503 Service Unavailable if context is not ready
func handleReadyz(w http.ResponseWriter, r *http.Request) {
	if !rs.IsReady() {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, "NOT_READY (status: %s)\n", rs.Status())
		if err := rs.Error(); err != nil {
			fmt.Fprintf(w, "Error: %v\n", err)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "READY\n")
}

// handleServiceSchema returns the schema of a service.
//
// GET /debug/schema/service?name=<FQN>
//
// Example: /debug/schema/service?name=company.user.v1.UserService
func handleServiceSchema(w http.ResponseWriter, r *http.Request) {
	serviceName := r.URL.Query().Get("name")
	if serviceName == "" {
		http.Error(w, "missing 'name' parameter", http.StatusBadRequest)
		return
	}

	var result map[string]interface{}

	err := rs.WithContext(func(ctx *protoctx.Context) error {
		sd, err := ctx.FindServiceDescriptor(serviceName)
		if err != nil {
			return fmt.Errorf("find service: %w", err)
		}

		comments, _ := ctx.CommentsIndex().GetByDescriptor(sd)

		methods := []map[string]interface{}{}
		for i := 0; i < sd.Methods().Len(); i++ {
			md := sd.Methods().Get(i)
			methodComments, _ := ctx.CommentsIndex().GetByDescriptor(md)

			streamingMode := "unary"
			if md.IsStreamingClient() && md.IsStreamingServer() {
				streamingMode = "bidi_streaming"
			} else if md.IsStreamingClient() {
				streamingMode = "client_streaming"
			} else if md.IsStreamingServer() {
				streamingMode = "server_streaming"
			}

			methods = append(methods, map[string]interface{}{
				"name":           string(md.Name()),
				"input_type":     string(md.Input().FullName()),
				"output_type":    string(md.Output().FullName()),
				"streaming_mode": streamingMode,
				"summary":        methodComments.Summary(),
				"description":    methodComments.Description(),
			})
		}

		result = map[string]interface{}{
			"fqn":         string(sd.FullName()),
			"name":        string(sd.Name()),
			"package":     string(sd.ParentFile().Package()),
			"file":        sd.ParentFile().Path(),
			"summary":     comments.Summary(),
			"description": comments.Description(),
			"methods":     methods,
		}

		return nil
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleDecode decodes wire-format bytes to JSON.
//
// POST /proto/decode?type=<FQN>
// Body: wire-format bytes
//
// Example: /proto/decode?type=company.user.v1.UserProfile
func handleDecode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	typeName := r.URL.Query().Get("type")
	if typeName == "" {
		http.Error(w, "missing 'type' parameter", http.StatusBadRequest)
		return
	}

	wire, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var jsonData []byte
	err = rs.WithContext(func(ctx *protoctx.Context) error {
		var err error
		jsonData, err = ctx.MarshalWireToJSON(typeName, wire)
		return err
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

// handleEncode encodes JSON to wire-format bytes.
//
// POST /proto/encode?type=<FQN>
// Body: JSON bytes
//
// Example: /proto/encode?type=company.user.v1.UserProfile
func handleEncode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	typeName := r.URL.Query().Get("type")
	if typeName == "" {
		http.Error(w, "missing 'type' parameter", http.StatusBadRequest)
		return
	}

	jsonData, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var wire []byte
	err = rs.WithContext(func(ctx *protoctx.Context) error {
		var err error
		wire, err = ctx.UnmarshalJSONToWire(typeName, jsonData)
		return err
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(wire)
}
