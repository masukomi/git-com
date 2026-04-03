package commit

import (
	"errors"
	"os/exec"
	"strings"
)

// HasStagedFiles checks if there are any staged files in the repository
func HasStagedFiles() (bool, error) {
	// git diff --cached --quiet exits 0 if no staged changes, 1 if there are staged changes.
	cmd := exec.Command("git", "diff", "--cached", "--quiet")
	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 {
				return true, nil
			}
		}
		return false, err
	}
	return false, nil
}

// CreateCommit creates a git commit with the given title and body
func CreateCommit(title, body string) error {
	message := buildCommitMessage(title, body)
	cmd := exec.Command("git", "commit", "-m", message)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(strings.TrimSpace(string(output)))
	}
	return nil
}

// AmendCommit amends the last commit with a new message
func AmendCommit(title, body string) error {
	message := buildCommitMessage(title, body)
	cmd := exec.Command("git", "commit", "--amend", "-m", message)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(strings.TrimSpace(string(output)))
	}
	return nil
}

// HasCommits checks if there are any commits in the repository
func HasCommits() (bool, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() != 0 {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetLastCommitBody returns the body of the last commit (everything after the 2nd line)
// Returns nil if there is no body or if the body is empty
func GetLastCommitBody() (*string, error) {
	cmd := exec.Command("git", "log", "-1", "--format=%B")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	message := string(output)
	lines := strings.SplitN(message, "\n", 3)
	if len(lines) < 3 {
		return nil, nil
	}

	body := strings.TrimSpace(lines[2])
	if body == "" {
		return nil, nil
	}

	return &body, nil
}

// buildCommitMessage constructs the full commit message from title and body
func buildCommitMessage(title, body string) string {
	title = strings.TrimSpace(title)
	body = strings.TrimSpace(body)

	if body == "" {
		return title
	}

	return title + "\n\n" + body
}
