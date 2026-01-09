package prompt

import (
	"testing"

	"git-com/config"
)

// Helper to create a string pointer
func strPtrMS(s string) *string {
	return &s
}

func TestFormatAsList(t *testing.T) {
	tests := []struct {
		name       string
		selections []string
		elem       config.Element
		expected   string
	}{
		{
			name:       "no format - default bullet",
			selections: []string{"foo", "bar"},
			elem:       config.Element{},
			expected:   "\n- foo\n- bar\n",
		},
		{
			name:       "no format - custom bullet",
			selections: []string{"foo", "bar"},
			elem:       config.Element{BulletString: "* "},
			expected:   "\n* foo\n* bar\n",
		},
		{
			name:       "with format - left padded",
			selections: []string{"foo", "bar"},
			elem:       config.Element{Format: strPtrMS("%-10s")},
			expected:   "\n- foo     \n- bar     \n",
		},
		{
			name:       "with format - right padded",
			selections: []string{"foo", "bar"},
			elem:       config.Element{Format: strPtrMS("%10s")},
			expected:   "\n     - foo\n     - bar\n",
		},
		{
			name:       "with format and custom bullet",
			selections: []string{"a", "bb", "ccc"},
			elem:       config.Element{BulletString: ">> ", Format: strPtrMS("%-10s")},
			expected:   "\n>> a      \n>> bb     \n>> ccc    \n",
		},
		{
			name:       "single selection with format",
			selections: []string{"item"},
			elem:       config.Element{Format: strPtrMS("%-15s")},
			expected:   "\n- item         \n",
		},
		{
			name:       "empty format pointer is not applied",
			selections: []string{"foo", "bar"},
			elem:       config.Element{Format: strPtrMS("")},
			expected:   "\n- foo\n- bar\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatAsList(tt.selections, tt.elem)
			if result != tt.expected {
				t.Errorf("formatAsList() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestFormatMultiSelectResult(t *testing.T) {
	tests := []struct {
		name       string
		selections []string
		elem       config.Element
		expected   string
	}{
		{
			name:       "list with format applies per-line",
			selections: []string{"foo", "bar"},
			elem: config.Element{
				RecordAs: config.RecordAsList,
				Format:   strPtrMS("%-10s"),
			},
			expected: "\n- foo     \n- bar     \n",
		},
		{
			name:       "joined-string does not apply format",
			selections: []string{"foo", "bar"},
			elem: config.Element{
				RecordAs: config.RecordAsJoinedString,
				Format:   strPtrMS("%-20s"),
			},
			expected: "foo, bar",
		},
		{
			name:       "empty selections",
			selections: []string{},
			elem: config.Element{
				RecordAs: config.RecordAsList,
				Format:   strPtrMS("%-10s"),
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatMultiSelectResult(tt.selections, tt.elem)
			if result != tt.expected {
				t.Errorf("formatMultiSelectResult() = %q, want %q", result, tt.expected)
			}
		})
	}
}
