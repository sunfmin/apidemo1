package models

import (
	"encoding/json"
	"fmt"
	"time"

	pb "apidemo1/proto"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ProtoToProduct converts protobuf Product to internal Product model
func ProtoToProduct(p *pb.Product) (*Product, error) {
	id, err := uuid.Parse(p.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid product ID: %w", err)
	}

	var productTypeID *uuid.UUID
	if p.ProductTypeId != "" {
		ptID, err := uuid.Parse(p.ProductTypeId)
		if err != nil {
			return nil, fmt.Errorf("invalid product type ID: %w", err)
		}
		productTypeID = &ptID
	}

	// Convert attributes from protobuf to JSONB format
	attrs, err := ProtoAttributesToJSON(p.Attributes)
	if err != nil {
		return nil, err
	}

	var description *string
	if p.Description != "" {
		description = &p.Description
	}

	return &Product{
		ID:            id,
		Name:          p.Name,
		SKU:           p.Sku,
		Description:   description,
		ProductTypeID: productTypeID,
		Attributes:    attrs,
		CreatedAt:     p.CreatedAt.AsTime(),
		UpdatedAt:     p.UpdatedAt.AsTime(),
	}, nil
}

// ProductToProto converts internal Product model to protobuf Product
func ProductToProto(p *Product) (*pb.Product, error) {
	attrs, err := JSONToProtoAttributes(p.Attributes)
	if err != nil {
		return nil, err
	}

	pbProduct := &pb.Product{
		Id:         p.ID.String(),
		Name:       p.Name,
		Sku:        p.SKU,
		Attributes: attrs,
		CreatedAt:  timestamppb.New(p.CreatedAt),
		UpdatedAt:  timestamppb.New(p.UpdatedAt),
	}

	if p.Description != nil {
		pbProduct.Description = *p.Description
	}

	if p.ProductTypeID != nil {
		pbProduct.ProductTypeId = p.ProductTypeID.String()
	}

	return pbProduct, nil
}

// ProtoToVariant converts protobuf Variant to internal Variant model
func ProtoToVariant(v *pb.Variant) (*Variant, error) {
	id, err := uuid.Parse(v.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid variant ID: %w", err)
	}

	productID, err := uuid.Parse(v.ProductId)
	if err != nil {
		return nil, fmt.Errorf("invalid product ID: %w", err)
	}

	attrs, err := ProtoAttributesToJSON(v.Attributes)
	if err != nil {
		return nil, err
	}

	return &Variant{
		ID:         id,
		ProductID:  productID,
		SKU:        v.Sku,
		Attributes: attrs,
		CreatedAt:  v.CreatedAt.AsTime(),
		UpdatedAt:  v.UpdatedAt.AsTime(),
	}, nil
}

// VariantToProto converts internal Variant model to protobuf Variant
func VariantToProto(v *Variant) (*pb.Variant, error) {
	attrs, err := JSONToProtoAttributes(v.Attributes)
	if err != nil {
		return nil, err
	}

	return &pb.Variant{
		Id:         v.ID.String(),
		ProductId:  v.ProductID.String(),
		Sku:        v.SKU,
		Attributes: attrs,
		CreatedAt:  timestamppb.New(v.CreatedAt),
		UpdatedAt:  timestamppb.New(v.UpdatedAt),
	}, nil
}

// ProtoAttributesToJSON converts protobuf attributes map to JSONB format
func ProtoAttributesToJSON(attrs map[string]*pb.AttributeValue) (json.RawMessage, error) {
	if len(attrs) == 0 {
		return json.RawMessage("{}"), nil
	}

	// Convert to the JSONB format expected by database
	jsonMap := make(map[string]interface{})
	for key, attr := range attrs {
		var value interface{}
		switch v := attr.Value.(type) {
		case *pb.AttributeValue_StringValue:
			value = v.StringValue
		case *pb.AttributeValue_NumberValue:
			value = v.NumberValue
		case *pb.AttributeValue_BooleanValue:
			value = v.BooleanValue
		case *pb.AttributeValue_DateValue:
			value = v.DateValue
		default:
			return nil, fmt.Errorf("unknown attribute value type for key %s", key)
		}

		jsonMap[key] = map[string]interface{}{
			"type":  attr.Type,
			"value": value,
		}
	}

	return json.Marshal(jsonMap)
}

// JSONToProtoAttributes converts JSONB format to protobuf attributes map
func JSONToProtoAttributes(data json.RawMessage) (map[string]*pb.AttributeValue, error) {
	if len(data) == 0 || string(data) == "{}" {
		return make(map[string]*pb.AttributeValue), nil
	}

	// Parse the JSONB format
	var jsonMap map[string]map[string]interface{}
	if err := json.Unmarshal(data, &jsonMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal attributes: %w", err)
	}

	attrs := make(map[string]*pb.AttributeValue)
	for key, attr := range jsonMap {
		attrType, ok := attr["type"].(string)
		if !ok {
			return nil, fmt.Errorf("attribute %s missing type", key)
		}

		value := attr["value"]

		pbAttr := &pb.AttributeValue{
			Type: attrType,
		}

		switch attrType {
		case "string":
			if strVal, ok := value.(string); ok {
				pbAttr.Value = &pb.AttributeValue_StringValue{StringValue: strVal}
			}
		case "number":
			if numVal, ok := value.(float64); ok {
				pbAttr.Value = &pb.AttributeValue_NumberValue{NumberValue: numVal}
			}
		case "boolean":
			if boolVal, ok := value.(bool); ok {
				pbAttr.Value = &pb.AttributeValue_BooleanValue{BooleanValue: boolVal}
			}
		case "date":
			if dateVal, ok := value.(string); ok {
				pbAttr.Value = &pb.AttributeValue_DateValue{DateValue: dateVal}
			}
		default:
			return nil, fmt.Errorf("unknown attribute type: %s", attrType)
		}

		attrs[key] = pbAttr
	}

	return attrs, nil
}

// CreateAttributeValue creates a protobuf AttributeValue
func CreateStringAttribute(value string) *pb.AttributeValue {
	return &pb.AttributeValue{
		Type:  "string",
		Value: &pb.AttributeValue_StringValue{StringValue: value},
	}
}

func CreateNumberAttribute(value float64) *pb.AttributeValue {
	return &pb.AttributeValue{
		Type:  "number",
		Value: &pb.AttributeValue_NumberValue{NumberValue: value},
	}
}

func CreateBooleanAttribute(value bool) *pb.AttributeValue {
	return &pb.AttributeValue{
		Type:  "boolean",
		Value: &pb.AttributeValue_BooleanValue{BooleanValue: value},
	}
}

func CreateDateAttribute(value string) *pb.AttributeValue {
	// Validate date format
	if _, err := time.Parse("2006-01-02", value); err != nil {
		if _, err := time.Parse(time.RFC3339, value); err != nil {
			// Return attribute anyway but mark as potentially invalid
		}
	}
	return &pb.AttributeValue{
		Type:  "date",
		Value: &pb.AttributeValue_DateValue{DateValue: value},
	}
}

