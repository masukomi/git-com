package prompt

import (
	"git-com/config"
	"git-com/output"
	"git-com/tui"
)

// HandleSelect processes a select element
func HandleSelect(elem config.Element, cfg *config.Config) (string, error) {
	options := buildSelectOptions(elem)

	for {
		selected, err := tui.Choose(options, 1, elem.Instructions)
		if err != nil {
			if isAbortError(err) {
				return "", ErrUserAborted
			}
			return "", err
		}

		result, retry, err := processSelectResult(selected, elem, cfg)
		if err != nil {
			return "", err
		}
		if retry {
			continue
		}

		return result, nil
	}
}

// buildSelectOptions builds the options list with optional "Other"
func buildSelectOptions(elem config.Element) []string {
	options := make([]string, len(elem.Options))
	copy(options, elem.Options)

	if elem.IsModifiable() {
		options = append(options, otherOption)
	}

	return options
}

// processSelectResult processes the user's selection
// Returns (result, shouldRetry, error)
func processSelectResult(selected []string, elem config.Element, cfg *config.Config) (string, bool, error) {
	var result string
	if len(selected) > 0 {
		result = selected[0]
	}

	// Handle "Other…" selection
	if result == otherOption {
		newValue, err := handleOtherSelection(elem.Name, cfg)
		if err == ErrUserAborted {
			return "", true, nil // Retry
		}
		if err != nil {
			return "", false, err
		}
		result = newValue
	}

	// Check if empty is allowed
	if result == "" && !elem.IsAllowEmpty() {
		output.PrintWarning("This input is required.")
		return "", true, nil // Retry
	}

	return result, false, nil
}
