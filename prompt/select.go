package prompt

import (
	"git-com/config"
	"git-com/output"
	"git-com/tui"
)

// HandleSelect processes a select element
func HandleSelect(elem config.Element, cfg *config.Config) (string, error) {
	options, emptyText := buildSelectOptions(elem)

	for {
		selected, err := tui.Choose(options, 1, elem.Instructions)
		if err != nil {
			if isAbortError(err) {
				return "", ErrUserAborted
			}
			return "", err
		}

		result, retry, err := processSelectResult(selected, emptyText, elem, cfg)
		if err != nil {
			return "", err
		}
		if retry {
			continue
		}

		return result, nil
	}
}

// buildSelectOptions builds the options list with optional empty selection and "Other"
// Returns the options and the empty selection text (if any)
func buildSelectOptions(elem config.Element) (options []string, emptyText string) {
	options = make([]string, 0, len(elem.Options)+2)

	if elem.IsAllowEmpty() {
		emptyText = Italicize(elem.GetEmptySelectionText())
		options = append(options, emptyText)
	}

	options = append(options, elem.Options...)

	if elem.IsModifiable() {
		options = append(options, otherOption)
	}

	return options, emptyText
}

// processSelectResult processes the user's selection
// Returns (result, shouldRetry, error)
func processSelectResult(selected []string, emptyText string, elem config.Element, cfg *config.Config) (string, bool, error) {
	var result string
	if len(selected) > 0 {
		result = selected[0]
	}

	// Check if user selected the empty selection option
	if emptyText != "" && result == emptyText {
		return "", false, nil
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
