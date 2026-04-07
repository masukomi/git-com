package prompt

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"git-com/config"
)

// boolPtr creates a *bool for use in test Element literals.
func boolPtr(b bool) *bool { return &b }

// makeConfig is a convenience helper for building a *config.Config from a slice of elements.
func makeConfig(elems ...config.Element) *config.Config {
	return &config.Config{Elements: elems}
}

// --- ProcessAnswers: text / multiline-text ---

func TestProcessAnswers_TextRequired(t *testing.T) {
	cfg := makeConfig(config.Element{
		Name:        "title",
		Type:        config.TypeText,
		Destination: config.DestTitle,
	})

	t.Run("valid answer appended to title", func(t *testing.T) {
		result, err := ProcessAnswers(cfg, map[string]interface{}{"title": "fixed the bug"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Title != "fixed the bug" {
			t.Errorf("Title = %q, want 'fixed the bug'", result.Title)
		}
	})

	t.Run("empty answer is an error", func(t *testing.T) {
		_, err := ProcessAnswers(cfg, map[string]interface{}{"title": ""})
		if err == nil {
			t.Fatal("expected error for empty required field, got nil")
		}
	})

	t.Run("missing key is an error", func(t *testing.T) {
		_, err := ProcessAnswers(cfg, map[string]interface{}{})
		if err == nil {
			t.Fatal("expected error for missing required field, got nil")
		}
	})
}

func TestProcessAnswers_TextAllowEmpty(t *testing.T) {
	cfg := makeConfig(config.Element{
		Name:        "body",
		Type:        config.TypeText,
		Destination: config.DestBody,
		AllowEmpty:  boolPtr(true),
	})

	t.Run("empty string skips element", func(t *testing.T) {
		result, err := ProcessAnswers(cfg, map[string]interface{}{"body": ""})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Body != "" {
			t.Errorf("Body = %q, want empty", result.Body)
		}
	})

	t.Run("missing key skips element", func(t *testing.T) {
		result, err := ProcessAnswers(cfg, map[string]interface{}{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Body != "" {
			t.Errorf("Body = %q, want empty", result.Body)
		}
	})

	t.Run("null value skips element (same as missing)", func(t *testing.T) {
		result, err := ProcessAnswers(cfg, map[string]interface{}{"body": nil})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Body != "" {
			t.Errorf("Body = %q, want empty", result.Body)
		}
	})
}

func TestProcessAnswers_TextDataTypeValidation(t *testing.T) {
	cfg := makeConfig(config.Element{
		Name:        "ticket",
		Type:        config.TypeText,
		Destination: config.DestBody,
		DataType:    config.DataTypeInteger,
	})

	t.Run("valid integer string passes", func(t *testing.T) {
		result, err := ProcessAnswers(cfg, map[string]interface{}{"ticket": "1234"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Body != "1234" {
			t.Errorf("Body = %q, want '1234'", result.Body)
		}
	})

	t.Run("unquoted integer YAML value coerced and passes", func(t *testing.T) {
		// YAML decodes ticket: 1234 (no quotes) as int, not string.
		result, err := ProcessAnswers(cfg, map[string]interface{}{"ticket": 1234})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Body != "1234" {
			t.Errorf("Body = %q, want '1234'", result.Body)
		}
	})

	t.Run("non-integer string fails", func(t *testing.T) {
		_, err := ProcessAnswers(cfg, map[string]interface{}{"ticket": "abc"})
		if err == nil {
			t.Fatal("expected error for non-integer value, got nil")
		}
	})
}

func TestProcessAnswers_MultilineText(t *testing.T) {
	cfg := makeConfig(config.Element{
		Name:        "description",
		Type:        config.TypeMultilineText,
		Destination: config.DestBody,
		AllowEmpty:  boolPtr(true),
	})

	t.Run("short text is not wrapped", func(t *testing.T) {
		result, err := ProcessAnswers(cfg, map[string]interface{}{"description": "line one\nline two"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Body != "line one\nline two" {
			t.Errorf("Body = %q, want 'line one\\nline two'", result.Body)
		}
	})

	t.Run(fmt.Sprintf("long text is wrapped at %d chars on spaces", hardWrapWidth), func(t *testing.T) {
		input := "The quick brown fox jumps over the lazy dog and then runs away into the forest"
		result, err := ProcessAnswers(cfg, map[string]interface{}{"description": input})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, line := range strings.Split(result.Body, "\n") {
			if len(line) > hardWrapWidth {
				t.Errorf("line exceeds %d chars (%d): %q", hardWrapWidth, len(line), line)
			}
		}
		// All original words must still be present
		if !strings.Contains(strings.ReplaceAll(result.Body, "\n", " "), "quick brown fox") {
			t.Errorf("wrapped output missing original words: %q", result.Body)
		}
	})

	t.Run(fmt.Sprintf("word with no spaces longer than %d chars is not broken", hardWrapWidth), func(t *testing.T) {
		longWord := "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrs"
		result, err := ProcessAnswers(cfg, map[string]interface{}{"description": longWord})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Body != longWord {
			t.Errorf("Body = %q, want %q", result.Body, longWord)
		}
	})

	t.Run("existing newlines preserved and each line wrapped independently", func(t *testing.T) {
		input := "First paragraph with enough words to exceed the sixty-five character wrapping limit here\nSecond paragraph also long enough to exceed sixty-five chars for wrapping"
		result, err := ProcessAnswers(cfg, map[string]interface{}{"description": input})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, line := range strings.Split(result.Body, "\n") {
			if len(line) > hardWrapWidth {
				t.Errorf("line exceeds %d chars (%d): %q", hardWrapWidth, len(line), line)
			}
		}
	})
}

func TestProcessAnswers_TextNotWrapped(t *testing.T) {
	cfg := makeConfig(config.Element{
		Name:        "title",
		Type:        config.TypeText,
		Destination: config.DestTitle,
	})

	long := "this is a text element whose value is longer than sixty-five characters total"
	result, err := ProcessAnswers(cfg, map[string]interface{}{"title": long})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Title != long {
		t.Errorf("text element should not be wrapped, got %q", result.Title)
	}
}

// --- hardWrap / wrapLine unit tests ---

func TestWrapLine(t *testing.T) {
	tests := []struct {
		name  string
		input string
		width int
		want  string
	}{
		{
			name:  "short line unchanged",
			input: "hello world",
			width: 65,
			want:  "hello world",
		},
		{
			name:  "exactly width unchanged",
			input: "hello",
			width: 5,
			want:  "hello",
		},
		{
			name:  "wraps on space boundary",
			input: "one two three",
			width: 7,
			want:  "one two\nthree",
		},
		{
			name:  "single word longer than width not broken",
			input: "superlongwordwithoutanyspaces",
			width: 10,
			want:  "superlongwordwithoutanyspaces",
		},
		{
			name:  "multiple wraps",
			input: "a b c d e f",
			width: 3,
			want:  "a b\nc d\ne f",
		},
		{
			name:  "empty string",
			input: "",
			width: 65,
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapLine(tt.input, tt.width)
			if got != tt.want {
				t.Errorf("wrapLine(%q, %d) = %q, want %q", tt.input, tt.width, got, tt.want)
			}
		})
	}
}

func TestHardWrap(t *testing.T) {
	tests := []struct {
		name  string
		input string
		width int
		want  string
	}{
		{
			name:  "preserves existing newlines",
			input: "short\nshort",
			width: 65,
			want:  "short\nshort",
		},
		{
			name:  "wraps each line independently",
			input: "one two three\nfour five six",
			width: 9,
			want:  "one two\nthree\nfour five\nsix",
		},
		{
			name:  "does not merge separate lines",
			input: "hello\nworld",
			width: 65,
			want:  "hello\nworld",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hardWrap(tt.input, tt.width)
			if got != tt.want {
				t.Errorf("hardWrap(%q, %d) = %q, want %q", tt.input, tt.width, got, tt.want)
			}
		})
	}
}

// --- ProcessAnswers: select ---

func TestProcessAnswers_Select(t *testing.T) {
	cfg := makeConfig(config.Element{
		Name:        "change-type",
		Type:        config.TypeSelect,
		Destination: config.DestTitle,
		Options:     []string{"fix", "add", "refactor"},
	})

	t.Run("valid option passes", func(t *testing.T) {
		result, err := ProcessAnswers(cfg, map[string]interface{}{"change-type": "fix"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Title != "fix" {
			t.Errorf("Title = %q, want 'fix'", result.Title)
		}
	})

	t.Run("invalid option is error", func(t *testing.T) {
		_, err := ProcessAnswers(cfg, map[string]interface{}{"change-type": "unknown"})
		if err == nil {
			t.Fatal("expected error for invalid option, got nil")
		}
	})

	t.Run("empty value on required element is error", func(t *testing.T) {
		_, err := ProcessAnswers(cfg, map[string]interface{}{"change-type": ""})
		if err == nil {
			t.Fatal("expected error for empty required select, got nil")
		}
	})
}

func TestProcessAnswers_SelectAllowEmpty(t *testing.T) {
	cfg := makeConfig(config.Element{
		Name:        "indicator",
		Type:        config.TypeSelect,
		Destination: config.DestTitle,
		Options:     []string{"breaking"},
		AllowEmpty:  boolPtr(true),
	})

	t.Run("empty string skips element", func(t *testing.T) {
		result, err := ProcessAnswers(cfg, map[string]interface{}{"indicator": ""})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Title != "" {
			t.Errorf("Title = %q, want empty", result.Title)
		}
	})

	t.Run("null value skips element (same as missing)", func(t *testing.T) {
		result, err := ProcessAnswers(cfg, map[string]interface{}{"indicator": nil})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Title != "" {
			t.Errorf("Title = %q, want empty", result.Title)
		}
	})
}

func TestProcessAnswers_SelectModifiable(t *testing.T) {
	cfg := makeConfig(config.Element{
		Name:        "change-type",
		Type:        config.TypeSelect,
		Destination: config.DestTitle,
		Options:     []string{"fix", "add"},
		Modifiable:  boolPtr(true),
	})

	t.Run("value not in options is accepted when modifiable", func(t *testing.T) {
		result, err := ProcessAnswers(cfg, map[string]interface{}{"change-type": "new-custom-type"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Title != "new-custom-type" {
			t.Errorf("Title = %q, want 'new-custom-type'", result.Title)
		}
	})
}

// --- ProcessAnswers: multi-select ---

func TestProcessAnswers_MultiSelect_JoinedString(t *testing.T) {
	cfg := makeConfig(config.Element{
		Name:        "sections",
		Type:        config.TypeMultiSelect,
		Destination: config.DestBody,
		RecordAs:    config.RecordAsJoinedString,
		AllowEmpty:  boolPtr(true),
		Options:     []string{"core", "docs", "tests"},
	})

	t.Run("list formatted as joined string", func(t *testing.T) {
		result, err := ProcessAnswers(cfg, map[string]interface{}{
			"sections": []interface{}{"core", "tests"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Body != "core, tests" {
			t.Errorf("Body = %q, want 'core, tests'", result.Body)
		}
	})

	t.Run("empty list skips element", func(t *testing.T) {
		result, err := ProcessAnswers(cfg, map[string]interface{}{"sections": []interface{}{}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Body != "" {
			t.Errorf("Body = %q, want empty", result.Body)
		}
	})

	t.Run("null value skips element (same as missing)", func(t *testing.T) {
		result, err := ProcessAnswers(cfg, map[string]interface{}{"sections": nil})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Body != "" {
			t.Errorf("Body = %q, want empty", result.Body)
		}
	})
}

func TestProcessAnswers_MultiSelect_List(t *testing.T) {
	cfg := makeConfig(config.Element{
		Name:        "sections",
		Type:        config.TypeMultiSelect,
		Destination: config.DestBody,
		RecordAs:    config.RecordAsList,
		Options:     []string{"core", "docs", "tests"},
	})

	result, err := ProcessAnswers(cfg, map[string]interface{}{
		"sections": []interface{}{"core", "docs"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "\n- core\n- docs\n"
	if result.Body != expected {
		t.Errorf("Body = %q, want %q", result.Body, expected)
	}
}

func TestProcessAnswers_MultiSelect_InvalidOption(t *testing.T) {
	cfg := makeConfig(config.Element{
		Name:        "sections",
		Type:        config.TypeMultiSelect,
		Destination: config.DestBody,
		RecordAs:    config.RecordAsJoinedString,
		Options:     []string{"core", "docs"},
	})

	_, err := ProcessAnswers(cfg, map[string]interface{}{
		"sections": []interface{}{"core", "unknown"},
	})
	if err == nil {
		t.Fatal("expected error for invalid multi-select option, got nil")
	}
}

func TestProcessAnswers_MultiSelect_Modifiable(t *testing.T) {
	cfg := makeConfig(config.Element{
		Name:        "sections",
		Type:        config.TypeMultiSelect,
		Destination: config.DestBody,
		RecordAs:    config.RecordAsJoinedString,
		Options:     []string{"core", "docs"},
		Modifiable:  boolPtr(true),
	})

	result, err := ProcessAnswers(cfg, map[string]interface{}{
		"sections": []interface{}{"core", "new-area"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Body != "core, new-area" {
		t.Errorf("Body = %q, want 'core, new-area'", result.Body)
	}
}

func TestProcessAnswers_MultiSelect_RequiredMissing(t *testing.T) {
	cfg := makeConfig(config.Element{
		Name:        "sections",
		Type:        config.TypeMultiSelect,
		Destination: config.DestBody,
		RecordAs:    config.RecordAsJoinedString,
		Options:     []string{"core"},
	})

	_, err := ProcessAnswers(cfg, map[string]interface{}{})
	if err == nil {
		t.Fatal("expected error for missing required multi-select, got nil")
	}
}

// --- ProcessAnswers: confirmation ---

func TestProcessAnswers_Confirmation(t *testing.T) {
	cfg := makeConfig(config.Element{
		Name: "proceed",
		Type: config.TypeConfirmation,
	})

	tests := []struct {
		name        string
		answer      interface{}
		present     bool
		wantAborted bool
	}{
		{"absent defaults to yes", nil, false, false},
		{"bool true proceeds", true, true, false},
		{"bool false aborts", false, true, true},
		{"string yes proceeds", "yes", true, false},
		{"string no aborts", "no", true, true},
		{"string Yes proceeds", "Yes", true, false},
		{"string NO aborts", "NO", true, true},
		{"string false aborts", "false", true, true},
		{"string true proceeds", "true", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			answers := map[string]interface{}{}
			if tt.present {
				answers["proceed"] = tt.answer
			}
			_, err := ProcessAnswers(cfg, answers)
			if tt.wantAborted {
				if !errors.Is(err, ErrAgentAborted) {
					t.Errorf("expected ErrAgentAborted, got %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// --- ProcessAnswers: decorators ---

func TestProcessAnswers_DecoratorsApplied(t *testing.T) {
	cfg := makeConfig(config.Element{
		Name:         "change-type",
		Type:         config.TypeSelect,
		Destination:  config.DestTitle,
		Options:      []string{"fix", "add"},
		BeforeString: "[",
		AfterString:  "]",
	})

	result, err := ProcessAnswers(cfg, map[string]interface{}{"change-type": "fix"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Title != "[fix]" {
		t.Errorf("Title = %q, want '[fix]'", result.Title)
	}
}

func TestProcessAnswers_FormatApplied(t *testing.T) {
	format := "%-12s"
	cfg := makeConfig(config.Element{
		Name:         "change-type",
		Type:         config.TypeSelect,
		Destination:  config.DestTitle,
		Options:      []string{"fix"},
		BeforeString: "[",
		AfterString:  "]",
		Format:       &format,
	})

	result, err := ProcessAnswers(cfg, map[string]interface{}{"change-type": "fix"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Title != "[fix]       " {
		t.Errorf("Title = %q, want '[fix]       '", result.Title)
	}
}

// --- ProcessAnswers: include-empty ---

func TestProcessAnswers_IncludeEmpty(t *testing.T) {
	cfg := makeConfig(config.Element{
		Name:         "spacer",
		Type:         config.TypeText,
		Destination:  config.DestBody,
		AllowEmpty:   boolPtr(true),
		IncludeEmpty: boolPtr(true),
		BeforeString: "\n",
	})

	result, err := ProcessAnswers(cfg, map[string]interface{}{"spacer": ""})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Even though value is empty, decorators should be applied due to include-empty.
	if result.Body != "\n" {
		t.Errorf("Body = %q, want '\\n'", result.Body)
	}
}

// --- ProcessAnswers: multiple errors collected ---

func TestProcessAnswers_AllErrorsCollected(t *testing.T) {
	cfg := makeConfig(
		config.Element{Name: "a", Type: config.TypeText, Destination: config.DestTitle},
		config.Element{Name: "b", Type: config.TypeText, Destination: config.DestTitle},
		config.Element{Name: "c", Type: config.TypeText, Destination: config.DestTitle},
	)

	// All three are required but absent — should return one aggregated error, not just the first.
	_, err := ProcessAnswers(cfg, map[string]interface{}{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// Error should not be ErrAgentAborted — it's a validation failure.
	if errors.Is(err, ErrAgentAborted) {
		t.Errorf("expected validation error, not ErrAgentAborted")
	}
}

// --- ProcessAnswers: title and body assembled correctly ---

func TestProcessAnswers_TitleAndBodyAssembly(t *testing.T) {
	format := "%-12s"
	cfg := makeConfig(
		config.Element{
			Name:         "change-type",
			Type:         config.TypeSelect,
			Destination:  config.DestTitle,
			Options:      []string{"fix", "add"},
			BeforeString: "[",
			AfterString:  "]",
			Format:       &format,
		},
		config.Element{
			Name:        "commit-title",
			Type:        config.TypeText,
			Destination: config.DestTitle,
		},
		config.Element{
			Name:        "description",
			Type:        config.TypeMultilineText,
			Destination: config.DestBody,
			AllowEmpty:  boolPtr(true),
		},
	)

	result, err := ProcessAnswers(cfg, map[string]interface{}{
		"change-type":  "fix",
		"commit-title": "corrected nil pointer dereference",
		"description":  "The handler did not check for nil before calling .String()",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantTitle := "[fix]       corrected nil pointer dereference"
	if result.Title != wantTitle {
		t.Errorf("Title = %q, want %q", result.Title, wantTitle)
	}
	wantBody := "The handler did not check for nil before calling .String()"
	if result.Body != wantBody {
		t.Errorf("Body = %q, want %q", result.Body, wantBody)
	}
}
