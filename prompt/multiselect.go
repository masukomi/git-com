package prompt

import (
	"strings"

	"git-com/config"
	"git-com/output"
	"git-com/tui"
)

// HandleMultiSelect processes a multi-select element
func HandleMultiSelect(elem config.Element, cfg *config.Config) (string, error) {
	// Determine the empty selection text if allow-empty is true
	var emptySelectionText string
	if elem.IsAllowEmpty() {
		emptySelectionText = elem.GetEmptySelectionText()
	}

	for {
		// Display instructions if present
		DisplayInstructions(elem.Instructions)

		// Build options list
		options := make([]string, 0, len(elem.Options)+2)
		if emptySelectionText != "" {
			// Add empty selection option at the top
			options = append(options, emptySelectionText)
		}
		options = append(options, elem.Options...)

		// Add "Other…" if modifiable
		if elem.IsModifiable() {
			options = append(options, otherOption)
		}

		// Get selections (limit 0 means no limit, or use elem.Limit)
		limit := elem.Limit
		if limit == 0 {
			limit = -1 // No limit in our tui.Choose
		}
		selections, err := tui.Choose(options, limit)
		if err != nil {
			if isAbortError(err) {
				return "", ErrUserAborted
			}
			return "", err
		}

		// Check if user selected the empty selection option
		if emptySelectionText != "" && containsEmptySelection(selections, emptySelectionText) {
			// Treat as empty response
			return "", nil
		}

		// Handle "Other…" selection
		if elem.IsModifiable() && containsOption(selections, otherOption) {
			newValue, err := handleOtherSelection(elem.Name, cfg)
			if err != nil {
				if err == ErrUserAborted {
					// User cancelled, re-show the multi-select
					continue
				}
				return "", err
			}
			// Replace "Other…" with the new value in selections
			selections = replaceOption(selections, otherOption, newValue)
		}

		// Check if empty is allowed
		if len(selections) == 0 && !elem.IsAllowEmpty() {
			output.PrintWarning("This input is required.")
			continue
		}

		// Format the result based on record-as
		result := formatMultiSelectResult(selections, elem)
		return result, nil
	}
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
func formatAsList(selections []string, bullet string) string {
	var lines []string
	for _, sel := range selections {
		lines = append(lines, bullet+sel)
	}
	return strings.Join(lines, "\n")
}

// formatAsJoinedString formats selections as a joined string
func formatAsJoinedString(selections []string, joiner string) string {
	return strings.Join(selections, joiner)
}
