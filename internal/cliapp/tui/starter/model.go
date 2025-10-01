package starter

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fsdevblog/gophkeeper/internal/cliapp/api"
	"github.com/fsdevblog/gophkeeper/internal/cliapp/tui"
	"github.com/fsdevblog/gophkeeper/internal/cliapp/tui/models"
	"github.com/fsdevblog/gophkeeper/internal/cliapp/tui/views"
	"strings"
)

var _ tea.Model = &Model{}

type Model struct {
	state        tui.State
	api          *api.Client
	child        tea.Model
	session      *models.Session
	authProvider views.AuthProvider

	message string
	width   int
	height  int
}

func New(state tui.State, authProvider views.AuthProvider) (*Model, error) {
	m := &Model{
		authProvider: authProvider,
		state:        state,
		session:      new(models.Session),
	}
	if err := m.setChild(); err != nil {
		return nil, fmt.Errorf("init child: %w", err)
	}
	return m, nil
}

func (m *Model) Init() tea.Cmd {
	return m.initChild()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typedMsg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = typedMsg.Width
		m.height = typedMsg.Height
	case tui.StateMsg:
		m.state = typedMsg.NewState
		if err := m.setChild(); err != nil {
			return m, tea.Quit
		}
		m.message = ""
		return m, m.initChild()
	case tui.AuthenticateUserMsg:
		m.session = &models.Session{
			User:            typedMsg.User,
			IsAuthenticated: true,
		}

		m.state = tui.StatePrivateMenu
		_ = m.setChild()
		return m, nil
	}
	var cmd tea.Cmd
	m.child, cmd = m.child.Update(msg)
	return m, cmd
}

func (m *Model) View() string {
	b := new(strings.Builder)
	if m.message != "" {
		b.WriteString(lipgloss.Place(m.width, 1, lipgloss.Center, lipgloss.Center, m.message) + "\n\n")
	}
	b.WriteString(m.child.View())
	return b.String()
}

func (m *Model) setChild() error {
	switch m.state {
	case tui.StateLogin:
		m.child = views.NewLoginForm(m.authProvider)
	case tui.StatePublicMenu:
		m.child = views.NewPublicMenu()
	case tui.StatePrivateMenu:
		m.child = views.NewPrivateMenu()
	}
	return nil
}

func (m *Model) initChild() tea.Cmd {
	var cmds []tea.Cmd
	cmd := m.child.Init()
	cmds = append(cmds, cmd)

	m.child, cmd = m.child.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
	cmds = append(cmds, cmd)
	return tea.Batch(cmds...)
}
