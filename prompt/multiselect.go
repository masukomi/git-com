package prompt

import (
	"strings"

	"git-com/config"
	"git-com/output"
	"git-com/tui"
)

// HandleMultiSelect processes a multi-select element
func HandleMultiSelect(elem config.Element, cfg *config.Config) (string, error) {
	options, emptyText := buildMultiSelectOptions(elem)
	limit := getMultiSelectLimit(elem)

	for {
		selections, err := tui.Choose(options, limit, elem.Instructions)
		if err != nil {
			if isAbortError(err) {
				return "", ErrUserAborted
			}
			return "", err
		}

		result, retry, err := processMultiSelectResult(selections, emptyText, elem, cfg)
		if err != nil {
			return "", err
		}
		if retry {
			continue
		}

		return result, nil
	}
}

// buildMultiSelectOptions builds the options list with optional empty selection and "Other"
func buildMultiSelectOptions(elem config.Element) (options []string, emptyText string) {
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

// getMultiSelectLimit returns the selection limit for the element
func getMultiSelectLimit(elem config.Element) int {
	if elem.Limit == 0 {
		return -1 // No limit in tui.Choose
	}
	return elem.Limit
}

// processMultiSelectResult processes the user's selections
// Returns (result, shouldRetry, error)
func processMultiSelectResult(selections []string, emptyText string, elem config.Element, cfg *config.Config) (string, bool, error) {
	// Check if user selected the empty selection option
	if emptyText != "" && containsOption(selections, emptyText) {
		return "", false, nil
	}

	// Handle "Other…" selection
	if elem.IsModifiable() && containsOption(selections, otherOption) {
		newValue, err := handleOtherSelection(elem.Name, cfg)
		if err == ErrUserAborted {
			return "", true, nil // Retry
		}
		if err != nil {
			return "", false, err
		}
		selections = replaceOption(selections, otherOption, newValue)
	}

	// Check if empty is allowed
	if len(selections) == 0 && !elem.IsAllowEmpty() {
		output.PrintWarning("This input is required.")
		return "", true, nil // Retry
	}

	return formatMultiSelectResult(selections, elem), false, nil
}

// containsOption checks if a specific option is in the selections
func containsOption(selections []string, option string) bool {
	for _, sel := range selections {
		if sel == option {
			return true
		}
	}
	return false
}

// containsEmptySelection checks if the empty selection text is in the selections
func containsEmptySelection(selections []string, emptySelectionText string) bool {
	return containsOption(selections, emptySelectionText)
}

// replaceOption replaces an option in selections with a new value
func replaceOption(selections []string, oldOption, newValue string) []string {
	result := make([]string, 0, len(selections))
	for _, sel := range selections {
		if sel == oldOption {
			result = append(result, newValue)
		} else {
			result = append(result, sel)
		}
	}
	return result
}

// formatMultiSelectResult formats the selected items based on record-as setting
func formatMultiSelectResult(selections []string, elem config.Element) string {
	if len(selections) == 0 {
		return ""
	}

	switch elem.RecordAs {
	case config.RecordAsList:
		return formatAsList(selections, elem.GetBulletString())
	case config.RecordAsJoinedString:
		return formatAsJoinedString(selections, elem.GetJoinString())
	default:
		// Default to joined string
		return formatAsJoinedString(selections, elem.GetJoinString())
	}
}

// formatAsList formats selections as a bulleted list
// Adds a leading newline (to separate from before-string) and a trailing newline
func formatAsList(selections []string, bullet string) string {
	var lines []string
	for _, sel := range selections {
		lines = append(lines, bullet+sel)
	}
	return "\n" + strings.Join(lines, "\n") + "\n"
}

// formatAsJoinedString formats selections as a joined string
func formatAsJoinedString(selections []string, joiner string) string {
	return strings.Join(selections, joiner)
}
