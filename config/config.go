package config

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
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

	elements, metaElements, cm, err := parseOrderedYAML(data)
	if err != nil {
		return nil, err
	}

	return &Config{
		Elements:     elements,
		MetaElements: metaElements,
		FilePath:     path,
		comments:     cm,
	}, nil
}

// parseOrderedYAML parses YAML while preserving the order of top-level elements
// and capturing comments for round-trip preservation.
// Meta elements (keys starting with "meta_element_") are returned separately
// and not processed as regular elements.
func parseOrderedYAML(data []byte) ([]Element, []MetaElement, yaml.CommentMap, error) {
	cm := yaml.CommentMap{}
	var ms yaml.MapSlice
	if err := yaml.UnmarshalWithOptions(data, &ms, yaml.CommentToMap(cm), yaml.UseOrderedMap()); err != nil {
		return nil, nil, nil, err
	}

	var elements []Element
	var metaElements []MetaElement

	for _, item := range ms {
		keyStr, ok := item.Key.(string)
		if !ok {
			continue
		}

		if strings.HasPrefix(keyStr, MetaElementPrefix) {
			metaElements = append(metaElements, MetaElement{
				Name:  keyStr,
				Value: item.Value,
			})
			continue
		}

		valueBytes, err := yaml.Marshal(item.Value)
		if err != nil {
			return nil, nil, nil, err
		}
		var elem Element
		if err := yaml.Unmarshal(valueBytes, &elem); err != nil {
			return nil, nil, nil, err
		}
		elem.Name = keyStr
		elements = append(elements, elem)
	}

	return elements, metaElements, cm, nil
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
	ms := yaml.MapSlice{}
	for _, elem := range cfg.Elements {
		ms = append(ms, yaml.MapItem{Key: elem.Name, Value: elem})
	}
	for _, meta := range cfg.MetaElements {
		ms = append(ms, yaml.MapItem{Key: meta.Name, Value: meta.Value})
	}

	data, err := yaml.MarshalWithOptions(ms, yaml.WithComment(cfg.comments))
	if err != nil {
		return err
	}

	return os.WriteFile(cfg.FilePath, data, 0644)
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
