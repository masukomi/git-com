package config

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

var configFileNames = []string{".git-com.yaml", ".git-com.yml"}

var (
	ErrConfigNotFound = errors.New("config file not found")
	ErrNotInGitRepo   = errors.New("not in a git repository")
)

// LoadConfig loads the configuration from the git repository root
// It checks for both .git-com.yaml and .git-com.yml, preferring .yaml
func LoadConfig() (*Config, error) {
	gitRoot, err := findGitRoot()
	if err != nil {
		return nil, err
	}

	for _, fileName := range configFileNames {
		configPath := filepath.Join(gitRoot, fileName)
		cfg, err := LoadConfigFromPath(configPath)
		if err == nil {
			return cfg, nil
		}
		if err != ErrConfigNotFound {
			return nil, err
		}
	}

	return nil, ErrConfigNotFound
}

// LoadConfigFromPath loads the configuration from a specific path
func LoadConfigFromPath(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrConfigNotFound
		}
		return nil, err
	}

	elements, err := parseOrderedYAML(data)
	if err != nil {
		return nil, err
	}

	return &Config{
		Elements: elements,
		FilePath: path,
	}, nil
}

// parseOrderedYAML parses YAML while preserving the order of elements
func parseOrderedYAML(data []byte) ([]Element, error) {
	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil {
		return nil, err
	}

	// Handle empty file
	if len(node.Content) == 0 {
		return nil, nil
	}

	// node.Content[0] is the document root (a MappingNode)
	docNode := node.Content[0]
	if docNode.Kind != yaml.MappingNode {
		return nil, errors.New("expected a mapping at the root of the YAML")
	}

	// Content contains alternating key/value nodes
	var elements []Element
	content := docNode.Content
	for i := 0; i < len(content); i += 2 {
		keyNode := content[i]
		valueNode := content[i+1]

		var elem Element
		if err := valueNode.Decode(&elem); err != nil {
			return nil, err
		}
		elem.Name = keyNode.Value
		elements = append(elements, elem)
	}

	return elements, nil
}

// findGitRoot finds the root directory of the current git repository
func findGitRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", ErrNotInGitRepo
	}
	return strings.TrimSpace(string(output)), nil
}

// SaveConfig saves the configuration back to the file
func SaveConfig(cfg *Config) error {
	// Build a map that preserves order using yaml.Node
	var docNode yaml.Node
	docNode.Kind = yaml.DocumentNode

	var mapNode yaml.Node
	mapNode.Kind = yaml.MappingNode

	for _, elem := range cfg.Elements {
		// Add key node
		var keyNode yaml.Node
		keyNode.Kind = yaml.ScalarNode
		keyNode.Value = elem.Name
		keyNode.Tag = "!!str"

		// Create element map for value
		elemMap := elementToMap(elem)

		var valueNode yaml.Node
		if err := valueNode.Encode(elemMap); err != nil {
			return err
		}

		mapNode.Content = append(mapNode.Content, &keyNode, &valueNode)
	}

	docNode.Content = append(docNode.Content, &mapNode)

	data, err := yaml.Marshal(&docNode)
	if err != nil {
		return err
	}

	return os.WriteFile(cfg.FilePath, data, 0644)
}

// elementToMap converts an Element to a map for YAML serialization
// This ensures we only include non-empty/non-default values
func elementToMap(elem Element) map[string]interface{} {
	m := make(map[string]interface{})

	m["destination"] = string(elem.Destination)

	addStringIfNotEmpty(m, "type", string(elem.Type))
	addStringIfNotEmpty(m, "instructions", elem.Instructions)
	addStringIfNotEmpty(m, "before-string", elem.BeforeString)
	addStringIfNotEmpty(m, "after-string", elem.AfterString)
	addBoolIfNotNil(m, "allow-empty", elem.AllowEmpty)
	addStringIfNotEmpty(m, "placeholder", elem.Placeholder)
	addStringIfNotEmpty(m, "data-type", string(elem.DataType))
	addOptionsIfNotEmpty(m, "options", elem.Options)
	addBoolIfNotNil(m, "modifiable", elem.Modifiable)
	addStringIfNotEmpty(m, "record-as", string(elem.RecordAs))
	addStringIfNotEmpty(m, "bullet-string", elem.BulletString)
	addStringIfNotEmpty(m, "join-string", elem.JoinString)
	addIntIfNotZero(m, "limit", elem.Limit)
	addStringIfNotEmpty(m, "empty-selection-text", elem.EmptySelectionText)

	return m
}

func addStringIfNotEmpty(m map[string]interface{}, key, value string) {
	if value != "" {
		m[key] = value
	}
}

func addBoolIfNotNil(m map[string]interface{}, key string, value *bool) {
	if value != nil {
		m[key] = *value
	}
}

func addIntIfNotZero(m map[string]interface{}, key string, value int) {
	if value != 0 {
		m[key] = value
	}
}

func addOptionsIfNotEmpty(m map[string]interface{}, key string, options []string) {
	if len(options) > 0 {
		m[key] = options
	}
}

// AddOptionToElement adds a new option to an element's options list
func (c *Config) AddOptionToElement(elementName, newOption string) error {
	for i, elem := range c.Elements {
		if elem.Name == elementName {
			c.Elements[i].Options = append(c.Elements[i].Options, newOption)
			return SaveConfig(c)
		}
	}
	return errors.New("element not found")
}
