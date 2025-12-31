package tui

import (
	"errors"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/paginator"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ErrAborted is returned when the user cancels the selection
var ErrAborted = errors.New("user aborted")

// Choose displays an interactive selection list and returns the selected items
func Choose(options []string, limit int, instructions string) ([]string, error) {
	if len(options) == 0 {
		return nil, errors.New("no options provided")
	}

	noLimit := limit <= 0
	if noLimit {
		limit = len(options)
	}

	// Build items
	items := make([]chooseItem, len(options))
	for i, opt := range options {
		items[i] = chooseItem{text: opt}
	}

	// Set up paginator
	height := 10
	if len(items) < height {
		height = len(items)
	}

	p := paginator.New()
	p.SetTotalPages((len(items) + height - 1) / height)
	p.PerPage = height
	p.Type = paginator.Dots

	km := chooseDefaultKeymap()
	if noLimit || limit > 1 {
		km.Toggle.SetEnabled(true)
	}
	if noLimit {
		km.ToggleAll.SetEnabled(true)
	}

	// For single select, we don't need prefixes
	selectedPrefix := "✓ "
	unselectedPrefix := "• "
	cursorPrefix := "• "
	if limit == 1 {
		selectedPrefix = ""
		unselectedPrefix = ""
		cursorPrefix = ""
	}

	m := chooseModel{
		height:           height,
		cursor:           "> ",
		selectedPrefix:   selectedPrefix,
		unselectedPrefix: unselectedPrefix,
		cursorPrefix:     cursorPrefix,
		header:           instructions,
		items:            items,
		limit:            limit,
		paginator:        p,
		showHelp:         true,
		help:             help.New(),
		keymap:           km,
		cursorStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("212")),
		itemStyle:        lipgloss.NewStyle(),
		selectedItemStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("212")),
	}

	tm, err := tea.NewProgram(m, tea.WithOutput(os.Stderr)).Run()
	if err != nil {
		return nil, err
	}

	m = tm.(chooseModel)
	if !m.submitted {
		return nil, ErrAborted
	}

	var selected []string
	for _, item := range m.items {
		if item.selected {
			selected = append(selected, item.text)
		}
	}

	return selected, nil
}

type chooseItem struct {
	text     string
	selected bool
	order    int
}

type chooseKeymap struct {
	Down, Up, Right, Left, Home, End key.Binding
	ToggleAll, Toggle                key.Binding
	Abort, Quit, Submit              key.Binding
}

func (k chooseKeymap) FullHelp() [][]key.Binding { return nil }
func (k chooseKeymap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Toggle,
		key.NewBinding(key.WithKeys("↑", "↓"), key.WithHelp("↑↓", "navigate")),
		k.Submit,
	}
}

func chooseDefaultKeymap() chooseKeymap {
	return chooseKeymap{
		Down:      key.NewBinding(key.WithKeys("down", "j", "ctrl+n")),
		Up:        key.NewBinding(key.WithKeys("up", "k", "ctrl+p")),
		Right:     key.NewBinding(key.WithKeys("right", "l")),
		Left:      key.NewBinding(key.WithKeys("left", "h")),
		Home:      key.NewBinding(key.WithKeys("g", "home")),
		End:       key.NewBinding(key.WithKeys("G", "end")),
		ToggleAll: key.NewBinding(key.WithKeys("a", "A", "ctrl+a"), key.WithHelp("ctrl+a", "select all"), key.WithDisabled()),
		Toggle:    key.NewBinding(key.WithKeys(" ", "tab", "x"), key.WithHelp("space", "toggle"), key.WithDisabled()),
		Abort:     key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "abort")),
		Quit:      key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "quit")),
		Submit:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "submit")),
	}
}

type chooseModel struct {
	height           int
	cursor           string
	selectedPrefix   string
	unselectedPrefix string
	cursorPrefix     string
	header           string
	items            []chooseItem
	quitting         bool
	submitted        bool
	index            int
	limit            int
	numSelected      int
	currentOrder     int
	paginator        paginator.Model
	showHelp         bool
	help             help.Model
	keymap           chooseKeymap
	cursorStyle      lipgloss.Style
	headerStyle      lipgloss.Style
	itemStyle        lipgloss.Style
	selectedItemStyle lipgloss.Style
}

func (m chooseModel) Init() tea.Cmd { return nil }

func (m chooseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m, nil
	case tea.KeyMsg:
		start, end := m.paginator.GetSliceBounds(len(m.items))
		km := m.keymap
		switch {
		case key.Matches(msg, km.Down):
			m.index++
			if m.index >= len(m.items) {
				m.index = 0
				m.paginator.Page = 0
			}
			if m.index >= end {
				m.paginator.NextPage()
			}
		case key.Matches(msg, km.Up):
			m.index--
			if m.index < 0 {
				m.index = len(m.items) - 1
				m.paginator.Page = m.paginator.TotalPages - 1
			}
			if m.index < start {
				m.paginator.PrevPage()
			}
		case key.Matches(msg, km.Right):
			m.paginator.NextPage()
			m.index = min(m.index+m.height, len(m.items)-1)
		case key.Matches(msg, km.Left):
			m.paginator.PrevPage()
			m.index = max(m.index-m.height, 0)
		case key.Matches(msg, km.End):
			m.index = len(m.items) - 1
			m.paginator.Page = m.paginator.TotalPages - 1
		case key.Matches(msg, km.Home):
			m.index = 0
			m.paginator.Page = 0
		case key.Matches(msg, km.ToggleAll):
			if m.limit <= 1 {
				break
			}
			if m.numSelected < len(m.items) && m.numSelected < m.limit {
				m = m.selectAll()
			} else {
				m = m.deselectAll()
			}
		case key.Matches(msg, km.Quit):
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, km.Abort):
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, km.Toggle):
			if m.limit == 1 {
				break
			}
			if m.items[m.index].selected {
				m.items[m.index].selected = false
				m.numSelected--
			} else if m.numSelected < m.limit {
				m.items[m.index].selected = true
				m.items[m.index].order = m.currentOrder
				m.numSelected++
				m.currentOrder++
			}
		case key.Matches(msg, km.Submit):
			m.quitting = true
			// If nothing is selected, select the cursor item
			if m.numSelected < 1 {
				m.items[m.index].selected = true
			}
			m.submitted = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.paginator, cmd = m.paginator.Update(msg)
	return m, cmd
}

func (m chooseModel) selectAll() chooseModel {
	for i := range m.items {
		if m.numSelected >= m.limit {
			break
		}
		if m.items[i].selected {
			continue
		}
		m.items[i].selected = true
		m.items[i].order = m.currentOrder
		m.numSelected++
		m.currentOrder++
	}
	return m
}

func (m chooseModel) deselectAll() chooseModel {
	for i := range m.items {
		m.items[i].selected = false
		m.items[i].order = 0
	}
	m.numSelected = 0
	m.currentOrder = 0
	return m
}

func (m chooseModel) View() string {
	if m.quitting {
		return ""
	}

	var s strings.Builder
	start, end := m.paginator.GetSliceBounds(len(m.items))

	for i, item := range m.items[start:end] {
		if i == m.index%m.height {
			s.WriteString(m.cursorStyle.Render(m.cursor))
		} else {
			s.WriteString(strings.Repeat(" ", lipgloss.Width(m.cursor)))
		}

		if item.selected {
			s.WriteString(m.selectedItemStyle.Render(m.selectedPrefix + item.text))
		} else if i == m.index%m.height {
			s.WriteString(m.cursorStyle.Render(m.cursorPrefix + item.text))
		} else {
			s.WriteString(m.itemStyle.Render(m.unselectedPrefix + item.text))
		}
		if i != m.height-1 && i != len(m.items[start:end])-1 {
			s.WriteRune('\n')
		}
	}

	if m.paginator.TotalPages > 1 {
		s.WriteString(strings.Repeat("\n", m.height-m.paginator.ItemsOnPage(len(m.items))+1))
		s.WriteString("  " + m.paginator.View())
	}

	var parts []string
	if m.header != "" {
		parts = append(parts, m.headerStyle.Render(m.header))
	}
	parts = append(parts, s.String())
	if m.showHelp {
		parts = append(parts, "", m.help.View(m.keymap))
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}
