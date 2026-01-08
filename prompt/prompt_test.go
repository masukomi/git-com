package prompt

import (
	"testing"

	"git-com/config"
)

// Helper to create a string pointer
func strPtr(s string) *string {
	return &s
}

func TestApplyDecorators(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		elem     config.Element
		expected string
	}{
		{
			name:     "no decorators",
			value:    "foo",
			elem:     config.Element{},
			expected: "foo",
		},
		{
			name:  "before-string only",
			value: "foo",
			elem: config.Element{
				BeforeString: "[",
			},
			expected: "[foo",
		},
		{
			name:  "after-string only",
			value: "foo",
			elem: config.Element{
				AfterString: "]",
			},
			expected: "foo]",
		},
		{
			name:  "before and after strings",
			value: "foo",
			elem: config.Element{
				BeforeString: "[",
				AfterString:  "]",
			},
			expected: "[foo]",
		},
		{
			name:  "format only - simple",
			value: "foo",
			elem: config.Element{
				Format: strPtr("%s"),
			},
			expected: "foo",
		},
		{
			name:  "format only - left padded",
			value: "foo",
			elem: config.Element{
				Format: strPtr("%-10s"),
			},
			expected: "foo       ",
		},
		{
			name:  "format only - right padded",
			value: "foo",
			elem: config.Element{
				Format: strPtr("%10s"),
			},
			expected: "       foo",
		},
		{
			name:  "before/after strings with format",
			value: "foo",
			elem: config.Element{
				BeforeString: "[",
				AfterString:  "]",
				Format:       strPtr("%-10s"),
			},
			expected: "[foo]     ",
		},
		{
			name:  "format with surrounding text",
			value: "foo",
			elem: config.Element{
				Format: strPtr("prefix %s suffix"),
			},
			expected: "prefix foo suffix",
		},
		{
			name:  "full example from issue",
			value: "ji-woo",
			elem: config.Element{
				BeforeString: "[",
				AfterString:  "]",
				Format:       strPtr("%-12s"),
			},
			expected: "[ji-woo]    ",
		},
		{
			name:  "empty value with format",
			value: "",
			elem: config.Element{
				BeforeString: "[",
				AfterString:  "]",
				Format:       strPtr("%-10s"),
			},
			expected: "[]        ",
		},
		{
			name:  "nil format is not applied",
			value: "foo",
			elem: config.Element{
				BeforeString: "[",
				AfterString:  "]",
				Format:       nil,
			},
			expected: "[foo]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := applyDecorators(tt.value, tt.elem)
			if result != tt.expected {
				t.Errorf("applyDecorators() = %q, want %q", result, tt.expected)
			}
		})
	}
}
