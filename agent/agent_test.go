package agent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"git-com/config"

	"github.com/goccy/go-yaml"
)

// parsedElement mirrors instructionElement for test assertions.
type parsedElement struct {
	Type         string   `yaml:"type"`
	Required     bool     `yaml:"required"`
	DataType     string   `yaml:"data-type"`
	Modifiable   *bool    `yaml:"modifiable"`
	Instructions string   `yaml:"instructions"`
	Options      []string `yaml:"options"`
	Limit        int      `yaml:"limit"`
	AgentHint    string   `yaml:"agent-hint"`
}

// parseDumpOutput unmarshals DumpInstructions output into an ordered slice of
// (name, parsedElement) pairs, plus a raw map for meta element checking.
func parseDumpOutput(t *testing.T, data []byte) ([]string, map[string]parsedElement, map[string]interface{}) {
	t.Helper()

	// Strip leading comment lines before unmarshaling.
	lines := strings.Split(string(data), "\n")
	var yamlLines []string
	for _, line := range lines {
		if !strings.HasPrefix(line, "#") {
			yamlLines = append(yamlLines, line)
		}
	}

	var outer struct {
		Elements yaml.MapSlice `yaml:"elements"`
	}
	if err := yaml.UnmarshalWithOptions([]byte(strings.Join(yamlLines, "\n")), &outer, yaml.UseOrderedMap()); err != nil {
		t.Fatalf("failed to parse dump output: %v", err)
	}

	var order []string
	elements := make(map[string]parsedElement)
	raw := make(map[string]interface{})

	for _, item := range outer.Elements {
		name, _ := item.Key.(string)
		order = append(order, name)

		// Re-marshal the value and unmarshal into parsedElement.
		valueBytes, err := yaml.Marshal(item.Value)
		if err != nil {
			t.Fatalf("failed to re-marshal element %q: %v", name, err)
		}
		var pe parsedElement
		if err := yaml.Unmarshal(valueBytes, &pe); err != nil {
			t.Fatalf("failed to unmarshal element %q: %v", name, err)
		}
		elements[name] = pe

		// Also store raw for meta element inspection.
		var rawVal interface{}
		_ = yaml.Unmarshal(valueBytes, &rawVal)
		raw[name] = rawVal
	}

	return order, elements, raw
}

func boolPtr(b bool) *bool { return &b }

// --- DumpInstructions ---

func TestDumpInstructions_HeaderComment(t *testing.T) {
	cfg := &config.Config{}
	out, err := DumpInstructions(cfg)
	if err != nil {
		t.Fatalf("DumpInstructions() error = %v", err)
	}
	if !strings.HasPrefix(string(out), "# git-com agent instructions\n") {
		t.Errorf("output missing expected header comment, got:\n%s", out)
	}
	if !strings.Contains(string(out), "git-com --answers") {
		t.Errorf("output missing --answers hint, got:\n%s", out)
	}
}

func TestDumpInstructions_ElementOrder(t *testing.T) {
	cfg := &config.Config{
		Elements: []config.Element{
			{Name: "zebra", Type: config.TypeText, Destination: config.DestTitle},
			{Name: "mango", Type: config.TypeText, Destination: config.DestBody},
			{Name: "apple", Type: config.TypeSelect, Destination: config.DestTitle},
		},
	}

	out, err := DumpInstructions(cfg)
	if err != nil {
		t.Fatalf("DumpInstructions() error = %v", err)
	}

	order, _, _ := parseDumpOutput(t, out)
	expected := []string{"zebra", "mango", "apple"}
	for i, name := range expected {
		if i >= len(order) || order[i] != name {
			t.Errorf("element order[%d] = %q, want %q (full order: %v)", i, order[i], name, order)
		}
	}
}

func TestDumpInstructions_Required(t *testing.T) {
	trueVal := true
	tests := []struct {
		name         string
		elem         config.Element
		wantRequired bool
	}{
		{
			name:         "no allow-empty defaults to required",
			elem:         config.Element{Name: "e", Type: config.TypeText, Destination: config.DestTitle},
			wantRequired: true,
		},
		{
			name:         "allow-empty false is required",
			elem:         config.Element{Name: "e", Type: config.TypeText, Destination: config.DestTitle, AllowEmpty: &[]bool{false}[0]},
			wantRequired: true,
		},
		{
			name:         "allow-empty true is not required",
			elem:         config.Element{Name: "e", Type: config.TypeText, Destination: config.DestTitle, AllowEmpty: &trueVal},
			wantRequired: false,
		},
		{
			name:         "confirmation is never required",
			elem:         config.Element{Name: "e", Type: config.TypeConfirmation, Destination: config.DestTitle},
			wantRequired: true, // confirmation has no allow-empty set, so required=true per formula
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{Elements: []config.Element{tt.elem}}
			out, err := DumpInstructions(cfg)
			if err != nil {
				t.Fatalf("DumpInstructions() error = %v", err)
			}
			_, elements, _ := parseDumpOutput(t, out)
			if elements["e"].Required != tt.wantRequired {
				t.Errorf("required = %v, want %v", elements["e"].Required, tt.wantRequired)
			}
		})
	}
}

func TestDumpInstructions_DataType(t *testing.T) {
	tests := []struct {
		name         string
		elem         config.Element
		wantDataType string
	}{
		{
			name:         "text with no data-type defaults to string",
			elem:         config.Element{Name: "e", Type: config.TypeText, Destination: config.DestTitle},
			wantDataType: "string",
		},
		{
			name:         "text with integer data-type",
			elem:         config.Element{Name: "e", Type: config.TypeText, Destination: config.DestBody, DataType: config.DataTypeInteger},
			wantDataType: "integer",
		},
		{
			name:         "multiline-text defaults to string",
			elem:         config.Element{Name: "e", Type: config.TypeMultilineText, Destination: config.DestBody},
			wantDataType: "string",
		},
		{
			name:         "select has no data-type",
			elem:         config.Element{Name: "e", Type: config.TypeSelect, Destination: config.DestTitle},
			wantDataType: "",
		},
		{
			name:         "multi-select has no data-type",
			elem:         config.Element{Name: "e", Type: config.TypeMultiSelect, Destination: config.DestBody},
			wantDataType: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{Elements: []config.Element{tt.elem}}
			out, err := DumpInstructions(cfg)
			if err != nil {
				t.Fatalf("DumpInstructions() error = %v", err)
			}
			_, elements, _ := parseDumpOutput(t, out)
			if elements["e"].DataType != tt.wantDataType {
				t.Errorf("data-type = %q, want %q", elements["e"].DataType, tt.wantDataType)
			}
		})
	}
}

func TestDumpInstructions_Modifiable(t *testing.T) {
	tests := []struct {
		name           string
		elem           config.Element
		wantModifiable *bool // nil means field should be absent
	}{
		{
			name:           "select without modifiable defaults to false",
			elem:           config.Element{Name: "e", Type: config.TypeSelect, Destination: config.DestTitle},
			wantModifiable: boolPtr(false),
		},
		{
			name:           "select with modifiable true",
			elem:           config.Element{Name: "e", Type: config.TypeSelect, Destination: config.DestTitle, Modifiable: boolPtr(true)},
			wantModifiable: boolPtr(true),
		},
		{
			name:           "multi-select without modifiable defaults to false",
			elem:           config.Element{Name: "e", Type: config.TypeMultiSelect, Destination: config.DestBody},
			wantModifiable: boolPtr(false),
		},
		{
			name:           "text has no modifiable field",
			elem:           config.Element{Name: "e", Type: config.TypeText, Destination: config.DestTitle},
			wantModifiable: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{Elements: []config.Element{tt.elem}}
			out, err := DumpInstructions(cfg)
			if err != nil {
				t.Fatalf("DumpInstructions() error = %v", err)
			}
			_, elements, _ := parseDumpOutput(t, out)
			got := elements["e"].Modifiable
			if tt.wantModifiable == nil {
				if got != nil {
					t.Errorf("modifiable = %v, want absent (nil)", *got)
				}
			} else {
				if got == nil {
					t.Errorf("modifiable = nil, want %v", *tt.wantModifiable)
				} else if *got != *tt.wantModifiable {
					t.Errorf("modifiable = %v, want %v", *got, *tt.wantModifiable)
				}
			}
		})
	}
}

func TestDumpInstructions_Instructions(t *testing.T) {
	tests := []struct {
		name      string
		elem      config.Element
		wantInstr string
	}{
		{
			name:      "instructions from config",
			elem:      config.Element{Name: "e", Type: config.TypeText, Destination: config.DestTitle, Instructions: "Describe the change"},
			wantInstr: "Describe the change",
		},
		{
			name:      "placeholder used as suggestion when no instructions",
			elem:      config.Element{Name: "e", Type: config.TypeText, Destination: config.DestTitle, Placeholder: "Enter a title"},
			wantInstr: "Suggestion: Enter a title",
		},
		{
			name:      "instructions takes precedence over placeholder",
			elem:      config.Element{Name: "e", Type: config.TypeText, Destination: config.DestTitle, Instructions: "Real instructions", Placeholder: "placeholder text"},
			wantInstr: "Real instructions",
		},
		{
			name:      "no instructions or placeholder produces empty instructions field",
			elem:      config.Element{Name: "e", Type: config.TypeText, Destination: config.DestTitle},
			wantInstr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{Elements: []config.Element{tt.elem}}
			out, err := DumpInstructions(cfg)
			if err != nil {
				t.Fatalf("DumpInstructions() error = %v", err)
			}
			_, elements, _ := parseDumpOutput(t, out)
			if elements["e"].Instructions != tt.wantInstr {
				t.Errorf("instructions = %q, want %q", elements["e"].Instructions, tt.wantInstr)
			}
		})
	}
}

func TestDumpInstructions_AgentHint(t *testing.T) {
	tests := []struct {
		name     string
		elem     config.Element
		wantHint string
	}{
		{
			name:     "text title gets short hint",
			elem:     config.Element{Name: "e", Type: config.TypeText, Destination: config.DestTitle},
			wantHint: shortTitleTextHint,
		},
		{
			name:     "text body gets length hint",
			elem:     config.Element{Name: "e", Type: config.TypeText, Destination: config.DestBody},
			wantHint: shortBodyTextHint,
		},
		{
			name:     "multiline-text gets no hint",
			elem:     config.Element{Name: "e", Type: config.TypeMultilineText, Destination: config.DestTitle},
			wantHint: "",
		},
		{
			name:     "select gets no hint",
			elem:     config.Element{Name: "e", Type: config.TypeSelect, Destination: config.DestTitle},
			wantHint: "",
		},
		{
			name:     "multi-select gets no hint",
			elem:     config.Element{Name: "e", Type: config.TypeMultiSelect, Destination: config.DestBody},
			wantHint: "",
		},
		{
			name:     "confirmation gets no hint",
			elem:     config.Element{Name: "e", Type: config.TypeConfirmation},
			wantHint: "",
		},
		{
			name:     "config-supplied agent-hint overrides auto-generated hint",
			elem:     config.Element{Name: "e", Type: config.TypeText, Destination: config.DestTitle, AgentHint: "custom hint from config"},
			wantHint: "custom hint from config",
		},
		{
			name:     "config-supplied agent-hint works on non-text types",
			elem:     config.Element{Name: "e", Type: config.TypeSelect, Destination: config.DestTitle, Options: []string{"a"}, AgentHint: "pick carefully"},
			wantHint: "pick carefully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{Elements: []config.Element{tt.elem}}
			out, err := DumpInstructions(cfg)
			if err != nil {
				t.Fatalf("DumpInstructions() error = %v", err)
			}
			_, elements, _ := parseDumpOutput(t, out)
			if elements["e"].AgentHint != tt.wantHint {
				t.Errorf("agent-hint = %q, want %q", elements["e"].AgentHint, tt.wantHint)
			}
		})
	}
}

func TestDumpInstructions_Limit(t *testing.T) {
	tests := []struct {
		name      string
		elem      config.Element
		wantLimit int
	}{
		{
			name:      "limit > 0 is included",
			elem:      config.Element{Name: "e", Type: config.TypeMultiSelect, Destination: config.DestBody, Limit: 3},
			wantLimit: 3,
		},
		{
			name:      "limit == 0 is omitted",
			elem:      config.Element{Name: "e", Type: config.TypeMultiSelect, Destination: config.DestBody, Limit: 0},
			wantLimit: 0,
		},
		{
			name:      "text element limit is omitted",
			elem:      config.Element{Name: "e", Type: config.TypeText, Destination: config.DestTitle, Limit: 5},
			wantLimit: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{Elements: []config.Element{tt.elem}}
			out, err := DumpInstructions(cfg)
			if err != nil {
				t.Fatalf("DumpInstructions() error = %v", err)
			}
			_, elements, _ := parseDumpOutput(t, out)
			if elements["e"].Limit != tt.wantLimit {
				t.Errorf("limit = %d, want %d", elements["e"].Limit, tt.wantLimit)
			}
		})
	}
}

func TestDumpInstructions_Options(t *testing.T) {
	cfg := &config.Config{
		Elements: []config.Element{
			{
				Name:        "pick-one",
				Type:        config.TypeSelect,
				Destination: config.DestTitle,
				Options:     []string{"fix", "add", "refactor"},
			},
		},
	}

	out, err := DumpInstructions(cfg)
	if err != nil {
		t.Fatalf("DumpInstructions() error = %v", err)
	}

	_, elements, _ := parseDumpOutput(t, out)
	want := []string{"fix", "add", "refactor"}
	if len(elements["pick-one"].Options) != len(want) {
		t.Fatalf("options length = %d, want %d", len(elements["pick-one"].Options), len(want))
	}
	for i, opt := range want {
		if elements["pick-one"].Options[i] != opt {
			t.Errorf("options[%d] = %q, want %q", i, elements["pick-one"].Options[i], opt)
		}
	}
}

func TestDumpInstructions_MetaElementsExcluded(t *testing.T) {
	cfg := &config.Config{
		Elements: []config.Element{
			{Name: "title", Type: config.TypeText, Destination: config.DestTitle},
		},
		MetaElements: []config.MetaElement{
			{
				Name: "meta_element_changelog",
				Value: yaml.MapSlice{
					{Key: "output_format", Value: "markdown"},
					{Key: "categories", Value: []interface{}{"fix", "add"}},
				},
			},
		},
	}

	out, err := DumpInstructions(cfg)
	if err != nil {
		t.Fatalf("DumpInstructions() error = %v", err)
	}

	order, _, _ := parseDumpOutput(t, out)

	// Only regular elements should appear — meta elements must be absent.
	if len(order) != 1 {
		t.Fatalf("expected 1 entry under elements, got %d: %v", len(order), order)
	}
	if order[0] != "title" {
		t.Errorf("order[0] = %q, want 'title'", order[0])
	}
	if strings.Contains(string(out), "meta_element_changelog") {
		t.Errorf("output must not contain meta element key:\n%s", out)
	}
	if strings.Contains(string(out), "output_format") {
		t.Errorf("output must not contain meta element content:\n%s", out)
	}
}

func TestDumpInstructions_ExcludedFields(t *testing.T) {
	format := "%-12s"
	cfg := &config.Config{
		Elements: []config.Element{
			{
				Name:         "e",
				Type:         config.TypeText,
				Destination:  config.DestTitle,
				BeforeString: "[",
				AfterString:  "]",
				Format:       &format,
				Placeholder:  "Enter value",
				AllowEmpty:   boolPtr(true),
			},
		},
	}

	out, err := DumpInstructions(cfg)
	if err != nil {
		t.Fatalf("DumpInstructions() error = %v", err)
	}

	outStr := string(out)
	for _, excluded := range []string{"before-string", "after-string", "format", "placeholder", "allow-empty"} {
		if strings.Contains(outStr, excluded+":") {
			t.Errorf("output contains excluded field %q:\n%s", excluded, outStr)
		}
	}
}

// --- LoadAnswers ---

func TestLoadAnswers_ValidFile(t *testing.T) {
	content := `change-type: "fix"
commit-title: "fixed a bug"
ticket-number: 1234
code-sections:
  - core
  - tests
`
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "answers.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	answers, err := LoadAnswers(path)
	if err != nil {
		t.Fatalf("LoadAnswers() error = %v", err)
	}

	if answers["change-type"] != "fix" {
		t.Errorf("change-type = %v, want 'fix'", answers["change-type"])
	}
	if answers["commit-title"] != "fixed a bug" {
		t.Errorf("commit-title = %v, want 'fixed a bug'", answers["commit-title"])
	}
	// ticket-number without quotes is decoded as int
	if CoerceToString(answers["ticket-number"]) != "1234" {
		t.Errorf("ticket-number coerced = %q, want '1234'", CoerceToString(answers["ticket-number"]))
	}
	// code-sections is a list
	sections, err := CoerceToStringSlice(answers["code-sections"])
	if err != nil {
		t.Fatalf("CoerceToStringSlice error = %v", err)
	}
	if len(sections) != 2 || sections[0] != "core" || sections[1] != "tests" {
		t.Errorf("code-sections = %v, want [core tests]", sections)
	}
}

func TestLoadAnswers_FileNotFound(t *testing.T) {
	_, err := LoadAnswers("/nonexistent/path/answers.yaml")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
	if !strings.Contains(err.Error(), "cannot read answers file") {
		t.Errorf("error message = %q, want to contain 'cannot read answers file'", err.Error())
	}
}

func TestLoadAnswers_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "answers.yaml")
	if err := os.WriteFile(path, []byte("invalid: yaml: [unclosed"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadAnswers(path)
	if err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}
	if !strings.Contains(err.Error(), "cannot parse answers file") {
		t.Errorf("error message = %q, want to contain 'cannot parse answers file'", err.Error())
	}
}

// --- CoerceToString ---

func TestCoerceToString(t *testing.T) {
	tests := []struct {
		name string
		raw  interface{}
		want string
	}{
		{"nil", nil, ""},
		{"string", "hello", "hello"},
		{"empty string", "", ""},
		{"int", 42, "42"},
		{"float64", 3.14, "3.14"},
		{"bool true", true, "true"},
		{"bool false", false, "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CoerceToString(tt.raw)
			if got != tt.want {
				t.Errorf("CoerceToString(%v) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

// --- CoerceToStringSlice ---

func TestCoerceToStringSlice(t *testing.T) {
	tests := []struct {
		name    string
		raw     interface{}
		want    []string
		wantErr bool
	}{
		{"nil", nil, []string{}, false},
		{"empty string", "", []string{}, false},
		{"single string", "core", []string{"core"}, false},
		{"string slice", []string{"a", "b"}, []string{"a", "b"}, false},
		{
			"interface slice of strings",
			[]interface{}{"core", "docs"},
			[]string{"core", "docs"},
			false,
		},
		{
			"interface slice with ints (coerced)",
			[]interface{}{1, 2},
			[]string{"1", "2"},
			false,
		},
		{"unsupported type", 42, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CoerceToStringSlice(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Fatalf("CoerceToStringSlice() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("len = %d, want %d (got %v, want %v)", len(got), len(tt.want), got, tt.want)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("index %d = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}
