package prompt

import (
	"errors"
	"fmt"
	"strings"

	"git-com/config"
	"git-com/output"
	"git-com/tui"

	"github.com/charmbracelet/lipgloss"
)

var (
	// ErrUserAborted indicates the user pressed Ctrl+C or Esc
	ErrUserAborted = errors.New("user aborted")

	// italicStyle is used for special options like "Other…" and empty selection text
	italicStyle = lipgloss.NewStyle().Italic(true)
)

// otherOption is the styled label for the "add new item" option in modifiable selects
var otherOption = italicStyle.Render("Other…")

// Italicize applies italic styling to text (used for special list options)
func Italicize(text string) string {
	return italicStyle.Render(text)
}

// ClearScreen clears the terminal screen
func ClearScreen() {
	fmt.Print("\033[H\033[2J")
}

// isAbortError checks if the error is an abort error from tui
func isAbortError(err error) bool {
	return errors.Is(err, tui.ErrAborted)
}

// handleOtherSelection handles when user selects "Other…" to add a new item
func handleOtherSelection(elementName string, cfg *config.Config) (string, error) {
	result, err := tui.Input("Enter new option…", "Add & select a new item")
	if err != nil {
		if isAbortError(err) {
			return "", ErrUserAborted
		}
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
