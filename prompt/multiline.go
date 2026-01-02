package prompt

import (
	"strings"

	"git-com/config"
	"git-com/output"
	"git-com/tui"
)

// HandleMultilineText processes a multiline text input element
// If initialContent is not nil, the text area will be pre-filled with that content
func HandleMultilineText(elem config.Element, initialContent *string) (string, error) {
	for {
		// Get multiline text input
		result, err := tui.Write("Write something…", elem.Instructions, initialContent)
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

		// Clear initialContent after first use so it doesn't persist
		// to the next invocation of HandleMultilineText
		// Although, it'd be pretty silly to have 2 multiline-text
		// elements that targeted body. But hey. I'm not your mom.
		// You do you kid.
		initialContent = nil

		return result, nil
	}
}
