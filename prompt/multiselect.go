package prompt

import (
	"strconv"
	"strings"

	"git-com/config"
	"git-com/output"
)

// HandleMultiSelect processes a multi-select element
func HandleMultiSelect(elem config.Element) (string, error) {
	// Determine the empty selection text if allow-empty is true
	var emptySelectionText string
	if elem.IsAllowEmpty() {
		emptySelectionText = elem.GetEmptySelectionText()
	}

	for {
		// Display instructions if present
		DisplayInstructions(elem.Instructions)

		// Display hint for multi-select
		DisplayHint("Use Space to select multiple. Hit Enter to submit.")

		// Build options list
		options := make([]string, 0, len(elem.Options)+1)
		if emptySelectionText != "" {
			// Add empty selection option at the top
			options = append(options, emptySelectionText)
		}
		options = append(options, elem.Options...)

		// Get selections
		selections, err := runMultiSelect(options, elem.Limit)
		if err != nil {
			if err == ErrUserAborted {
				return "", ErrUserAborted
			}
			return "", err
		}

		// Check if user selected the empty selection option
		if emptySelectionText != "" && containsEmptySelection(selections, emptySelectionText) {
			// Treat as empty response
			return "", nil
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

// containsEmptySelection checks if the empty selection text is in the selections
func containsEmptySelection(selections []string, emptySelectionText string) bool {
	for _, sel := range selections {
		if sel == emptySelectionText {
			return true
		}
	}
	return false
}

// runMultiSelect runs gum choose with multi-select and returns the results
func runMultiSelect(options []string, limit int) ([]string, error) {
	args := []string{"choose"}

	if limit > 0 {
		args = append(args, "--limit", strconv.Itoa(limit))
	} else {
		args = append(args, "--no-limit")
	}

	args = append(args, options...)

	result, err := runGumCommand(args...)
	if err != nil {
		return nil, err
	}

	if result == "" {
		return nil, nil
	}

	// Split by newline (gum's default output delimiter)
	return strings.Split(result, "\n"), nil
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
