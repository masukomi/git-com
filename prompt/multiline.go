package prompt

import (
	"strings"

	"git-com/config"
	"git-com/output"
)

// HandleMultilineText processes a multiline text input element
func HandleMultilineText(elem config.Element) (string, error) {
	for {
		// Display instructions if present
		DisplayInstructions(elem.Instructions)

		// Display hint for multiline input
		DisplayHint("Ctrl+d to Submit")

		// Get multiline text input
		result, err := runMultilineInput()
		if err != nil {
			if err == ErrUserAborted {
				return "", ErrUserAborted
			}
			return "", err
		}

		// Trim whitespace
		result = strings.TrimSpace(result)

		// Check if empty is allowed
		if result == "" && !elem.IsAllowEmpty() {
			output.PrintWarning("This input is required.")
			continue
		}

		return result, nil
	}
}

// runMultilineInput runs gum write and returns the result
func runMultilineInput() (string, error) {
	return runGumCommand("write", "--placeholder", "Write something...")
}
