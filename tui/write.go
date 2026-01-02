package tui

import (
	"os"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Write displays an interactive multiline text input and returns the entered text
// If initialContent is not nil, the textarea will be pre-filled with that content
func Write(placeholder string, instructions string, initialContent *string) (string, error) {
	ta := textarea.New()
	ta.Placeholder = placeholder
	ta.Focus()
	ta.CharLimit = 0 // No limit
	ta.SetWidth(80)
	ta.SetHeight(10)

	// Pre-fill with initial content if provided
	if initialContent != nil {
		ta.SetValue(*initialContent)
	}

	km := writeDefaultKeymap()
	ta.KeyMap.InsertNewline = km.InsertNewline

	m := writeModel{
		textarea:  ta,
		autoWidth: true,
		header:    instructions,
		showHelp:  true,
		help:      help.New(),
		keymap:    km,
	}

	tm, err := tea.NewProgram(m, tea.WithOutput(os.Stderr)).Run()
	if err != nil {
		return "", err
	}

	m = tm.(writeModel)
	if !m.submitted {
		return "", ErrAborted
	}

	return m.textarea.Value(), nil
}

type writeKeymap struct {
	textarea.KeyMap
	Submit key.Binding
	Abort  key.Binding
	Quit   key.Binding
}

func (k writeKeymap) FullHelp() [][]key.Binding { return nil }
func (k writeKeymap) ShortHelp() []key.Binding {
	return []key.Binding{k.InsertNewline, k.Submit}
}

func writeDefaultKeymap() writeKeymap {
	km := textarea.DefaultKeyMap
	// Keep default: Enter inserts newline (natural for multiline input)
	km.InsertNewline = key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "new line"),
	)
	return writeKeymap{
		KeyMap: km,
		Submit: key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("ctrl+d", "submit")),
		Abort:  key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "abort")),
		Quit:   key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "quit")),
	}
}

type writeModel struct {
	textarea    textarea.Model
	autoWidth   bool
	header      string
	headerStyle lipgloss.Style
	quitting    bool
	submitted   bool
	showHelp    bool
	help        help.Model
	keymap      writeKeymap
}

func (m writeModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m writeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if m.autoWidth {
			m.textarea.SetWidth(msg.Width)
		}
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.Abort):
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, m.keymap.Quit):
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, m.keymap.Submit):
			m.quitting = true
			m.submitted = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m writeModel) View() string {
	if m.quitting {
		return ""
	}

	var parts []string
	if m.header != "" {
		parts = append(parts, m.headerStyle.Render(m.header))
	}
	parts = append(parts, m.textarea.View())
	if m.showHelp {
		parts = append(parts, "", m.help.View(m.keymap))
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}
