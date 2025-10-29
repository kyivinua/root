// Package generator provides template functions for documentation generation.
package generator

import (
	"strings"
	"text/template"
)

// GetTemplateFunctions returns custom template functions.
func GetTemplateFunctions() template.FuncMap {
	return template.FuncMap{
		"title":      strings.Title,
		"lower":      strings.ToLower,
		"upper":      strings.ToUpper,
		"replace":    strings.ReplaceAll,
		"contains":   strings.Contains,
		"hasPrefix":  strings.HasPrefix,
		"hasSuffix":  strings.HasSuffix,
		"trim":       strings.TrimSpace,
		"goType":     goType,
		"pythonType": pythonType,
		"jsonType":   jsonType,
		"join":       strings.Join,
	}
}

// goType converts proto type to Go type.
func goType(protoType string) string {
	switch protoType {
	case "string":
		return "string"
	case "int32":
		return "int32"
	case "int64":
		return "int64"
	case "uint32":
		return "uint32"
	case "uint64":
		return "uint64"
	case "bool":
		return "bool"
	case "float":
		return "float32"
	case "double":
		return "float64"
	case "bytes":
		return "[]byte"
	default:
		return "*" + protoType
	}
}

// pythonType converts proto type to Python type.
func pythonType(protoType string) string {
	switch protoType {
	case "string":
		return "str"
	case "int32", "int64", "uint32", "uint64":
		return "int"
	case "bool":
		return "bool"
	case "float", "double":
		return "float"
	case "bytes":
		return "bytes"
	default:
		return protoType
	}
}

// jsonType converts proto type to JSON type.
func jsonType(protoType string) string {
	switch protoType {
	case "string":
		return "string"
	case "int32", "int64", "uint32", "uint64":
		return "number"
	case "bool":
		return "boolean"
	case "float", "double":
		return "number"
	case "bytes":
		return "string (base64)"
	default:
		return "object"
	}
}
