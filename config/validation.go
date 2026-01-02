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
			output.PrintError(fmt.Sprintf("\"%s\" was not configured correctly in .git-com.y[a]ml", elem.Name))
			valid = false
		}
	}

	if !hasTitleElement {
		output.PrintError("At least one element in .git-com.y[a]ml must have destination: title")
		valid = false
	}

	return valid
}

// validateElement validates a single element based on its type
func validateElement(elem Element) error {
	elemType := inferElementType(elem)

	if elemType == "" {
		return fmt.Errorf("missing type")
	}

	// Confirmation elements have different validation rules
	if elemType == TypeConfirmation {
		return validateConfirmationElement(elem)
	}

	if err := validateDestination(elem); err != nil {
		return err
	}

	if err := validateTitleConstraints(elem); err != nil {
		return err
	}

	return validateByType(elemType, elem)
}

// inferElementType returns the element type, inferring from data-type if needed
func inferElementType(elem Element) ElementType {
	if elem.Type == "" && elem.DataType != "" {
		return TypeText
	}
	return elem.Type
}

// validateConfirmationElement validates confirmation-specific rules
func validateConfirmationElement(elem Element) error {
	if elem.Destination != "" {
		return fmt.Errorf("confirmation elements cannot have a destination")
	}
	return nil
}

// validateDestination checks that the destination is valid for non-confirmation elements
func validateDestination(elem Element) error {
	if elem.Destination != DestTitle && elem.Destination != DestBody {
		return fmt.Errorf("invalid destination: %s", elem.Destination)
	}
	return nil
}

// validateTitleConstraints checks title-specific constraints (no newlines)
func validateTitleConstraints(elem Element) error {
	if elem.Destination != DestTitle {
		return nil
	}
	if strings.Contains(elem.BeforeString, "\n") {
		return fmt.Errorf("before-string cannot contain newlines for title destination")
	}
	if strings.Contains(elem.AfterString, "\n") {
		return fmt.Errorf("after-string cannot contain newlines for title destination")
	}
	return nil
}

// validateByType dispatches to type-specific validation
func validateByType(elemType ElementType, elem Element) error {
	switch elemType {
	case TypeText:
		return validateTextElement(elem)
	case TypeMultilineText:
		return nil
	case TypeSelect:
		return validateSelectElement(elem)
	case TypeMultiSelect:
		return validateMultiSelectElement(elem)
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
	// Title destination requires joined-string (list would have newlines)
	if elem.Destination == DestTitle && elem.RecordAs == RecordAsList {
		return fmt.Errorf("multi-select with destination title must use record-as: joined-string")
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
