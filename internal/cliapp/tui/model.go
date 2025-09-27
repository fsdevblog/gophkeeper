package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsdevblog/gophkeeper/internal/cliapp/api"
)

type Model struct {
	CurrentState appState

	loginForm    *LoginForm
	registerForm *RegisterForm
	mainMenu     *MainMenu
	api          *api.Client
	alertBanner  string
	spinner      spinner.Model
}

func NewModel() (*Model, error) {
	client, err := api.New("http://localhost:8080")
	if err != nil {
		return nil, fmt.Errorf("initialization terminal UI state: %w", err)
	}

	s := spinner.New()
	s.Spinner = spinner.Line
	return &Model{
		CurrentState: AppStateMain,
		loginForm:    NewLoginForm(),
		registerForm: NewRegisterForm(client),
		mainMenu:     NewMainMenu(),
		api:          client,
		spinner:      s,
	}, nil
}

func (m Model) Init() tea.Cmd {
	cmds := make([]tea.Cmd, 2) //nolint:mnd
	cmds[0] = textinput.Blink
	cmds[1] = m.spinner.Tick
	return tea.Batch(cmds...)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	switch t := msg.(type) {
	case tea.KeyMsg:
		userInput := t.String()
		switch userInput {
		case ctrlC, ctrlQ:
			return m, tea.Quit
		case up, down, tab, shiftTab, enter, esc:
			if userInput == enter || userInput == esc {
				m.alertBanner = ""
			}
			switch m.CurrentState {
			case AppStateMain:
				return m.handleMain(userInput)
			case AppStateLogin:
				return m.handleLogin(ctx, userInput)
			case AppStateRegister:
				return m.handleRegister(ctx, userInput)
			}
		default:
			//nolint:exhaustive
			switch m.CurrentState {
			case AppStateLogin:
				cmd := m.loginForm.UpdateValues(msg)
				return m, cmd
			case AppStateRegister:
				cmd := m.registerForm.UpdateValues(msg)
				return m, cmd
			}
		}
	case cursor.BlinkMsg:
		return m, m.UpdateForms(msg)
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *Model) UpdateForms(msg tea.Msg) tea.Cmd {
	var cmds = make([]tea.Cmd, 2) //nolint:mnd
	cmds[0] = m.loginForm.UpdateValues(msg)
	cmds[1] = m.registerForm.UpdateValues(msg)
	return tea.Batch(cmds...)
}
func (m Model) View() string {
	var b strings.Builder
	if m.alertBanner != "" {
		b.WriteString(alertBannerStyle.Render(m.alertBanner) + "\n")
	}

	switch m.CurrentState {
	case AppStateMain:
		b.WriteString(m.mainMenu.Render())
	case AppStateLogin:
		b.WriteString(m.loginForm.Render())
	case AppStateRegister:
		b.WriteString(m.registerForm.Render())
	}
	b.WriteString(m.spinner.View())
	return mainStyle.Render("\n" + b.String() + "\n\n")
}
