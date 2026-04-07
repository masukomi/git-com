---
name: git-com-structured-commits
description: 'Use when about to make a git commit or run git commit/git commit --amend. Checks whether git-com is installed and a .git-com.yaml config exists, then creates a structured commit using git-com instead of plain git commit. Trigger phrases: "commit", "commit changes", "make a commit", "git commit", "commit my work", "commit these changes".'
---

## When to Use This Skill

Before running `git commit` or `git commit --amend`, always check whether git-com
should be used instead. Use git-com when ALL of the following are true:

1. `which git-com` succeeds (git-com is installed)
2. A `.git-com.yaml` or `.git-com.yml` exists in the git repository root

If either check fails, fall back to normal `git commit`.

## Steps

### 1. Detect git-com

```bash
which git-com
```

If this fails, use `git commit` normally and stop.

```bash
ls "$(git rev-parse --show-toplevel)"/.git-com.y*ml >/dev/null 2>&1
```

If this returns a non-zero exit code, use `git commit` normally and stop.

### 2. Get the commit schema

```bash
git-com --dump-instructions
```

This outputs a YAML document under an `elements:` key. Each key under `elements:`
is a field you must fill in.

**Reading the output:**

| Field | Meaning |
|---|---|
| `type: text` | Free-text string value |
| `type: multiline-text` | Multi-line string value |
| `type: select` | Must be one of `options:` (unless `modifiable: true`, then any value is accepted) |
| `type: multi-select` | YAML list of values from `options:` (unless `modifiable: true`) |
| `type: confirmation` | `true` to proceed, `false` to abort the commit |
| `required: true` | Field must have a non-empty, non-null value |
| `required: false` | Field should be `null` if you have nothing to provide |
| `agent-hint:` | Guidance for the value you provide — follow it |
| `instructions:` | Describes what the field is asking for |

### 3. Create a temporary answers file

Create a file at `/tmp/git-com-answers-<timestamp>.yaml`.

The answers file is a flat YAML key-value document. It must contain one key for
every key listed under `elements:` in the dump-instructions output.

Rules:
- For `required: true` elements: provide a value that satisfies the type constraints
- For `required: false` elements: provide a value or use `null` if you have nothing to say
- For `type: select`: value must be from `options:` unless `modifiable: true`
- For `type: multi-select`: value is a YAML list (e.g. `[core, docs]`) or `[]` if none
- For `type: confirmation`: use `true` to proceed (almost always correct)
- For `type: text` with `data-type: integer`: provide an unquoted number or a quoted digit string

Example answers file:

```yaml
change-type: "fix"
breaking-indicator: null
commit-title: "corrected nil pointer in auth handler"
commit-description: "The handler did not guard against nil session before calling .User()"
code-sections:
  - core
  - auth
ticket-number: null
```

### 4. Run git-com with the answers file

```bash
git-com --answers /tmp/git-com-answers-<timestamp>.yaml
```

For amending the last commit instead:

```bash
git-com --amend --answers /tmp/git-com-answers-<timestamp>.yaml
```

### 5. Clean up

- If the commit succeeded: delete the temporary answers file
- If the commit failed (non-zero exit): leave the file for inspection, report the error

## Error Handling

- **Validation errors** (missing required field, invalid option): git-com prints each
  error to stderr. Fix the answers file and retry.
- **No staged files**: git-com exits with code 64. Stage files first, then retry.
- **git-com not found or no config**: fall back to `git commit -m "..."` as normal.
