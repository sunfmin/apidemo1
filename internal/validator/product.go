package validator

import (
	"fmt"
	"html"
	"regexp"

	"apidemo1/internal/models"
)

var (
	// SKU pattern: alphanumeric, hyphens, underscores only
	skuPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

// ValidateProductCreate validates a product create request
func ValidateProductCreate(req *models.ProductCreateRequest) error {
	// Use built-in validation
	if err := req.Validate(); err != nil {
		return err
	}

	// Validate SKU format
	if !skuPattern.MatchString(req.SKU) {
		return models.NewValidationError("sku must contain only alphanumeric characters, hyphens, and underscores")
	}

	// Sanitize text fields for XSS
	req.Name = SanitizeText(req.Name)
	if req.Description != nil {
		sanitized := SanitizeText(*req.Description)
		req.Description = &sanitized
	}

	// Validate attributes if provided
	if req.Attributes != nil {
		if err := models.ValidateAttributes(req.Attributes); err != nil {
			return models.NewValidationError(fmt.Sprintf("invalid attributes: %s", err.Error()))
		}
	}

	return nil
}

// ValidateProductUpdate validates a product update request
func ValidateProductUpdate(req *models.ProductUpdateRequest) error {
	// Use built-in validation
	if err := req.Validate(); err != nil {
		return err
	}

	// Sanitize text fields for XSS
	if req.Name != nil {
		sanitized := SanitizeText(*req.Name)
		req.Name = &sanitized
	}
	if req.Description != nil {
		sanitized := SanitizeText(*req.Description)
		req.Description = &sanitized
	}

	// Validate attributes if provided
	if req.Attributes != nil {
		if err := models.ValidateAttributes(req.Attributes); err != nil {
			return models.NewValidationError(fmt.Sprintf("invalid attributes: %s", err.Error()))
		}
	}

	return nil
}

// SanitizeText sanitizes text input to prevent XSS attacks
func SanitizeText(text string) string {
	// Escape HTML special characters
	sanitized := html.EscapeString(text)
	
	// Remove any remaining script tags (defense in depth)
	sanitized = removeScriptTags(sanitized)
	
	return sanitized
}

// removeScriptTags removes <script> tags from text
func removeScriptTags(text string) string {
	// Case-insensitive removal of <script> tags
	re := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	return re.ReplaceAllString(text, "")
}

// ValidateSKU validates SKU format
func ValidateSKU(sku string) error {
	if sku == "" {
		return models.NewValidationError("sku is required")
	}
	if len(sku) > 100 {
		return models.NewValidationError("sku must be at most 100 characters")
	}
	if !skuPattern.MatchString(sku) {
		return models.NewValidationError("sku must contain only alphanumeric characters, hyphens, and underscores")
	}
	return nil
}

// ValidateName validates product name
func ValidateName(name string) error {
	if name == "" {
		return models.NewValidationError("name is required")
	}
	if len(name) > 500 {
		return models.NewValidationError("name must be at most 500 characters")
	}
	return nil
}

// ValidateDescription validates product description
func ValidateDescription(description *string) error {
	if description != nil && len(*description) > 10000 {
		return models.NewValidationError("description must be at most 10000 characters")
	}
	return nil
}

// ValidatePagination validates pagination parameters
func ValidatePagination(page, pageSize int) (int, int, error) {
	if page < 1 {
		return 0, 0, models.NewValidationError("page must be at least 1")
	}
	if pageSize < 1 {
		return 0, 0, models.NewValidationError("page_size must be at least 1")
	}
	if pageSize > 100 {
		return 0, 0, models.NewValidationError("page_size must be at most 100")
	}
	return page, pageSize, nil
}

