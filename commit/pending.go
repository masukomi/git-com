package commit

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const pendingCommitFile = "GIT_COM_PENDING_COMMIT"

func pendingCommitPath() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return filepath.Join(strings.TrimSpace(string(out)), pendingCommitFile), nil
}

// SavePendingCommit writes the commit message to the gitdir so it survives
// a failed pre-commit hook and can be offered on the next invocation.
func SavePendingCommit(title, body string) error {
	path, err := pendingCommitPath()
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(buildCommitMessage(title, body)), 0644)
}

// LoadPendingCommit returns the saved title and body.
// found is false (with no error) when no pending commit file exists.
func LoadPendingCommit() (title, body string, found bool, err error) {
	path, err := pendingCommitPath()
	if err != nil {
		return "", "", false, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", "", false, nil
		}
		return "", "", false, err
	}
	lines := strings.SplitN(strings.TrimSpace(string(data)), "\n", 3)
	title = strings.TrimSpace(lines[0])
	if len(lines) == 3 {
		body = strings.TrimSpace(lines[2])
	}
	return title, body, true, nil
}

// DeletePendingCommit removes the pending commit file. A missing file is not an error.
func DeletePendingCommit() error {
	path, err := pendingCommitPath()
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
