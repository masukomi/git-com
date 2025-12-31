package config

// ElementType represents the type of input element
type ElementType string

const (
	TypeText          ElementType = "text"
	TypeMultilineText ElementType = "multiline-text"
	TypeSelect        ElementType = "select"
	TypeMultiSelect   ElementType = "multi-select"
	TypeConfirmation  ElementType = "confirmation"
)

// Destination represents where the element value goes
type Destination string

const (
	DestTitle Destination = "title"
	DestBody  Destination = "body"
)

// DataType for text validation
type DataType string

const (
	DataTypeString  DataType = "string"
	DataTypeInteger DataType = "integer"
	DataTypeFloat   DataType = "float"
)

// RecordAs for multi-select output format
type RecordAs string

const (
	RecordAsList         RecordAs = "list"
	RecordAsJoinedString RecordAs = "joined-string"
)

// Default text that will be used when not provided by the user
const (
	// JoinSeparator is the string used to join array elements.
    JoinSeparator		= ", "

    // BulletListPrefix is the prefix that appears at the start of every
    // line in a bullet‑point list.
    BulletListPrefix	= "- "
)

// Element represents a single YAML element configuration
type Element struct {
	Name        string      // The top-level key name (populated during parsing)
	Destination Destination `yaml:"destination"`
	Type        ElementType `yaml:"type"`

	// Optional common attributes
	Instructions string `yaml:"instructions,omitempty"`
	BeforeString string `yaml:"before-string,omitempty"`
	AfterString  string `yaml:"after-string,omitempty"`
	AllowEmpty   *bool  `yaml:"allow-empty,omitempty"` // Pointer to distinguish unset from false

	// Text-specific attributes
	Placeholder string   `yaml:"placeholder,omitempty"`
	DataType    DataType `yaml:"data-type,omitempty"`

	// Select/Multi-select attributes
	Options    []string `yaml:"options,omitempty"`
	Modifiable *bool    `yaml:"modifiable,omitempty"`

	// Multi-select specific attributes
	RecordAs           RecordAs `yaml:"record-as,omitempty"`
	BulletString       string   `yaml:"bullet-string,omitempty"`
	JoinString         string   `yaml:"join-string,omitempty"`
	Limit              int      `yaml:"limit,omitempty"`
	EmptySelectionText string   `yaml:"empty-selection-text,omitempty"`
}

// Config holds the ordered list of elements parsed from YAML
type Config struct {
	Elements []Element
	FilePath string // Path to the config file for saving modifications
}

// IsAllowEmpty returns true if empty input is allowed
func (e *Element) IsAllowEmpty() bool {
	return e.AllowEmpty != nil && *e.AllowEmpty
}

// IsModifiable returns true if the element allows adding new options
func (e *Element) IsModifiable() bool {
	return e.Modifiable != nil && *e.Modifiable
}

// GetBulletString returns the bullet string with default
func (e *Element) GetBulletString() string {
	if e.BulletString == "" {
		return BulletListPrefix
	}
	return e.BulletString
}

// GetJoinString returns the join string with default
func (e *Element) GetJoinString() string {
	if e.JoinString == "" {
		return JoinSeparator
	}
	return e.JoinString
}

// GetEmptySelectionText returns the empty selection text with default
func (e *Element) GetEmptySelectionText() string {
	if e.EmptySelectionText == "" {
		return "No Selection"
	}
	return e.EmptySelectionText
}

// HasEmptySelectionText returns true if empty-selection-text was explicitly set
func (e *Element) HasEmptySelectionText() bool {
	return e.EmptySelectionText != ""
}
