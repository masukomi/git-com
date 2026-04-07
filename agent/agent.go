package agent

import (
	"fmt"
	"os"

	"git-com/config"

	"github.com/goccy/go-yaml"
)

const shortTitleTextHint = "Keep under 50 characters if possible; shorter is better"
const shortBodyTextHint = "Keep under 120 characters if possible"

// instructionElement is the structure for each element in the dump-instructions output.
type instructionElement struct {
	Type         config.ElementType `yaml:"type"`
	Required     bool               `yaml:"required"`
	DataType     string             `yaml:"data-type,omitempty"`
	Modifiable   *bool              `yaml:"modifiable,omitempty"`
	Instructions string             `yaml:"instructions,omitempty"`
	Options      []string           `yaml:"options,omitempty"`
	Limit        int                `yaml:"limit,omitempty"`
	AgentHint    string             `yaml:"agent-hint,omitempty"`
}

// DumpInstructions generates a YAML document describing the commit schema for agents.
// meta_element_* blocks from the config are included verbatim.
func DumpInstructions(cfg *config.Config) ([]byte, error) {
	ms := yaml.MapSlice{}

	for _, elem := range cfg.Elements {
		ms = append(ms, yaml.MapItem{Key: elem.Name, Value: buildInstructionElement(elem)})
	}

	outer := yaml.MapSlice{
		{Key: "elements", Value: ms},
	}

	data, err := yaml.Marshal(outer)
	if err != nil {
		return nil, err
	}

	header := []byte("# git-com agent instructions\n# Fill in values and run: git-com -answers <answers-file>\n")
	return append(header, data...), nil
}

func buildInstructionElement(elem config.Element) instructionElement {
	effectiveType := config.GetEffectiveType(elem)

	ie := instructionElement{
		Type:     effectiveType,
		Required: !elem.IsAllowEmpty(),
	}

	// data-type for text/multiline-text only
	if effectiveType == config.TypeText || effectiveType == config.TypeMultilineText {
		if elem.DataType != "" {
			ie.DataType = string(elem.DataType)
		} else {
			ie.DataType = "string"
		}
	}

	// modifiable always present for select/multi-select
	if effectiveType == config.TypeSelect || effectiveType == config.TypeMultiSelect {
		modifiable := elem.IsModifiable()
		ie.Modifiable = &modifiable
	}

	// instructions — use placeholder as fallback with "Suggestion: " prefix
	if elem.Instructions != "" {
		ie.Instructions = elem.Instructions
	} else if elem.Placeholder != "" {
		ie.Instructions = "Suggestion: " + elem.Placeholder
	}

	// options
	if len(elem.Options) > 0 {
		ie.Options = elem.Options
	}

	// limit for multi-select only when > 0
	if effectiveType == config.TypeMultiSelect && elem.Limit > 0 {
		ie.Limit = elem.Limit
	}

	// agent-hint for text type only, based on destination
	if effectiveType == config.TypeText {
		if elem.Destination == config.DestTitle {
			ie.AgentHint = shortTitleTextHint
		} else if elem.Destination == config.DestBody {
			ie.AgentHint = shortBodyTextHint
		}
	}

	return ie
}

// LoadAnswers reads an answers YAML file and returns a raw map.
func LoadAnswers(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read answers file: %w", err)
	}

	var answers map[string]interface{}
	if err := yaml.Unmarshal(data, &answers); err != nil {
		return nil, fmt.Errorf("cannot parse answers file: %w", err)
	}

	return answers, nil
}

// CoerceToString converts an interface{} value to string for text/select elements.
// Handles nil, string, and numeric types decoded from YAML (e.g. ticket-number: 1234).
func CoerceToString(raw interface{}) string {
	if raw == nil {
		return ""
	}
	switch v := raw.(type) {
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

// CoerceToStringSlice converts an interface{} value to []string for multi-select elements.
// Accepts nil, []interface{}, []string, or a bare string (treated as single-item list).
func CoerceToStringSlice(raw interface{}) ([]string, error) {
	if raw == nil {
		return []string{}, nil
	}
	switch v := raw.(type) {
	case []interface{}:
		result := make([]string, 0, len(v))
		for _, item := range v {
			result = append(result, CoerceToString(item))
		}
		return result, nil
	case []string:
		return v, nil
	case string:
		if v == "" {
			return []string{}, nil
		}
		return []string{v}, nil
	default:
		return nil, fmt.Errorf("expected a list, got %T", raw)
	}
}
