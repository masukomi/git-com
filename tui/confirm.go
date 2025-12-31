package tui

import (
	"os"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Confirm displays an interactive confirmation dialog
// Returns true for affirmative, false for negative
// Returns ErrAborted if cancelled
func Confirm(prompt string) (bool, error) {
	km := confirmDefaultKeymap()
	m := confirmModel{
		prompt:          prompt,
		affirmative:     "Yes",
		negative:        "No",
		confirmation:    true, // Default to Yes
		showHelp:        true,
		help:            help.New(),
		keys:            km,
		promptStyle:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99")),
		selectedStyle:   lipgloss.NewStyle().Background(lipgloss.Color("212")).Foreground(lipgloss.Color("230")).Padding(0, 3),
		unselectedStyle: lipgloss.NewStyle().Background(lipgloss.Color("235")).Foreground(lipgloss.Color("254")).Padding(0, 3),
	}

	tm, err := tea.NewProgram(m, tea.WithOutput(os.Stderr)).Run()
	if err != nil {
		return false, err
	}

	m = tm.(confirmModel)
	if m.aborted {
		return false, ErrAborted
	}

	return m.confirmation, nil
}

type confirmKeymap struct {
	Abort       key.Binding
	Quit        key.Binding
	Negative    key.Binding
	Affirmative key.Binding
	Toggle      key.Binding
	Submit      key.Binding
}

func (k confirmKeymap) FullHelp() [][]key.Binding { return nil }
func (k confirmKeymap) ShortHelp() []key.Binding {
	return []key.Binding{k.Toggle, k.Submit, k.Affirmative, k.Negative}
}

func confirmDefaultKeymap() confirmKeymap {
	return confirmKeymap{
		Abort:       key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "cancel")),
		Quit:        key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "quit")),
		Negative:    key.NewBinding(key.WithKeys("n", "N"), key.WithHelp("n", "no")),
		Affirmative: key.NewBinding(key.WithKeys("y", "Y"), key.WithHelp("y", "yes")),
		Toggle:      key.NewBinding(key.WithKeys("left", "right", "h", "l", "tab"), key.WithHelp("←→", "toggle")),
		Submit:      key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "submit")),
	}
}

type confirmModel struct {
	prompt       string
	affirmative  string
	negative     string
	quitting     bool
	aborted      bool
	showHelp     bool
	help         help.Model
	keys         confirmKeymap
	confirmation bool

	promptStyle     lipgloss.Style
	selectedStyle   lipgloss.Style
	unselectedStyle lipgloss.Style
}

func (m confirmModel) Init() tea.Cmd { return nil }

func (m confirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m, nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Abort):
			m.aborted = true
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, m.keys.Quit):
			m.confirmation = false
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, m.keys.Negative):
			m.confirmation = false
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, m.keys.Toggle):
			m.confirmation = !m.confirmation
		case key.Matches(msg, m.keys.Submit):
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, m.keys.Affirmative):
			m.confirmation = true
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m confirmModel) View() string {
	if m.quitting {
		return ""
	}

	var aff, neg string
	if m.confirmation {
		aff = m.selectedStyle.Render(m.affirmative)
		neg = m.unselectedStyle.Render(m.negative)
	} else {
		aff = m.unselectedStyle.Render(m.affirmative)
		neg = m.selectedStyle.Render(m.negative)
	}

	parts := []string{
		m.promptStyle.Render(m.prompt),
		"",
		lipgloss.JoinHorizontal(lipgloss.Left, aff, " ", neg),
	}

	if m.showHelp {
		parts = append(parts, "", m.help.View(m.keys))
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts…)
}
