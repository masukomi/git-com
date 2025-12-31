package prompt

import (
	"git-com/config"
)

// Result holds the final commit message components
type Result struct {
	Title string
	Body  string
}

// ProcessElements processes all elements and builds the commit message
func ProcessElements(cfg *config.Config) (*Result, error) {
	result := &Result{
		Title: "",
		Body:  "",
	}

	for _, elem := range cfg.Elements {
		// Clear screen before each element
		ClearScreen()

		// Process element based on type
		value, err := processElement(elem, cfg)
		if err != nil {
			return nil, err
		}

		// Skip if value is empty
		if value == "" {
			continue
		}

		// Apply before-string and after-string
		finalValue := applyDecorators(value, elem)

		// Append to appropriate destination
		switch elem.Destination {
		case config.DestTitle:
			result.Title += finalValue
		case config.DestBody:
			result.Body += finalValue
		}
	}

	return result, nil
}

// processElement routes to the appropriate handler based on element type
func processElement(elem config.Element, cfg *config.Config) (string, error) {
	// Get effective type (handles inference from data-type)
	elemType := config.GetEffectiveType(elem)

	switch elemType {
	case config.TypeText:
		return HandleText(elem)
	case config.TypeMultilineText:
		return HandleMultilineText(elem)
	case config.TypeSelect:
		return HandleSelect(elem, cfg)
	case config.TypeMultiSelect:
		return HandleMultiSelect(elem)
	case config.TypeConfirmation:
		return HandleConfirmation(elem)
	default:
		// Fallback to text input
		return HandleText(elem)
	}
}

// applyDecorators applies before-string and after-string to a value
func applyDecorators(value string, elem config.Element) string {
	return elem.BeforeString + value + elem.AfterString
}
