package tui

import (
	"os"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Input displays an interactive text input and returns the entered text
func Input(placeholder string) (string, error) {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Focus()
	ti.CharLimit = 0 // No limit
	ti.Width = 60

	m := inputModel{
		textinput: ti,
		autoWidth: true,
		showHelp:  true,
		help:      help.New(),
		keymap:    inputDefaultKeymap(),
	}

	tm, err := tea.NewProgram(m, tea.WithOutput(os.Stderr)).Run()
	if err != nil {
		return "", err
	}

	m = tm.(inputModel)
	if !m.submitted {
		return "", ErrAborted
	}

	return m.textinput.Value(), nil
}

type inputKeymap struct {
	Submit key.Binding
	Abort  key.Binding
	Quit   key.Binding
}

func (k inputKeymap) FullHelp() [][]key.Binding { return nil }
func (k inputKeymap) ShortHelp() []key.Binding {
	return []key.Binding{k.Submit}
}

func inputDefaultKeymap() inputKeymap {
	return inputKeymap{
		Submit: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "submit")),
		Abort:  key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "abort")),
		Quit:   key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "quit")),
	}
}

type inputModel struct {
	textinput   textinput.Model
	autoWidth   bool
	header      string
	headerStyle lipgloss.Style
	quitting    bool
	submitted   bool
	showHelp    bool
	help        help.Model
	keymap      inputKeymap
}

func (m inputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m inputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if m.autoWidth {
			m.textinput.Width = msg.Width - lipgloss.Width(m.textinput.Prompt) - 1
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
	m.textinput, cmd = m.textinput.Update(msg)
	return m, cmd
}

func (m inputModel) View() string {
	if m.quitting {
		return ""
	}

	var parts []string
	if m.header != "" {
		parts = append(parts, m.headerStyle.Render(m.header))
	}
	parts = append(parts, m.textinput.View())
	if m.showHelp {
		parts = append(parts, "", m.help.View(m.keymap))
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}
