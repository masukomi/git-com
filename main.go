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
		verifyHasCommitsToAmend()
	}

	// Initialize oldCommitMessage as nil
	var oldCommitMessage *string

	// If amending, check for multiline-text elements with destination=body
	// and retrieve the last commit's body if such elements exist
	if !creatingNewCommit && hasMultilineTextBodyElement(cfg) {
		// body is already nil if empty, so just assign it
		oldCommitMessage = getOldCommitMessageBody()
	}

	// Check if there are staged files (only for new commits, not amends)
	if creatingNewCommit {
		verifyStagedFiles()

		if result := offerPendingCommit(); result != nil {
			performFinalConfirmation(result)
			commitOrAmend(creatingNewCommit, result)
			os.Exit(0)
		}
	}

	// Process all elements
	result, err := prompt.ProcessElements(cfg, oldCommitMessage)
	if err != nil {
		if errors.Is(err, prompt.ErrUserAborted) {
			// User pressed Ctrl+C, exit silently
			os.Exit(0)
		}
		output.PrintError("Error processing input: " + err.Error())
		os.Exit(1)
	}

	// Confirm with user (shows preview and allows editing if rejected)
	performFinalConfirmation(result)

	// Create or amend the commit based on the flag
	commitOrAmend(creatingNewCommit, result)

	os.Exit(0)
}

// attempts to get the body of the last commit
// prints an error and exits if there was a problem
func getOldCommitMessageBody() *string {
	body, err := commit.GetLastCommitBody()
	if err != nil {
		output.PrintError("Error getting last commit body: " + err.Error())
		os.Exit(1)
	}
	return body
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

// Create or amend the commit based on what the user indicated at launch.
func commitOrAmend(creatingNewCommit bool, result *prompt.Result) {
	// Save before attempting so the message survives a hook failure.
	if err := commit.SavePendingCommit(result.Title, result.Body); err != nil {
		output.PrintWarning("Could not save pending commit message: " + err.Error())
	}

	var commitErr error
	if creatingNewCommit {
		commitErr = commit.CreateCommit(result.Title, result.Body)
	} else {
		commitErr = commit.AmendCommit(result.Title, result.Body)
	}

	if commitErr != nil {
		os.Exit(1) // pending file remains for next run
	}

	_ = commit.DeletePendingCommit() // success — clean up
}

// offerPendingCommit checks for a saved commit message from a previous failed attempt.
// If found, shows the title and asks the user whether to reuse it.
// Returns a Result to jump straight to confirmation, or nil to proceed normally.
func offerPendingCommit() *prompt.Result {
	title, body, found, err := commit.LoadPendingCommit()
	if err != nil || !found {
		return nil
	}

	fmt.Fprintln(os.Stderr, "You have a pending commit message from a previous attempt:")
	fmt.Fprintf(os.Stderr, "  %s\n\n", title)

	useIt, err := tui.Confirm("Use this message?")
	if err != nil {
		if errors.Is(err, tui.ErrAborted) {
			os.Exit(1)
		}
		return nil
	}

	if !useIt {
		_ = commit.DeletePendingCommit()
		return nil
	}

	// Don't delete here — commitOrAmend will overwrite with the final
	// (possibly edited) version before committing, then delete on success.
	return &prompt.Result{Title: title, Body: body}
}

// tests if there are any commits.
// if it has problem determining this it will print an error
// if there are no commits it will print a warning
// if it doesn't find any commits it will exit
func verifyHasCommitsToAmend() {
	hasCommits, err := commit.HasCommits()
	if err != nil {
		output.PrintError("Error checking for commits: " + err.Error())
		os.Exit(1)
	}
	if !hasCommits {
		output.PrintWarning("There are no commits to amend.")
		os.Exit(1)
	}
}

// checks if the user has staged any files
// prints warning and exits if they haven't.
func verifyStagedFiles() {
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

// asks the user if they're ok with the commit message
// they've created. If they reject it, allows them to edit
// the title and body, then re-confirm.
// Exits if user aborts (Ctrl+C).
func performFinalConfirmation(result *prompt.Result) {
	for {
		// Show preview
		prompt.ClearScreen()
		fmt.Fprintln(os.Stderr, result.Title)
		if result.Body != "" {
			fmt.Fprintln(os.Stderr)
			fmt.Fprintln(os.Stderr, result.Body)
		}
		fmt.Fprintln(os.Stderr)

		// Confirm
		confirmed, err := tui.Confirm("Is this good?")
		if err != nil {
			if errors.Is(err, tui.ErrAborted) {
				os.Exit(1)
			}
			output.PrintError("Error during confirmation: " + err.Error())
			os.Exit(1)
		}

		if confirmed {
			return // proceed to commit
		}

		// Edit title
		prompt.ClearScreen()
		newTitle, err := tui.InputWithValue("", "Edit Title", &result.Title)
		if err != nil {
			if errors.Is(err, tui.ErrAborted) {
				os.Exit(1)
			}
			output.PrintError("Error editing title: " + err.Error())
			os.Exit(1)
		}
		result.Title = newTitle

		// Edit body
		newBody, err := tui.Write("", "Edit Description", &result.Body)
		if err != nil {
			if errors.Is(err, tui.ErrAborted) {
				os.Exit(1)
			}
			output.PrintError("Error editing description: " + err.Error())
			os.Exit(1)
		}
		result.Body = newBody

		// Loop back to show preview and confirm again
	}
}
