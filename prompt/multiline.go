package prompt

import (
	"strings"

	"git-com/config"
	"git-com/output"
	"git-com/tui"
)

// HandleMultilineText processes a multiline text input element
func HandleMultilineText(elem config.Element) (string, error) {
	for {
		// Get multiline text input
		result, err := tui.Write("Write something…", elem.Instructions)
		if err != nil {
			if isAbortError(err) {
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
