package prompt

import (
	"git-com/config"
	"git-com/tui"
)

// HandleConfirmation processes a confirmation element
// Returns empty string on affirmative, ErrUserAborted on negative or cancel
func HandleConfirmation(elem config.Element) (string, error) {
	prompt := "Are you sure?"
	if elem.Instructions != "" {
		prompt = elem.Instructions
	}

	confirmed, err := tui.Confirm(prompt)
	if err != nil {
		if isAbortError(err) {
			return "", ErrUserAborted
		}
		return "", err
	}

	if !confirmed {
		return "", ErrUserAborted
	}

	return "", nil
}
