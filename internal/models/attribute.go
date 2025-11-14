package models

import (
	"encoding/json"
	"fmt"
	"time"
)

// AttributeValue represents a typed attribute value
type AttributeValue struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

// ValidAttributeTypes are the allowed attribute types
var ValidAttributeTypes = []string{"string", "number", "boolean", "date"}

// ValidateAttributeValue validates that an attribute value matches its declared type
func ValidateAttributeValue(attr AttributeValue) error {
	// Validate type is allowed
	validType := false
	for _, t := range ValidAttributeTypes {
		if attr.Type == t {
			validType = true
			break
		}
	}
	if !validType {
		return fmt.Errorf("invalid attribute type: %s (must be one of: string, number, boolean, date)", attr.Type)
	}

	// Validate value matches type
	switch attr.Type {
	case "string":
		if _, ok := attr.Value.(string); !ok {
			return fmt.Errorf("attribute value must be a string for type 'string'")
		}
	case "number":
		// JSON numbers can be float64 or int
		switch attr.Value.(type) {
		case float64, int, int64, int32:
			// Valid
		default:
			return fmt.Errorf("attribute value must be a number for type 'number'")
		}
	case "boolean":
		if _, ok := attr.Value.(bool); !ok {
			return fmt.Errorf("attribute value must be a boolean for type 'boolean'")
		}
	case "date":
		// Date should be a string in ISO 8601 format
		dateStr, ok := attr.Value.(string)
		if !ok {
			return fmt.Errorf("attribute value must be a string (ISO 8601 date) for type 'date'")
		}
		// Try to parse as ISO 8601 date
		if _, err := time.Parse("2006-01-02", dateStr); err != nil {
			// Also try parsing as full timestamp
			if _, err := time.Parse(time.RFC3339, dateStr); err != nil {
				return fmt.Errorf("attribute value must be a valid ISO 8601 date string")
			}
		}
	}

	return nil
}

// ValidateAttributes validates all attributes in a map
func ValidateAttributes(attributes map[string]interface{}) error {
	for key, value := range attributes {
		// Each attribute value should be a map with "type" and "value" keys
		attrMap, ok := value.(map[string]interface{})
		if !ok {
			return fmt.Errorf("attribute '%s' must be an object with 'type' and 'value' fields", key)
		}

		// Extract type and value
		attrType, ok := attrMap["type"].(string)
		if !ok {
			return fmt.Errorf("attribute '%s' must have a 'type' field", key)
		}

		attrValue, ok := attrMap["value"]
		if !ok {
			return fmt.Errorf("attribute '%s' must have a 'value' field", key)
		}

		// Validate the attribute
		attr := AttributeValue{
			Type:  attrType,
			Value: attrValue,
		}
		if err := ValidateAttributeValue(attr); err != nil {
			return fmt.Errorf("attribute '%s': %w", key, err)
		}
	}

	return nil
}

// SanitizeAttributes converts attributes map to JSONB format
func SanitizeAttributes(attributes map[string]interface{}) (json.RawMessage, error) {
	if attributes == nil {
		return json.RawMessage("{}"), nil
	}

	// Validate attributes first
	if err := ValidateAttributes(attributes); err != nil {
		return nil, err
	}

	// Convert to JSON
	jsonData, err := json.Marshal(attributes)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal attributes: %w", err)
	}

	return jsonData, nil
}

