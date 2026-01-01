package main

import (
	"errors"
	"fmt"
	"os"

	"git-com/commit"
	"git-com/config"
	"git-com/output"
	"git-com/prompt"
	"git-com/tui"
)

func main() {
	os.Exit(run())
}

func run() int {
	// Load configuration from git root
	cfg, err := config.LoadConfig()
	if err != nil {
		if errors.Is(err, config.ErrConfigNotFound) {
			output.PrintError("Config file .git-com.yaml not found in git repository root")
		} else if errors.Is(err, config.ErrNotInGitRepo) {
			output.PrintError("Not in a git repository")
		} else {
			output.PrintError("Error loading config: " + err.Error())
		}
		return 1
	}

	// Validate configuration
	if !config.ValidateConfig(cfg) {
		return 1
	}

	// Check if there are staged files
	hasStaged, err := commit.HasStagedFiles()
	if err != nil {
		output.PrintError("Error checking staged files: " + err.Error())
		return 1
	}
	if !hasStaged {
		output.PrintWarningToStderr("You need to stage some files before we can commit.")
		return 64
	}

	// Process all elements
	result, err := prompt.ProcessElements(cfg)
	if err != nil {
		if errors.Is(err, prompt.ErrUserAborted) {
			// User pressed Ctrl+C, exit silently
			return 1
		}
		output.PrintError("Error processing input: " + err.Error())
		return 1
	}

	// Clear screen and show commit preview
	prompt.ClearScreen()
	fmt.Fprintln(os.Stderr, result.Title)
	if result.Body != "" {
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, result.Body)
	}
	fmt.Fprintln(os.Stderr)

	// Confirm with user
	confirmed, err := tui.Confirm("Is this good?")
	if err != nil {
		if errors.Is(err, tui.ErrAborted) {
			return 1
		}
		output.PrintError("Error during confirmation: " + err.Error())
		return 1
	}
	if !confirmed {
		return 0
	}

	// Create the commit
	if err := commit.CreateCommit(result.Title, result.Body); err != nil {
		output.PrintError("Error creating commit: " + err.Error())
		return 1
	}

	return 0
}
