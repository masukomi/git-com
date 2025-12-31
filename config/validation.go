package config

import (
	"fmt"
	"strings"

	"git-com/output"
)

// ValidateConfig validates all elements in the configuration
// Returns true if all elements are valid, false otherwise
// Prints error messages to stderr for invalid elements
func ValidateConfig(cfg *Config) bool {
	valid := true
	hasTitleElement := false

	for _, elem := range cfg.Elements {
		if elem.Destination == DestTitle {
			hasTitleElement = true
		}
		if err := validateElement(elem); err != nil {
			output.PrintError(fmt.Sprintf("\"%s\" was not configured correctly in .git-com.yaml", elem.Name))
			valid = false
		}
	}

	if !hasTitleElement {
		output.PrintError("At least one element in .git-com.yaml must have destination: title")
		valid = false
	}

	return valid
}

// validateElement validates a single element based on its type
func validateElement(elem Element) error {
	// Check required destination
	if elem.Destination != DestTitle && elem.Destination != DestBody {
		return fmt.Errorf("invalid destination: %s", elem.Destination)
	}

	// Infer type from data-type if not specified
	elemType := elem.Type
	if elemType == "" && elem.DataType != "" {
		elemType = TypeText
	}

	// Check required type
	if elemType == "" {
		return fmt.Errorf("missing type")
	}

	// Validate based on type
	switch elemType {
	case TypeText:
		return validateTextElement(elem)
	case TypeMultilineText:
		return nil // No additional requirements
	case TypeSelect:
		return validateSelectElement(elem)
	case TypeMultiSelect:
		return validateMultiSelectElement(elem)
	case TypeConfirmation:
		return nil // No additional requirements
	default:
		return fmt.Errorf("unknown type: %s", elemType)
	}
}

// validateTextElement validates a text element
func validateTextElement(elem Element) error {
	// Validate data-type if present
	if elem.DataType != "" {
		switch elem.DataType {
		case DataTypeString, DataTypeInteger, DataTypeFloat:
			// Valid
		default:
			return fmt.Errorf("invalid data-type: %s", elem.DataType)
		}
	}
	return nil
}

// validateSelectElement validates a select element
func validateSelectElement(elem Element) error {
	if len(elem.Options) == 0 {
		return fmt.Errorf("select element must have options")
	}
	return nil
}

// validateMultiSelectElement validates a multi-select element
func validateMultiSelectElement(elem Element) error {
	if len(elem.Options) == 0 {
		return fmt.Errorf("multi-select element must have options")
	}
	if elem.RecordAs == "" {
		return fmt.Errorf("multi-select element must have record-as")
	}
	if elem.RecordAs != RecordAsList && elem.RecordAs != RecordAsJoinedString {
		return fmt.Errorf("invalid record-as: %s", elem.RecordAs)
	}
	// Cannot define empty-selection-text if allow-empty is false or not present
	if elem.HasEmptySelectionText() && !elem.IsAllowEmpty() {
		return fmt.Errorf("cannot define empty-selection-text when allow-empty is false or not set")
	}
	return nil
}

// GetEffectiveType returns the effective type of an element
// (handles inference from data-type)
func GetEffectiveType(elem Element) ElementType {
	if elem.Type == "" && elem.DataType != "" {
		return TypeText
	}
	return elem.Type
}
