package prompt

import (
	"errors"
	"fmt"
	"strings"

	"git-com/agent"
	"git-com/config"
	"git-com/output"
)

// the width we'll attempt to wrap
// multi-line values at.
const hardWrapWidth = 65

// ErrAgentAborted is returned when a confirmation element receives a negative answer.
var ErrAgentAborted = errors.New("confirmation element answered no")

// ProcessAnswers processes config elements using pre-supplied answers instead of the TUI.
// All validation errors are collected and reported together before returning.
func ProcessAnswers(cfg *config.Config, answers map[string]interface{}) (*Result, error) {
	result := &Result{}
	var errs []string

	for _, elem := range cfg.Elements {
		value, skip, err := processAnswer(elem, answers)
		if err != nil {
			if errors.Is(err, ErrAgentAborted) {
				return nil, ErrAgentAborted
			}
			errs = append(errs, err.Error())
			continue
		}

		if skip && !elem.IsIncludeEmpty() {
			continue
		}

		finalValue := applyDecorators(value, elem)

		switch elem.Destination {
		case config.DestTitle:
			result.Title += finalValue
		case config.DestBody:
			result.Body += finalValue
		}
	}

	if len(errs) > 0 {
		for _, e := range errs {
			output.PrintError(e)
		}
		return nil, fmt.Errorf("answers validation failed")
	}

	return result, nil
}

func processAnswer(elem config.Element, answers map[string]interface{}) (value string, skip bool, err error) {
	raw, present := answers[elem.Name]

	switch config.GetEffectiveType(elem) {
	case config.TypeText, config.TypeMultilineText:
		return processTextAnswer(elem, raw, present)
	case config.TypeSelect:
		return processSelectAnswer(elem, raw, present)
	case config.TypeMultiSelect:
		return processMultiSelectAnswer(elem, raw, present)
	case config.TypeConfirmation:
		return processConfirmationAnswer(raw, present)
	default:
		return processTextAnswer(elem, raw, present)
	}
}

func processTextAnswer(elem config.Element, raw interface{}, present bool) (string, bool, error) {
	if !present {
		if elem.IsAllowEmpty() {
			return "", true, nil
		}
		return "", false, fmt.Errorf("element %q is required but not present in answers", elem.Name)
	}

	value := strings.TrimSpace(agent.CoerceToString(raw))

	if value == "" {
		if elem.IsAllowEmpty() {
			return "", true, nil
		}
		return "", false, fmt.Errorf("element %q is required", elem.Name)
	}

	if elem.DataType != "" && !validateDataType(value, elem.DataType) {
		return "", false, fmt.Errorf("element %q: value %q is not a valid %s", elem.Name, value, elem.DataType)
	}

	if config.GetEffectiveType(elem) == config.TypeMultilineText {
		value = hardWrap(value, hardWrapWidth)
	}

	return value, false, nil
}


// hardWrap wraps text at the given column width, breaking only on spaces.
// Existing newlines are preserved. Words longer than width with no spaces
// are left unwrapped.
func hardWrap(text string, width int) string {
	lines := strings.Split(text, "\n")
	wrapped := make([]string, len(lines))
	for i, line := range lines {
		wrapped[i] = wrapLine(line, width)
	}
	return strings.Join(wrapped, "\n")
}

func wrapLine(line string, width int) string {
	if len(line) <= width {
		return line
	}
	words := strings.Split(line, " ")
	var out []string
	current := ""
	for _, word := range words {
		if current == "" {
			current = word
		} else if len(current)+1+len(word) <= width {
			current += " " + word
		} else {
			out = append(out, current)
			current = word
		}
	}
	if current != "" {
		out = append(out, current)
	}
	return strings.Join(out, "\n")
}

func processSelectAnswer(elem config.Element, raw interface{}, present bool) (string, bool, error) {
	if !present {
		if elem.IsAllowEmpty() {
			return "", true, nil
		}
		return "", false, fmt.Errorf("element %q is required but not present in answers", elem.Name)
	}

	value := strings.TrimSpace(agent.CoerceToString(raw))

	if value == "" {
		if elem.IsAllowEmpty() {
			return "", true, nil
		}
		return "", false, fmt.Errorf("element %q is required", elem.Name)
	}

	if !elem.IsModifiable() {
		valid := false
		for _, opt := range elem.Options {
			if opt == value {
				valid = true
				break
			}
		}
		if !valid {
			return "", false, fmt.Errorf("element %q: %q is not a valid option (valid: %v)", elem.Name, value, elem.Options)
		}
	}

	return value, false, nil
}

func processMultiSelectAnswer(elem config.Element, raw interface{}, present bool) (string, bool, error) {
	if !present {
		if elem.IsAllowEmpty() {
			return "", true, nil
		}
		return "", false, fmt.Errorf("element %q is required but not present in answers", elem.Name)
	}

	selections, err := agent.CoerceToStringSlice(raw)
	if err != nil {
		return "", false, fmt.Errorf("element %q: %w", elem.Name, err)
	}

	filtered := make([]string, 0, len(selections))
	for _, s := range selections {
		if s = strings.TrimSpace(s); s != "" {
			filtered = append(filtered, s)
		}
	}

	if len(filtered) == 0 {
		if elem.IsAllowEmpty() {
			return "", true, nil
		}
		return "", false, fmt.Errorf("element %q is required", elem.Name)
	}

	if !elem.IsModifiable() {
		for _, sel := range filtered {
			valid := false
			for _, opt := range elem.Options {
				if opt == sel {
					valid = true
					break
				}
			}
			if !valid {
				return "", false, fmt.Errorf("element %q: %q is not a valid option (valid: %v)", elem.Name, sel, elem.Options)
			}
		}
	}

	return formatMultiSelectResult(filtered, elem), false, nil
}

func processConfirmationAnswer(raw interface{}, present bool) (string, bool, error) {
	if !present || raw == nil {
		return "", true, nil // absent = default yes
	}

	switch v := raw.(type) {
	case bool:
		if !v {
			return "", false, ErrAgentAborted
		}
	case string:
		s := strings.ToLower(strings.TrimSpace(v))
		if s == "no" || s == "false" {
			return "", false, ErrAgentAborted
		}
	}

	return "", true, nil // yes/true → proceed
}
