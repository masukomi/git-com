package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"git-com/commit"
	"git-com/config"
	"git-com/output"
	"git-com/prompt"
	"git-com/tui"
)

func main() {
	// Parse command-line flags
	amendFlag := flag.Bool("amend", false, "Amend the last commit")
	flag.Parse()

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
		os.Exit(1)
	}

	// Validate configuration
	if !config.ValidateConfig(cfg) {
		os.Exit(1)
	}

	// Determine if we are creating a new commit or amending
	creatingNewCommit := !*amendFlag

	// If amending, check that there are commits to amend
	if !creatingNewCommit {
		hasCommits, err := commit.HasCommits()
		if err != nil {
			output.PrintError("Error checking for commits: " + err.Error())
			os.Exit(1)
		}
		if !hasCommits {
			output.PrintError("There are no commits to amend.")
			os.Exit(1)
		}
	}

	// Initialize oldCommitMessage as nil
	var oldCommitMessage *string

	// If amending, check for multiline-text elements with destination=body
	// and retrieve the last commit's body if such elements exist
	if !creatingNewCommit {
		if hasMultilineTextBodyElement(cfg) {
			body, err := commit.GetLastCommitBody()
			if err != nil {
				output.PrintError("Error getting last commit body: " + err.Error())
				os.Exit(1)
			}
			// body is already nil if empty, so just assign it
			oldCommitMessage = body
		}
	}

	// Check if there are staged files (only for new commits, not amends)
	if creatingNewCommit {
		hasStaged, err := commit.HasStagedFiles()
		if err != nil {
			output.PrintError("Error checking staged files: " + err.Error())
			os.Exit(1)
		}
		if !hasStaged {
			output.PrintWarningToStderr("You need to stage some files before we can commit.")
			os.Exit(64)
		}
	}

	// Process all elements
	result, err := prompt.ProcessElements(cfg, oldCommitMessage)
	if err != nil {
		if errors.Is(err, prompt.ErrUserAborted) {
			// User pressed Ctrl+C, exit silently
			os.Exit(1)
		}
		output.PrintError("Error processing input: " + err.Error())
		os.Exit(1)
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
			os.Exit(1)
		}
		output.PrintError("Error during confirmation: " + err.Error())
		os.Exit(1)
	}
	if !confirmed {
		os.Exit(0)
	}

	// Create or amend the commit based on the flag
	if creatingNewCommit {
		if err := commit.CreateCommit(result.Title, result.Body); err != nil {
			output.PrintError("Error creating commit: " + err.Error())
			os.Exit(1)
		}
	} else {
		if err := commit.AmendCommit(result.Title, result.Body); err != nil {
			output.PrintError("Error amending commit: " + err.Error())
			os.Exit(1)
		}
	}

	os.Exit(0)
}

// hasMultilineTextBodyElement checks if the config has any multiline-text
// elements with destination=body
func hasMultilineTextBodyElement(cfg *config.Config) bool {
	for _, elem := range cfg.Elements {
		elemType := config.GetEffectiveType(elem)
		if elemType == config.TypeMultilineText && elem.Destination == config.DestBody {
			return true
		}
	}
	return false
}
