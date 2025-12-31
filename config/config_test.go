package config

import (
	"os"
	"path/filepath"
	"testing"
)

// Helper to create a bool pointer
func boolPtr(b bool) *bool {
	return &b
}

// --- types.go tests ---

func TestIsAllowEmpty(t *testing.T) {
	tests := []struct {
		name     string
		elem     Element
		expected bool
	}{
		{"nil pointer", Element{}, false},
		{"false value", Element{AllowEmpty: boolPtr(false)}, false},
		{"true value", Element{AllowEmpty: boolPtr(true)}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.elem.IsAllowEmpty(); got != tt.expected {
				t.Errorf("IsAllowEmpty() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsModifiable(t *testing.T) {
	tests := []struct {
		name     string
		elem     Element
		expected bool
	}{
		{"nil pointer", Element{}, false},
		{"false value", Element{Modifiable: boolPtr(false)}, false},
		{"true value", Element{Modifiable: boolPtr(true)}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.elem.IsModifiable(); got != tt.expected {
				t.Errorf("IsModifiable() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetBulletString(t *testing.T) {
	tests := []struct {
		name     string
		elem     Element
		expected string
	}{
		{"default value", Element{}, "- "},
		{"custom value", Element{BulletString: "* "}, "* "},
		{"empty string uses default", Element{BulletString: ""}, "- "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.elem.GetBulletString(); got != tt.expected {
				t.Errorf("GetBulletString() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestGetJoinString(t *testing.T) {
	tests := []struct {
		name     string
		elem     Element
		expected string
	}{
		{"default value", Element{}, ", "},
		{"custom value", Element{JoinString: " | "}, " | "},
		{"empty string uses default", Element{JoinString: ""}, ", "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.elem.GetJoinString(); got != tt.expected {
				t.Errorf("GetJoinString() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestGetEmptySelectionText(t *testing.T) {
	tests := []struct {
		name     string
		elem     Element
		expected string
	}{
		{"default value", Element{}, "No Selection"},
		{"custom value", Element{EmptySelectionText: "Skip"}, "Skip"},
		{"empty string uses default", Element{EmptySelectionText: ""}, "No Selection"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.elem.GetEmptySelectionText(); got != tt.expected {
				t.Errorf("GetEmptySelectionText() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestHasEmptySelectionText(t *testing.T) {
	tests := []struct {
		name     string
		elem     Element
		expected bool
	}{
		{"not set", Element{}, false},
		{"empty string", Element{EmptySelectionText: ""}, false},
		{"set value", Element{EmptySelectionText: "Skip"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.elem.HasEmptySelectionText(); got != tt.expected {
				t.Errorf("HasEmptySelectionText() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// --- config.go tests ---

func TestLoadConfigFromPath(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".git-com.yaml")

		yaml := `code-section:
  destination: title
  type: select
  options:
    - input
    - git
commit-title:
  destination: title
  type: text
  placeholder: Enter title
`
		if err := os.WriteFile(configPath, []byte(yaml), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := LoadConfigFromPath(configPath)
		if err != nil {
			t.Fatalf("LoadConfigFromPath() error = %v", err)
		}

		if len(cfg.Elements) != 2 {
			t.Errorf("expected 2 elements, got %d", len(cfg.Elements))
		}
		if cfg.Elements[0].Name != "code-section" {
			t.Errorf("expected first element name 'code-section', got %q", cfg.Elements[0].Name)
		}
		if cfg.Elements[1].Name != "commit-title" {
			t.Errorf("expected second element name 'commit-title', got %q", cfg.Elements[1].Name)
		}
		if cfg.FilePath != configPath {
			t.Errorf("expected FilePath %q, got %q", configPath, cfg.FilePath)
		}
	})

	t.Run("missing file", func(t *testing.T) {
		_, err := LoadConfigFromPath("/nonexistent/path/.git-com.yaml")
		if err != ErrConfigNotFound {
			t.Errorf("expected ErrConfigNotFound, got %v", err)
		}
	})

	t.Run("invalid YAML", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".git-com.yaml")

		if err := os.WriteFile(configPath, []byte("invalid: yaml: content:"), 0644); err != nil {
			t.Fatal(err)
		}

		_, err := LoadConfigFromPath(configPath)
		if err == nil {
			t.Error("expected error for invalid YAML, got nil")
		}
	})

	t.Run("empty file", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".git-com.yaml")

		if err := os.WriteFile(configPath, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		cfg, err := LoadConfigFromPath(configPath)
		if err != nil {
			t.Fatalf("LoadConfigFromPath() error = %v", err)
		}
		if len(cfg.Elements) != 0 {
			t.Errorf("expected 0 elements for empty file, got %d", len(cfg.Elements))
		}
	})
}

func TestConfigFileNames(t *testing.T) {
	// Verify both .yaml and .yml extensions are supported
	if len(configFileNames) != 2 {
		t.Errorf("expected 2 config file names, got %d", len(configFileNames))
	}
	if configFileNames[0] != ".git-com.yaml" {
		t.Errorf("expected first config file name '.git-com.yaml', got %q", configFileNames[0])
	}
	if configFileNames[1] != ".git-com.yml" {
		t.Errorf("expected second config file name '.git-com.yml', got %q", configFileNames[1])
	}
}

func TestLoadConfigFromPath_YmlExtension(t *testing.T) {
	// Verify .yml files can be loaded via LoadConfigFromPath
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".git-com.yml")

	yaml := `title:
  destination: title
  type: text
`
	if err := os.WriteFile(configPath, []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfigFromPath(configPath)
	if err != nil {
		t.Fatalf("LoadConfigFromPath() error = %v", err)
	}

	if len(cfg.Elements) != 1 {
		t.Errorf("expected 1 element, got %d", len(cfg.Elements))
	}
	if cfg.Elements[0].Name != "title" {
		t.Errorf("expected element name 'title', got %q", cfg.Elements[0].Name)
	}
}

func TestParseOrderedYAML(t *testing.T) {
	t.Run("preserves order", func(t *testing.T) {
		yaml := `first:
  destination: title
  type: text
second:
  destination: body
  type: text
third:
  destination: title
  type: select
  options:
    - a
`
		elements, err := parseOrderedYAML([]byte(yaml))
		if err != nil {
			t.Fatalf("parseOrderedYAML() error = %v", err)
		}

		if len(elements) != 3 {
			t.Fatalf("expected 3 elements, got %d", len(elements))
		}

		expectedNames := []string{"first", "second", "third"}
		for i, name := range expectedNames {
			if elements[i].Name != name {
				t.Errorf("element %d: expected name %q, got %q", i, name, elements[i].Name)
			}
		}
	})

	t.Run("parses all fields", func(t *testing.T) {
		yaml := `test-element:
  destination: body
  type: multi-select
  instructions: Pick some options
  before-string: "["
  after-string: "]"
  allow-empty: true
  options:
    - option1
    - option2
  modifiable: true
  record-as: joined-string
  join-string: ", "
  limit: 3
  empty-selection-text: None
`
		elements, err := parseOrderedYAML([]byte(yaml))
		if err != nil {
			t.Fatalf("parseOrderedYAML() error = %v", err)
		}

		if len(elements) != 1 {
			t.Fatalf("expected 1 element, got %d", len(elements))
		}

		elem := elements[0]
		if elem.Name != "test-element" {
			t.Errorf("Name = %q, want 'test-element'", elem.Name)
		}
		if elem.Destination != DestBody {
			t.Errorf("Destination = %q, want 'body'", elem.Destination)
		}
		if elem.Type != TypeMultiSelect {
			t.Errorf("Type = %q, want 'multi-select'", elem.Type)
		}
		if elem.Instructions != "Pick some options" {
			t.Errorf("Instructions = %q, want 'Pick some options'", elem.Instructions)
		}
		if elem.BeforeString != "[" {
			t.Errorf("BeforeString = %q, want '['", elem.BeforeString)
		}
		if elem.AfterString != "]" {
			t.Errorf("AfterString = %q, want ']'", elem.AfterString)
		}
		if !elem.IsAllowEmpty() {
			t.Error("AllowEmpty should be true")
		}
		if len(elem.Options) != 2 {
			t.Errorf("Options length = %d, want 2", len(elem.Options))
		}
		if !elem.IsModifiable() {
			t.Error("Modifiable should be true")
		}
		if elem.RecordAs != RecordAsJoinedString {
			t.Errorf("RecordAs = %q, want 'joined-string'", elem.RecordAs)
		}
		if elem.JoinString != ", " {
			t.Errorf("JoinString = %q, want ', '", elem.JoinString)
		}
		if elem.Limit != 3 {
			t.Errorf("Limit = %d, want 3", elem.Limit)
		}
		if elem.EmptySelectionText != "None" {
			t.Errorf("EmptySelectionText = %q, want 'None'", elem.EmptySelectionText)
		}
	})
}

func TestSaveConfig(t *testing.T) {
	t.Run("saves and preserves order", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".git-com.yaml")

		cfg := &Config{
			FilePath: configPath,
			Elements: []Element{
				{
					Name:        "first",
					Destination: DestTitle,
					Type:        TypeText,
				},
				{
					Name:        "second",
					Destination: DestBody,
					Type:        TypeSelect,
					Options:     []string{"a", "b"},
				},
			},
		}

		if err := SaveConfig(cfg); err != nil {
			t.Fatalf("SaveConfig() error = %v", err)
		}

		// Reload and verify
		loaded, err := LoadConfigFromPath(configPath)
		if err != nil {
			t.Fatalf("LoadConfigFromPath() error = %v", err)
		}

		if len(loaded.Elements) != 2 {
			t.Fatalf("expected 2 elements, got %d", len(loaded.Elements))
		}
		if loaded.Elements[0].Name != "first" {
			t.Errorf("first element name = %q, want 'first'", loaded.Elements[0].Name)
		}
		if loaded.Elements[1].Name != "second" {
			t.Errorf("second element name = %q, want 'second'", loaded.Elements[1].Name)
		}
	})

	t.Run("preserves all fields", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".git-com.yaml")

		cfg := &Config{
			FilePath: configPath,
			Elements: []Element{
				{
					Name:               "test",
					Destination:        DestBody,
					Type:               TypeMultiSelect,
					Instructions:       "Test instructions",
					BeforeString:       "<<",
					AfterString:        ">>",
					AllowEmpty:         boolPtr(true),
					Options:            []string{"x", "y", "z"},
					Modifiable:         boolPtr(true),
					RecordAs:           RecordAsList,
					BulletString:       "* ",
					Limit:              2,
					EmptySelectionText: "Skip",
				},
			},
		}

		if err := SaveConfig(cfg); err != nil {
			t.Fatalf("SaveConfig() error = %v", err)
		}

		loaded, err := LoadConfigFromPath(configPath)
		if err != nil {
			t.Fatalf("LoadConfigFromPath() error = %v", err)
		}

		elem := loaded.Elements[0]
		if elem.Instructions != "Test instructions" {
			t.Errorf("Instructions = %q, want 'Test instructions'", elem.Instructions)
		}
		if elem.BeforeString != "<<" {
			t.Errorf("BeforeString = %q, want '<<'", elem.BeforeString)
		}
		if elem.AfterString != ">>" {
			t.Errorf("AfterString = %q, want '>>'", elem.AfterString)
		}
		if !elem.IsAllowEmpty() {
			t.Error("AllowEmpty should be true")
		}
		if len(elem.Options) != 3 {
			t.Errorf("Options length = %d, want 3", len(elem.Options))
		}
		if !elem.IsModifiable() {
			t.Error("Modifiable should be true")
		}
		if elem.RecordAs != RecordAsList {
			t.Errorf("RecordAs = %q, want 'list'", elem.RecordAs)
		}
		if elem.BulletString != "* " {
			t.Errorf("BulletString = %q, want '* '", elem.BulletString)
		}
		if elem.Limit != 2 {
			t.Errorf("Limit = %d, want 2", elem.Limit)
		}
		if elem.EmptySelectionText != "Skip" {
			t.Errorf("EmptySelectionText = %q, want 'Skip'", elem.EmptySelectionText)
		}
	})
}

func TestAddOptionToElement(t *testing.T) {
	t.Run("adds option to existing element", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".git-com.yaml")

		cfg := &Config{
			FilePath: configPath,
			Elements: []Element{
				{
					Name:        "my-select",
					Destination: DestTitle,
					Type:        TypeSelect,
					Options:     []string{"a", "b"},
				},
			},
		}

		// Save initial config
		if err := SaveConfig(cfg); err != nil {
			t.Fatal(err)
		}

		// Add new option
		if err := cfg.AddOptionToElement("my-select", "c"); err != nil {
			t.Fatalf("AddOptionToElement() error = %v", err)
		}

		// Verify in memory
		if len(cfg.Elements[0].Options) != 3 {
			t.Errorf("expected 3 options in memory, got %d", len(cfg.Elements[0].Options))
		}
		if cfg.Elements[0].Options[2] != "c" {
			t.Errorf("expected third option 'c', got %q", cfg.Elements[0].Options[2])
		}

		// Reload and verify persisted
		loaded, err := LoadConfigFromPath(configPath)
		if err != nil {
			t.Fatal(err)
		}
		if len(loaded.Elements[0].Options) != 3 {
			t.Errorf("expected 3 options on disk, got %d", len(loaded.Elements[0].Options))
		}
		if loaded.Elements[0].Options[2] != "c" {
			t.Errorf("expected third option 'c' on disk, got %q", loaded.Elements[0].Options[2])
		}
	})

	t.Run("returns error for missing element", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".git-com.yaml")

		cfg := &Config{
			FilePath: configPath,
			Elements: []Element{
				{
					Name:        "existing",
					Destination: DestTitle,
					Type:        TypeText,
				},
			},
		}

		err := cfg.AddOptionToElement("nonexistent", "value")
		if err == nil {
			t.Error("expected error for nonexistent element, got nil")
		}
	})
}

func TestElementToMap(t *testing.T) {
	t.Run("includes only set fields", func(t *testing.T) {
		elem := Element{
			Name:        "test",
			Destination: DestTitle,
			Type:        TypeText,
		}

		m := elementToMap(elem)

		if _, ok := m["destination"]; !ok {
			t.Error("destination should always be included")
		}
		if _, ok := m["type"]; !ok {
			t.Error("type should be included when set")
		}
		if _, ok := m["instructions"]; ok {
			t.Error("instructions should not be included when empty")
		}
		if _, ok := m["options"]; ok {
			t.Error("options should not be included when empty")
		}
	})

	t.Run("includes all set fields", func(t *testing.T) {
		elem := Element{
			Name:               "test",
			Destination:        DestBody,
			Type:               TypeMultiSelect,
			Instructions:       "Test",
			BeforeString:       "[",
			AfterString:        "]",
			AllowEmpty:         boolPtr(true),
			Options:            []string{"a"},
			Modifiable:         boolPtr(false),
			RecordAs:           RecordAsList,
			BulletString:       "- ",
			JoinString:         ", ",
			Limit:              5,
			EmptySelectionText: "None",
		}

		m := elementToMap(elem)

		expectedKeys := []string{
			"destination", "type", "instructions", "before-string", "after-string",
			"allow-empty", "options", "modifiable", "record-as", "bullet-string",
			"join-string", "limit", "empty-selection-text",
		}

		for _, key := range expectedKeys {
			if _, ok := m[key]; !ok {
				t.Errorf("expected key %q to be present", key)
			}
		}
	})
}
