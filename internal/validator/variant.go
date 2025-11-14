package validator

import (
	"fmt"

	"apidemo1/internal/models"
)

// ValidateVariantCreate validates a variant create request
func ValidateVariantCreate(req *models.VariantCreateRequest) error {
	// Use built-in validation
	if err := req.Validate(); err != nil {
		return err
	}

	// Validate SKU format (same pattern as products)
	if err := ValidateSKU(req.SKU); err != nil {
		return err
	}

	// Validate attributes if provided
	if req.Attributes != nil {
		if err := models.ValidateAttributes(req.Attributes); err != nil {
			return models.NewValidationError(fmt.Sprintf("invalid attributes: %s", err.Error()))
		}
	}

	return nil
}

// ValidateVariantUpdate validates a variant update request
func ValidateVariantUpdate(req *models.VariantUpdateRequest) error {
	// Validate attributes if provided
	if req.Attributes != nil {
		if err := models.ValidateAttributes(req.Attributes); err != nil {
			return models.NewValidationError(fmt.Sprintf("invalid attributes: %s", err.Error()))
		}
	}

	return nil
}

