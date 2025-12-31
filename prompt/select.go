package prompt

import (
	"git-com/config"
	"git-com/output"
	"git-com/tui"
)

// HandleSelect processes a select element
func HandleSelect(elem config.Element, cfg *config.Config) (string, error) {
	for {
		// Display instructions if present
		DisplayInstructions(elem.Instructions)

		// Build options list
		options := make([]string, len(elem.Options))
		copy(options, elem.Options)

		// Add "Other…" if modifiable
		if elem.IsModifiable() {
			options = append(options, otherOption)
		}

		// Get selection (limit 1 for single select)
		selected, err := tui.Choose(options, 1)
		if err != nil {
			if isAbortError(err) {
				return "", ErrUserAborted
			}
			return "", err
		}

		var result string
		if len(selected) > 0 {
			result = selected[0]
		}

		// Handle "Other…" selection
		if result == otherOption {
			newValue, err := handleOtherSelection(elem.Name, cfg)
			if err != nil {
				if err == ErrUserAborted {
					// User cancelled, re-show the select
					continue
				}
				return "", err
			}
			result = newValue
		}

		// Check if empty is allowed
		if result == "" && !elem.IsAllowEmpty() {
			output.PrintWarning("This input is required.")
			continue
		}

		return result, nil
	}
}
