package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/zalando/go-keyring"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsdevblog/gophkeeper/internal/cliapp/api"
)

type Model struct {
	currentState appState
	currentUser  *User

	loginForm      *LoginForm
	registerForm   *RegisterForm
	mainMenu       *MainMenu
	authorizedMain *AuthorizedMain

	api         *api.Client
	alertBanner string
}

func NewModel() (*Model, error) {
	client, err := api.New("http://localhost:8080")
	if err != nil {
		return nil, fmt.Errorf("initialization terminal UI state: %w", err)
	}
	currentUser := new(User)
	authorizedMain := NewAuthorizedMain(currentUser)

	return &Model{
		currentState:   AppStateMain,
		loginForm:      NewLoginForm(),
		registerForm:   NewRegisterForm(client),
		mainMenu:       NewMainMenu(),
		authorizedMain: authorizedMain,
		currentUser:    currentUser,
		api:            client,
	}, nil
}

func (m Model) Init() tea.Cmd {
	cmds := make([]tea.Cmd, 2) //nolint:mnd
	cmds[0] = textinput.Blink
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
			switch m.currentState {
			case AppStateMain:
				m.handleMain(userInput)
				return m, nil
			case AppStateLogin:
				return m, m.handleLogin(ctx, userInput)
			case AppStateRegister:
				return m, m.handleRegister(ctx, userInput)
			case AppStateAuthorizedMain:
				return m, nil
			}
		default:
			//nolint:exhaustive
			switch m.currentState {
			case AppStateLogin:
				cmd := m.loginForm.UpdateValues(msg)
				return m, cmd
			case AppStateRegister:
				cmd := m.registerForm.UpdateValues(msg)
				return m, cmd
			case AppStateAuthorizedMain:
				return m, nil
			}
		}
	case cursor.BlinkMsg:
		return m, m.UpdateForms(msg)
	}
	return m, nil
}

func (m Model) UpdateForms(msg tea.Msg) tea.Cmd {
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

	switch m.currentState {
	case AppStateMain:
		b.WriteString(m.mainMenu.Render())
	case AppStateLogin:
		b.WriteString(m.loginForm.Render())
	case AppStateRegister:
		b.WriteString(m.registerForm.Render())
	case AppStateAuthorizedMain:
		m.authorizedMain = NewAuthorizedMain(m.currentUser)
		b.WriteString(m.authorizedMain.Render())
	}
	return mainStyle.Render("\n" + b.String() + "\n\n")
}

func (m *Model) authenticateUser(user User, token string) error {
	if errStore := keyring.Set("gophkeeper", user.Username, token); errStore != nil {
		return fmt.Errorf("authorize user: %w", errStore)
	}
	m.currentUser = &user
	return nil
}
