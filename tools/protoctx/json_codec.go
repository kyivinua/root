package protoctx

import (
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

// JSONMarshalOptions defines options for JSON marshaling.
var JSONMarshalOptions = protojson.MarshalOptions{
	UseProtoNames:   true,  // Use field names as defined in .proto
	EmitUnpopulated: false, // Don't emit zero values
	Indent:          "  ",  // Two-space indentation for readability
}

// JSONUnmarshalOptions defines options for JSON unmarshaling.
var JSONUnmarshalOptions = protojson.UnmarshalOptions{
	DiscardUnknown: true, // Ignore unknown fields (security best practice)
}

// NewDynamicMessage creates a new dynamic message by FQN.
//
// This is useful when you need to create a message instance without
// generated Go code.
//
// Example:
//
//	msg, err := ctx.NewDynamicMessage("company.user.v1.UserProfile")
func (c *Context) NewDynamicMessage(fqn string) (*dynamicpb.Message, error) {
	md, err := c.FindMessageDescriptor(fqn)
	if err != nil {
		return nil, err
	}

	return dynamicpb.NewMessage(md), nil
}

// UnmarshalWireToDynamic unmarshals wire-format bytes to a dynamic message.
//
// Example:
//
//	msg, err := ctx.UnmarshalWireToDynamic("company.user.v1.UserProfile", wireBytes)
func (c *Context) UnmarshalWireToDynamic(fqn string, wire []byte) (*dynamicpb.Message, error) {
	msg, err := c.NewDynamicMessage(fqn)
	if err != nil {
		return nil, err
	}

	if err := proto.Unmarshal(wire, msg); err != nil {
		return nil, fmt.Errorf("unmarshal wire to %s: %w", fqn, err)
	}

	return msg, nil
}

// MarshalDynamicToWire marshals a dynamic message to wire-format bytes.
func (c *Context) MarshalDynamicToWire(msg *dynamicpb.Message) ([]byte, error) {
	wire, err := proto.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("marshal dynamic message: %w", err)
	}

	return wire, nil
}

// MarshalWireToJSON converts wire-format bytes to JSON.
//
// This is the most common operation for debugging and API gateways.
//
// Example:
//
//	jsonData, err := ctx.MarshalWireToJSON("company.user.v1.UserProfile", wireBytes)
func (c *Context) MarshalWireToJSON(fqn string, wire []byte) ([]byte, error) {
	msg, err := c.UnmarshalWireToDynamic(fqn, wire)
	if err != nil {
		return nil, err
	}

	jsonData, err := JSONMarshalOptions.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("marshal to JSON: %w", err)
	}

	return jsonData, nil
}

// UnmarshalJSONToWire converts JSON to wire-format bytes.
//
// Example:
//
//	wireBytes, err := ctx.UnmarshalJSONToWire("company.user.v1.UserProfile", jsonData)
func (c *Context) UnmarshalJSONToWire(fqn string, jsonData []byte) ([]byte, error) {
	msg, err := c.NewDynamicMessage(fqn)
	if err != nil {
		return nil, err
	}

	if err := JSONUnmarshalOptions.Unmarshal(jsonData, msg); err != nil {
		return nil, fmt.Errorf("unmarshal JSON to %s: %w", fqn, err)
	}

	return c.MarshalDynamicToWire(msg)
}

// MarshalMessageToJSON marshals a protoreflect.Message to JSON.
//
// This is a convenience method for messages that implement protoreflect.ProtoMessage.
func (c *Context) MarshalMessageToJSON(msg protoreflect.ProtoMessage) ([]byte, error) {
	jsonData, err := JSONMarshalOptions.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("marshal message to JSON: %w", err)
	}

	return jsonData, nil
}

// UnmarshalJSONToMessage unmarshals JSON to a protoreflect.Message.
func (c *Context) UnmarshalJSONToMessage(jsonData []byte, msg protoreflect.ProtoMessage) error {
	if err := JSONUnmarshalOptions.Unmarshal(jsonData, msg); err != nil {
		return fmt.Errorf("unmarshal JSON to message: %w", err)
	}

	return nil
}
