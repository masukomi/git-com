package prompt

import (
	"errors"
	"fmt"

	"git-com/output"
	"git-com/tui"
)

var (
	// ErrUserAborted indicates the user pressed Ctrl+C or Esc
	ErrUserAborted = errors.New("user aborted")
)

// ClearScreen clears the terminal screen
func ClearScreen() {
	fmt.Print("\033[H\033[2J")
}

// DisplayInstructions prints instructions text without coloring
func DisplayInstructions(text string) {
	if text != "" {
		output.Print(text)
	}
}

// isAbortError checks if the error is an abort error from tui
func isAbortError(err error) bool {
	return errors.Is(err, tui.ErrAborted)
}
