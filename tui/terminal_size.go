package tui

import (
	"os"
	"strings"

	"golang.org/x/term"
)

// terminalWidth returns the current terminal width.
// Falls back to 80 if terminal size cannot be determined.
func terminalWidth() int {
	width, _, err := term.GetSize(int(os.Stderr.Fd()))
	if err != nil || width == 0 {
		return 80
	}
	return width
}
// terminalHeight returns the current terminal height.
// Falls back to 10 if terminal size cannot be determined.
func terminalHeight() int {
	_, height, err := term.GetSize(int(os.Stderr.Fd()))
	if err != nil || height == 0 {
		return 10
	}
	return height
}

// calculateHeight determines the available height for TUI components
// by querying terminal size and subtracting space needed for instructions and help.
// Returns terminal height - instruction lines - 2 (empty line + help line).
func calculateHeight(instructions string) int {
	termHeight := terminalHeight()

	instructionLines := 0
	if instructions != "" {
		instructionLines = strings.Count(instructions, "\n") + 1
	}

	height := termHeight - instructionLines - 2
	if height < 1 {
		height = 1
	}
	return height
}
