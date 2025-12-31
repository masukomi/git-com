package prompt

import (
	"strings"

	"git-com/config"
	"git-com/output"
)

const otherOption = "Other..."

// HandleSelect processes a select element
func HandleSelect(elem config.Element, cfg *config.Config) (string, error) {
	for {
		// Display instructions if present
		DisplayInstructions(elem.Instructions)

		// Build options list
		options := make([]string, len(elem.Options))
		copy(options, elem.Options)

		// Add "Other..." if modifiable
		if elem.IsModifiable() {
			options = append(options, otherOption)
		}

		// Get selection
		result, err := runSelect(options)
		if err != nil {
			if err == ErrUserAborted {
				return "", ErrUserAborted
			}
			return "", err
		}

		result = strings.TrimSpace(result)

		// Handle "Other..." selection
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

// runSelect runs gum choose with limit 1 and returns the result
func runSelect(options []string) (string, error) {
	args := []string{"choose", "--limit", "1"}
	args = append(args, options...)
	return runGumCommand(args...)
}

// handleOtherSelection handles when user selects "Other..." to add a new item
func handleOtherSelection(elementName string, cfg *config.Config) (string, error) {
	ClearScreen()
	DisplayInstructions("Add & select a new item")

	result, err := runGumCommand("input", "--placeholder", "Enter new option...")
	if err != nil {
		return "", err
	}

	newValue := strings.TrimSpace(result)
	if newValue == "" {
		return "", ErrUserAborted
	}

	// Save the new option to the config file
	if cfg != nil {
		if err := cfg.AddOptionToElement(elementName, newValue); err != nil {
			// Log the error but don't fail - the value is still usable
			output.PrintWarning("Could not save new option to config: " + err.Error())
		}
	}

	return newValue, nil
}
