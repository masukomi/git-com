package prompt

import (
	"git-com/config"
)

// HandleConfirmation processes a confirmation element
// Returns empty string on affirmative, ErrUserAborted on negative or cancel
func HandleConfirmation(elem config.Element) (string, error) {
	// Display instructions if present
	DisplayInstructions(elem.Instructions)

	prompt := "Are you sure?"
	if elem.Instructions != "" {
		prompt = elem.Instructions
	}

	_, err := runGumCommand("confirm", prompt)
	if err != nil {
		// confirm returns exit code 1 for "No" selection
		return "", ErrUserAborted
	}

	return "", nil
}
