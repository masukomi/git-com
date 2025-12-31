package config

import (
	"testing"
)

func TestValidateElement(t *testing.T) {
	tests := []struct {
		name    string
		elem    Element
		wantErr bool
	}{
		// Destination validation
		{
			name:    "valid destination title",
			elem:    Element{Destination: DestTitle, Type: TypeText},
			wantErr: false,
		},
		{
			name:    "valid destination body",
			elem:    Element{Destination: DestBody, Type: TypeText},
			wantErr: false,
		},
		{
			name:    "invalid destination",
			elem:    Element{Destination: "invalid", Type: TypeText},
			wantErr: true,
		},
		{
			name:    "empty destination",
			elem:    Element{Destination: "", Type: TypeText},
			wantErr: true,
		},

		// Title newline validation
		{
			name:    "title with before-string containing newline",
			elem:    Element{Destination: DestTitle, Type: TypeText, BeforeString: "prefix\n"},
			wantErr: true,
		},
		{
			name:    "title with after-string containing newline",
			elem:    Element{Destination: DestTitle, Type: TypeText, AfterString: "\nsuffix"},
			wantErr: true,
		},
		{
			name:    "title with newline in middle of before-string",
			elem:    Element{Destination: DestTitle, Type: TypeText, BeforeString: "pre\nfix"},
			wantErr: true,
		},
		{
			name:    "title with before-string no newline",
			elem:    Element{Destination: DestTitle, Type: TypeText, BeforeString: "["},
			wantErr: false,
		},
		{
			name:    "title with after-string no newline",
			elem:    Element{Destination: DestTitle, Type: TypeText, AfterString: "] "},
			wantErr: false,
		},
		{
			name:    "body with before-string containing newline is allowed",
			elem:    Element{Destination: DestBody, Type: TypeText, BeforeString: "\n\nSection: "},
			wantErr: false,
		},
		{
			name:    "body with after-string containing newline is allowed",
			elem:    Element{Destination: DestBody, Type: TypeText, AfterString: "\n"},
			wantErr: false,
		},

		// Type validation
		{
			name:    "missing type without data-type",
			elem:    Element{Destination: DestTitle},
			wantErr: true,
		},
		{
			name:    "unknown type",
			elem:    Element{Destination: DestTitle, Type: "unknown"},
			wantErr: true,
		},

		// Text type
		{
			name:    "valid text element",
			elem:    Element{Destination: DestTitle, Type: TypeText},
			wantErr: false,
		},
		{
			name:    "text with valid data-type string",
			elem:    Element{Destination: DestTitle, Type: TypeText, DataType: DataTypeString},
			wantErr: false,
		},
		{
			name:    "text with valid data-type integer",
			elem:    Element{Destination: DestTitle, Type: TypeText, DataType: DataTypeInteger},
			wantErr: false,
		},
		{
			name:    "text with valid data-type float",
			elem:    Element{Destination: DestTitle, Type: TypeText, DataType: DataTypeFloat},
			wantErr: false,
		},
		{
			name:    "text with invalid data-type",
			elem:    Element{Destination: DestTitle, Type: TypeText, DataType: "invalid"},
			wantErr: true,
		},

		// Multiline-text type
		{
			name:    "valid multiline-text element",
			elem:    Element{Destination: DestBody, Type: TypeMultilineText},
			wantErr: false,
		},

		// Confirmation type
		{
			name:    "valid confirmation element",
			elem:    Element{Destination: DestTitle, Type: TypeConfirmation},
			wantErr: false,
		},

		// Select type
		{
			name:    "valid select element",
			elem:    Element{Destination: DestTitle, Type: TypeSelect, Options: []string{"a", "b"}},
			wantErr: false,
		},
		{
			name:    "select without options",
			elem:    Element{Destination: DestTitle, Type: TypeSelect},
			wantErr: true,
		},
		{
			name:    "select with empty options",
			elem:    Element{Destination: DestTitle, Type: TypeSelect, Options: []string{}},
			wantErr: true,
		},

		// Multi-select type
		{
			name: "valid multi-select with list",
			elem: Element{
				Destination: DestBody,
				Type:        TypeMultiSelect,
				Options:     []string{"a", "b"},
				RecordAs:    RecordAsList,
			},
			wantErr: false,
		},
		{
			name: "valid multi-select with joined-string",
			elem: Element{
				Destination: DestBody,
				Type:        TypeMultiSelect,
				Options:     []string{"a", "b"},
				RecordAs:    RecordAsJoinedString,
			},
			wantErr: false,
		},
		{
			name: "multi-select without options",
			elem: Element{
				Destination: DestBody,
				Type:        TypeMultiSelect,
				RecordAs:    RecordAsList,
			},
			wantErr: true,
		},
		{
			name: "multi-select without record-as",
			elem: Element{
				Destination: DestBody,
				Type:        TypeMultiSelect,
				Options:     []string{"a", "b"},
			},
			wantErr: true,
		},
		{
			name: "multi-select with invalid record-as",
			elem: Element{
				Destination: DestBody,
				Type:        TypeMultiSelect,
				Options:     []string{"a", "b"},
				RecordAs:    "invalid",
			},
			wantErr: true,
		},
		{
			name: "multi-select with empty-selection-text but allow-empty false",
			elem: Element{
				Destination:        DestBody,
				Type:               TypeMultiSelect,
				Options:            []string{"a", "b"},
				RecordAs:           RecordAsList,
				EmptySelectionText: "Skip",
				AllowEmpty:         boolPtr(false),
			},
			wantErr: true,
		},
		{
			name: "multi-select with empty-selection-text but allow-empty not set",
			elem: Element{
				Destination:        DestBody,
				Type:               TypeMultiSelect,
				Options:            []string{"a", "b"},
				RecordAs:           RecordAsList,
				EmptySelectionText: "Skip",
			},
			wantErr: true,
		},
		{
			name: "multi-select with empty-selection-text and allow-empty true",
			elem: Element{
				Destination:        DestBody,
				Type:               TypeMultiSelect,
				Options:            []string{"a", "b"},
				RecordAs:           RecordAsList,
				EmptySelectionText: "Skip",
				AllowEmpty:         boolPtr(true),
			},
			wantErr: false,
		},
		{
			name: "multi-select title destination with record-as list is invalid",
			elem: Element{
				Destination: DestTitle,
				Type:        TypeMultiSelect,
				Options:     []string{"a", "b"},
				RecordAs:    RecordAsList,
			},
			wantErr: true,
		},
		{
			name: "multi-select title destination with record-as joined-string is valid",
			elem: Element{
				Destination: DestTitle,
				Type:        TypeMultiSelect,
				Options:     []string{"a", "b"},
				RecordAs:    RecordAsJoinedString,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateElement(tt.elem)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateElement() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetEffectiveType(t *testing.T) {
	tests := []struct {
		name     string
		elem     Element
		expected ElementType
	}{
		{
			name:     "explicit type",
			elem:     Element{Type: TypeSelect},
			expected: TypeSelect,
		},
		{
			name:     "infer text from data-type",
			elem:     Element{DataType: DataTypeInteger},
			expected: TypeText,
		},
		{
			name:     "explicit type takes precedence over data-type",
			elem:     Element{Type: TypeMultilineText, DataType: DataTypeString},
			expected: TypeMultilineText,
		},
		{
			name:     "no type or data-type",
			elem:     Element{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetEffectiveType(tt.elem); got != tt.expected {
				t.Errorf("GetEffectiveType() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestValidateTextElement(t *testing.T) {
	tests := []struct {
		name    string
		elem    Element
		wantErr bool
	}{
		{
			name:    "no data-type",
			elem:    Element{},
			wantErr: false,
		},
		{
			name:    "data-type string",
			elem:    Element{DataType: DataTypeString},
			wantErr: false,
		},
		{
			name:    "data-type integer",
			elem:    Element{DataType: DataTypeInteger},
			wantErr: false,
		},
		{
			name:    "data-type float",
			elem:    Element{DataType: DataTypeFloat},
			wantErr: false,
		},
		{
			name:    "invalid data-type",
			elem:    Element{DataType: "boolean"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTextElement(tt.elem)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateTextElement() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateSelectElement(t *testing.T) {
	tests := []struct {
		name    string
		elem    Element
		wantErr bool
	}{
		{
			name:    "with options",
			elem:    Element{Options: []string{"a", "b", "c"}},
			wantErr: false,
		},
		{
			name:    "single option",
			elem:    Element{Options: []string{"only"}},
			wantErr: false,
		},
		{
			name:    "no options",
			elem:    Element{},
			wantErr: true,
		},
		{
			name:    "empty options slice",
			elem:    Element{Options: []string{}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSelectElement(tt.elem)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSelectElement() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateMultiSelectElement(t *testing.T) {
	tests := []struct {
		name    string
		elem    Element
		wantErr bool
	}{
		{
			name: "valid with list",
			elem: Element{
				Options:  []string{"a", "b"},
				RecordAs: RecordAsList,
			},
			wantErr: false,
		},
		{
			name: "valid with joined-string",
			elem: Element{
				Options:  []string{"a", "b"},
				RecordAs: RecordAsJoinedString,
			},
			wantErr: false,
		},
		{
			name: "missing options",
			elem: Element{
				RecordAs: RecordAsList,
			},
			wantErr: true,
		},
		{
			name: "missing record-as",
			elem: Element{
				Options: []string{"a", "b"},
			},
			wantErr: true,
		},
		{
			name: "invalid record-as",
			elem: Element{
				Options:  []string{"a", "b"},
				RecordAs: "csv",
			},
			wantErr: true,
		},
		{
			name: "empty-selection-text without allow-empty",
			elem: Element{
				Options:            []string{"a", "b"},
				RecordAs:           RecordAsList,
				EmptySelectionText: "Skip",
			},
			wantErr: true,
		},
		{
			name: "empty-selection-text with allow-empty true",
			elem: Element{
				Options:            []string{"a", "b"},
				RecordAs:           RecordAsList,
				EmptySelectionText: "Skip",
				AllowEmpty:         boolPtr(true),
			},
			wantErr: false,
		},
		{
			name: "allow-empty without empty-selection-text is valid",
			elem: Element{
				Options:    []string{"a", "b"},
				RecordAs:   RecordAsList,
				AllowEmpty: boolPtr(true),
			},
			wantErr: false,
		},
		{
			name: "title destination with record-as list is invalid",
			elem: Element{
				Destination: DestTitle,
				Options:     []string{"a", "b"},
				RecordAs:    RecordAsList,
			},
			wantErr: true,
		},
		{
			name: "title destination with record-as joined-string is valid",
			elem: Element{
				Destination: DestTitle,
				Options:     []string{"a", "b"},
				RecordAs:    RecordAsJoinedString,
			},
			wantErr: false,
		},
		{
			name: "body destination with record-as list is valid",
			elem: Element{
				Destination: DestBody,
				Options:     []string{"a", "b"},
				RecordAs:    RecordAsList,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMultiSelectElement(tt.elem)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateMultiSelectElement() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTypeInferenceFromDataType(t *testing.T) {
	// Test that elements with data-type but no type are treated as text
	yaml := `ticket-number:
  destination: body
  data-type: integer
  allow-empty: true
`
	elements, err := parseOrderedYAML([]byte(yaml))
	if err != nil {
		t.Fatalf("parseOrderedYAML() error = %v", err)
	}

	if len(elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(elements))
	}

	elem := elements[0]

	// Type should be empty in the parsed element
	if elem.Type != "" {
		t.Errorf("Type should be empty, got %q", elem.Type)
	}

	// But GetEffectiveType should infer text
	if got := GetEffectiveType(elem); got != TypeText {
		t.Errorf("GetEffectiveType() = %v, want %v", got, TypeText)
	}

	// And validation should pass
	if err := validateElement(elem); err != nil {
		t.Errorf("validateElement() should pass for inferred text type, got error: %v", err)
	}
}

func TestValidateConfig_RequiresTitleElement(t *testing.T) {
	tests := []struct {
		name     string
		elements []Element
		want     bool
	}{
		{
			name: "config with title element",
			elements: []Element{
				{Name: "title", Destination: DestTitle, Type: TypeText},
			},
			want: true,
		},
		{
			name: "config with title and body elements",
			elements: []Element{
				{Name: "title", Destination: DestTitle, Type: TypeText},
				{Name: "body", Destination: DestBody, Type: TypeText},
			},
			want: true,
		},
		{
			name: "config with only body elements",
			elements: []Element{
				{Name: "desc", Destination: DestBody, Type: TypeText},
				{Name: "notes", Destination: DestBody, Type: TypeMultilineText},
			},
			want: false,
		},
		{
			name:     "empty config",
			elements: []Element{},
			want:     false,
		},
		{
			name:     "nil elements",
			elements: nil,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Elements: tt.elements}
			if got := ValidateConfig(cfg); got != tt.want {
				t.Errorf("ValidateConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
