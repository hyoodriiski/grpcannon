package proto

import (
	"encoding/json"
	"fmt"
	"os"
)

// MethodDescriptor is the JSON-serialisable form of a method definition.
type MethodDescriptor struct {
	FullMethod string `json:"full_method"`
	InputType  string `json:"input_type"`
	OutputType string `json:"output_type"`
}

// SchemaFile is the top-level structure of a proto schema JSON file.
type SchemaFile struct {
	Methods []MethodDescriptor `json:"methods"`
}

// LoadSchema reads a JSON schema file and populates the given registry.
func LoadSchema(path string, r *Registry) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("proto: read schema: %w", err)
	}
	var sf SchemaFile
	if err := json.Unmarshal(data, &sf); err != nil {
		return fmt.Errorf("proto: parse schema: %w", err)
	}
	if len(sf.Methods) == 0 {
		return fmt.Errorf("proto: schema contains no methods")
	}
	for _, m := range sf.Methods {
		if err := r.Register(MethodInfo{
			FullMethod: m.FullMethod,
			InputType:  m.InputType,
			OutputType: m.OutputType,
		}); err != nil {
			return err
		}
	}
	return nil
}
