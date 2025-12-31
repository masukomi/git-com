package prompt

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"git-com/output"
)

var (
	// ErrUserAborted indicates the user pressed Ctrl+C
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

// DisplayHint prints hint text in gray
func DisplayHint(text string) {
	if text != "" {
		output.PrintHint(text)
	}
}

// runGumCommand runs a gum command and returns the output
func runGumCommand(args ...string) (string, error) {
	cmd := exec.Command("gum", args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	err := cmd.Run()
	if err != nil {
		// Check if it's an exit error (user cancelled or made no selection)
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 || exitErr.ExitCode() == 130 {
				return "", ErrUserAborted
			}
		}
		return "", err
	}

	return strings.TrimSuffix(stdout.String(), "\n"), nil
}

// CheckGumInstalled checks if gum is available in PATH
func CheckGumInstalled() error {
	_, err := exec.LookPath("gum")
	if err != nil {
		return errors.New("gum is not installed. Please install it: https://github.com/charmbracelet/gum#installation")
	}
	return nil
}
