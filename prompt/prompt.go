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
// If oldCommitMessage is not nil, it will be used to pre-fill the first
// multiline-text element with destination=body
func ProcessElements(cfg *config.Config, oldCommitMessage *string) (*Result, error) {
	result := &Result{
		Title: "",
		Body:  "",
	}

	for _, elem := range cfg.Elements {
		// Clear screen before each element
		ClearScreen()

		// Process element based on type
		value, err := processElement(elem, cfg, &oldCommitMessage)
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
// oldCommitMessage is a pointer to a pointer so we can set it to nil after use
func processElement(elem config.Element, cfg *config.Config, oldCommitMessage **string) (string, error) {
	// Get effective type (handles inference from data-type)
	elemType := config.GetEffectiveType(elem)

	switch elemType {
	case config.TypeText:
		return HandleText(elem)
	case config.TypeMultilineText:
		// Only pass oldCommitMessage if destination is body
		var initialContent *string
		if elem.Destination == config.DestBody && oldCommitMessage != nil && *oldCommitMessage != nil {
			initialContent = *oldCommitMessage
			// Set to nil after use so it's only used once
			*oldCommitMessage = nil
		}
		return HandleMultilineText(elem, initialContent)
	case config.TypeSelect:
		return HandleSelect(elem, cfg)
	case config.TypeMultiSelect:
		return HandleMultiSelect(elem, cfg)
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
