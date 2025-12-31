package main

import (
	"errors"
	"os"

	"git-com/commit"
	"git-com/config"
	"git-com/output"
	"git-com/prompt"
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

	// Clear screen before commit
	prompt.ClearScreen()

	// Create the commit
	if err := commit.CreateCommit(result.Title, result.Body); err != nil {
		output.PrintError("Error creating commit: " + err.Error())
		return 1
	}

	return 0
}
