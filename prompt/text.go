package prompt

import (
	"regexp"
	"strings"

	"git-com/config"
	"git-com/output"
)

var (
	integerRegex = regexp.MustCompile(`^\d+$`)
	floatRegex   = regexp.MustCompile(`^\d+\.\d+$`)
)

// HandleText processes a text input element
func HandleText(elem config.Element) (string, error) {
	for {
		// Display instructions if present
		DisplayInstructions(elem.Instructions)

		// Get text input
		result, err := runTextInput(elem.Placeholder)
		if err != nil {
			if err == ErrUserAborted {
				return "", ErrUserAborted
			}
			return "", err
		}

		// Trim whitespace
		result = strings.TrimSpace(result)

		// Validate data type if specified
		if elem.DataType != "" && result != "" {
			if !validateDataType(result, elem.DataType) {
				output.PrintError("Your input must be a " + string(elem.DataType))
				continue
			}
		}

		// Check if empty is allowed
		if result == "" && !elem.IsAllowEmpty() {
			output.PrintWarning("This input is required.")
			continue
		}

		return result, nil
	}
}

// runTextInput runs gum input and returns the result
func runTextInput(placeholder string) (string, error) {
	args := []string{"input"}
	if placeholder != "" {
		args = append(args, "--placeholder", placeholder)
	}
	return runGumCommand(args...)
}

// validateDataType validates input against the specified data type
func validateDataType(value string, dataType config.DataType) bool {
	switch dataType {
	case config.DataTypeInteger:
		return integerRegex.MatchString(value)
	case config.DataTypeFloat:
		return floatRegex.MatchString(value)
	case config.DataTypeString:
		return true
	default:
		return true
	}
}
