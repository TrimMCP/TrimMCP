package filter

import (
	"fmt"
	"strings"

	"github.com/slimmcp/slimmcp/pkg/types"
)

// ValidateFieldSelection validates a field selection specification
func ValidateFieldSelection(spec *types.FieldSelectionSpec) error {
	if spec == nil {
		return nil
	}

	// Cannot have both include and exclude
	if len(spec.Include) > 0 && len(spec.Exclude) > 0 {
		return types.ErrInvalidFieldSelection
	}

	// Validate field names in include list
	for _, field := range spec.Include {
		if err := validateFieldName(field); err != nil {
			return fmt.Errorf("invalid field in include list '%s': %w", field, err)
		}
	}

	// Validate field names in exclude list
	for _, field := range spec.Exclude {
		if err := validateFieldName(field); err != nil {
			return fmt.Errorf("invalid field in exclude list '%s': %w", field, err)
		}
	}

	// Validate max depth
	if spec.MaxDepth < 0 || spec.MaxDepth > 20 {
		return types.ErrInvalidMaxDepth
	}

	return nil
}

// validateFieldName validates a field name (supports dot notation)
func validateFieldName(field string) error {
	if field == "" {
		return types.ErrInvalidFieldName
	}

	// Check for invalid characters
	// Field names can contain letters, numbers, underscores, dots, and hyphens
	for _, char := range field {
		if !isValidFieldChar(char) {
			return types.ErrInvalidFieldName
		}
	}

	// If using dot notation, validate each part
	if strings.Contains(field, ".") {
		parts := strings.Split(field, ".")
		for _, part := range parts {
			if part == "" {
				return types.ErrInvalidFieldName
			}
		}
	}

	return nil
}

// isValidFieldChar checks if a character is valid in a field name
func isValidFieldChar(char rune) bool {
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9') ||
		char == '_' ||
		char == '-' ||
		char == '.'
}
