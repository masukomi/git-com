package commit

import (
	"errors"
	"os/exec"
	"strings"
	"time"

	git "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing/object"
)

// HasStagedFiles checks if there are any staged files in the repository
func HasStagedFiles() (bool, error) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return false, err
	}

	wt, err := repo.Worktree()
	if err != nil {
		return false, err
	}

	status, err := wt.Status()
	if err != nil {
		return false, err
	}

	for _, s := range status {
		// Check if file has staged changes (Added, Modified, Deleted, Renamed, Copied)
		if s.Staging != git.Unmodified && s.Staging != git.Untracked {
			return true, nil
		}
	}

	return false, nil
}

// CreateCommit creates a git commit with the given title and body
func CreateCommit(title, body string) error {
	// Open the repository
	repo, err := git.PlainOpen(".")
	if err != nil {
		return err
	}

	// Get the worktree
	wt, err := repo.Worktree()
	if err != nil {
		return err
	}

	// Get author info from git config (handles includes properly)
	author, err := getAuthorFromGitConfig()
	if err != nil {
		return err
	}

	// Build the commit message
	message := buildCommitMessage(title, body)

	// Create the commit
	_, err = wt.Commit(message, &git.CommitOptions{
		Author: author,
	})

	return err
}

// getAuthorFromGitConfig gets author info using git config command
// This properly handles [include] directives in .gitconfig
func getAuthorFromGitConfig() (*object.Signature, error) {
	name, err := runGitConfig("user.name")
	if err != nil {
		return nil, errors.New("author field is required: could not get user.name from git config")
	}

	email, err := runGitConfig("user.email")
	if err != nil {
		return nil, errors.New("author field is required: could not get user.email from git config")
	}

	return &object.Signature{
		Name:  name,
		Email: email,
		When:  time.Now(),
	}, nil
}

// runGitConfig runs git config to get a value
func runGitConfig(key string) (string, error) {
	cmd := exec.Command("git", "config", "--get", key)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
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

// HasCommits checks if there are any commits in the repository
func HasCommits() (bool, error) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return false, err
	}

	head, err := repo.Head()
	if err != nil {
		// If there's no HEAD, there are no commits
		return false, nil
	}

	return head != nil, nil
}

// GetLastCommitBody returns the body of the last commit (everything after the 2nd line)
// Returns nil if there is no body or if the body is empty
func GetLastCommitBody() (*string, error) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return nil, err
	}

	head, err := repo.Head()
	if err != nil {
		return nil, err
	}

	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return nil, err
	}

	message := commit.Message
	// Split message into lines
	lines := strings.SplitN(message, "\n", 3)
	if len(lines) < 3 {
		// No body
		return nil, nil
	}

	// The body is everything after the first line, trimmed of leading/trailing whitespace
	body := strings.TrimSpace(lines[2])
	if body == "" {
		return nil, nil
	}

	return &body, nil
}

// AmendCommit amends the last commit with a new message
func AmendCommit(title, body string) error {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return err
	}

	wt, err := repo.Worktree()
	if err != nil {
		return err
	}

	// Get author info from git config
	author, err := getAuthorFromGitConfig()
	if err != nil {
		return err
	}

	// Build the commit message
	message := buildCommitMessage(title, body)

	// Create the commit with amend option
	_, err = wt.Commit(message, &git.CommitOptions{
		Author: author,
		Amend:  true,
	})

	return err
}
