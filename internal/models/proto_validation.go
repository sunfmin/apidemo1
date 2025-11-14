package models

import (
	"fmt"

	pb "apidemo1/proto"
)

// ValidateProtoAttributes validates protobuf attributes
func ValidateProtoAttributes(attrs map[string]*pb.AttributeValue) error {
	if attrs == nil {
		return nil
	}

	for key, attr := range attrs {
		if attr == nil {
			return fmt.Errorf("attribute '%s' cannot be nil", key)
		}

		// Validate type
		validType := false
		for _, t := range ValidAttributeTypes {
			if attr.Type == t {
				validType = true
				break
			}
		}
		if !validType {
			return fmt.Errorf("attribute '%s' has invalid type: %s (must be one of: string, number, boolean, date)", key, attr.Type)
		}

		// Validate value exists and matches type
		if attr.Value == nil {
			return fmt.Errorf("attribute '%s' must have a value", key)
		}

		switch attr.Type {
		case "string":
			if _, ok := attr.Value.(*pb.AttributeValue_StringValue); !ok {
				return fmt.Errorf("attribute '%s' type is 'string' but value is not StringValue", key)
			}
		case "number":
			if _, ok := attr.Value.(*pb.AttributeValue_NumberValue); !ok {
				return fmt.Errorf("attribute '%s' type is 'number' but value is not NumberValue", key)
			}
		case "boolean":
			if _, ok := attr.Value.(*pb.AttributeValue_BooleanValue); !ok {
				return fmt.Errorf("attribute '%s' type is 'boolean' but value is not BooleanValue", key)
			}
		case "date":
			if _, ok := attr.Value.(*pb.AttributeValue_DateValue); !ok {
				return fmt.Errorf("attribute '%s' type is 'date' but value is not DateValue", key)
			}
		}
	}

	return nil
}

// ValidateProductCreateRequest validates protobuf ProductCreateRequest
func ValidateProductCreateRequest(req *pb.ProductCreateRequest) error {
	if req.Name == "" {
		return NewValidationError("name is required")
	}
	if len(req.Name) > 500 {
		return NewValidationError("name must be at most 500 characters")
	}
	if req.Sku == "" {
		return NewValidationError("sku is required")
	}
	if len(req.Sku) > 100 {
		return NewValidationError("sku must be at most 100 characters")
	}
	if len(req.Description) > 10000 {
		return NewValidationError("description must be at most 10000 characters")
	}

	// Validate attributes
	if err := ValidateProtoAttributes(req.Attributes); err != nil {
		return NewValidationError(err.Error())
	}

	return nil
}

// ValidateProductUpdateRequest validates protobuf ProductUpdateRequest
func ValidateProductUpdateRequest(req *pb.ProductUpdateRequest) error {
	if req.Name != "" && len(req.Name) > 500 {
		return NewValidationError("name must be at most 500 characters")
	}
	if len(req.Description) > 10000 {
		return NewValidationError("description must be at most 10000 characters")
	}

	// Validate attributes
	if err := ValidateProtoAttributes(req.Attributes); err != nil {
		return NewValidationError(err.Error())
	}

	return nil
}

// ValidateVariantCreateRequest validates protobuf VariantCreateRequest
func ValidateVariantCreateRequest(req *pb.VariantCreateRequest) error {
	if req.Sku == "" {
		return NewValidationError("sku is required")
	}
	if len(req.Sku) > 100 {
		return NewValidationError("sku must be at most 100 characters")
	}

	// Validate attributes
	if err := ValidateProtoAttributes(req.Attributes); err != nil {
		return NewValidationError(err.Error())
	}

	return nil
}

// ValidateVariantUpdateRequest validates protobuf VariantUpdateRequest  
func ValidateVariantUpdateRequest(req *pb.VariantUpdateRequest) error {
	// Validate attributes
	if err := ValidateProtoAttributes(req.Attributes); err != nil {
		return NewValidationError(err.Error())
	}

	return nil
}

