package tui

import (
	"context"
	"errors"
	"fmt"
	"github.com/fsdevblog/gophkeeper/internal/cliapp/api"
	"github.com/google/uuid"
	"net/http"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type LoginForm struct {
	Username   textinput.Model
	Password   textinput.Model
	api        *api.Client
	focusIndex int
}

func NewLoginForm(apiClient *api.Client) *LoginForm {
	username := textinput.New()
	username.Placeholder = "Username"
	username.CharLimit = 15
	username.Width = 15
	username.Focus()

	password := textinput.New()
	password.Width = 15
	password.Placeholder = "Password"
	password.EchoMode = textinput.EchoPassword

	return &LoginForm{
		Username:   username,
		Password:   password,
		focusIndex: 0,
		api:        apiClient,
	}
}

type LoginResult struct {
	Username string
	ID       uuid.UUID
	Token    string
}

func (l *LoginForm) Submit(ctx context.Context) (*LoginResult, error) {
	resp, token, err := l.api.Login(ctx, api.LoginParams{
		Username: l.Username.Value(),
		Password: l.Password.Value(),
	})
	if err != nil {
		var errResponse *api.UnexpectedHTTPStatusCodeError
		if errors.As(err, &errResponse) {
			switch errResponse.StatusCode {
			case http.StatusUnauthorized:
				return nil, ErrLoginCredentialsWrong
			}
			return nil, fmt.Errorf("login: %w", errResponse)
		}
		return nil, fmt.Errorf("login: %w", err)
	}
	return &LoginResult{
		Username: resp.Username,
		ID:       resp.ID,
		Token:    token,
	}, nil
}

func (l *LoginForm) UpdateValues(msg tea.Msg) tea.Cmd {
	var cmds = make([]tea.Cmd, len(l.inputs()))

	l.Username, cmds[0] = l.Username.Update(msg)
	l.Password, cmds[1] = l.Password.Update(msg)
	return tea.Batch(cmds...)
}

func (l *LoginForm) Render() string {
	var b strings.Builder

	title := lipgloss.NewStyle().
		Bold(true).
		Render("Please login")

	b.WriteString(title)
	b.WriteRune('\n')

	b.WriteString(l.Username.View() + "\n")
	b.WriteString(l.Password.View() + "\n")
	if l.Password.Err != nil {
		errMsg := errStyle.Render(l.Password.Err.Error())
		b.WriteString(errMsg + "\n")
	}

	b.WriteString("\nUsing:\n")
	b.WriteString("Tab/Shift+Tab: switching form fields\n")
	b.WriteString("Enter: submit the form\n")
	b.WriteString("Esc: back to main\n")

	return b.String()
}
func (l *LoginForm) IncrementFocus() tea.Cmd {
	var cmd tea.Cmd
	if l.focusIndex < len(l.inputs())-1 {
		l.inputs()[l.focusIndex].Blur()
		l.focusIndex++
		cmd = l.inputs()[l.focusIndex].Focus()
	}
	return cmd
}

func (l *LoginForm) DecrementFocus() tea.Cmd {
	var cmd tea.Cmd
	if l.focusIndex > 0 {
		l.inputs()[l.focusIndex].Blur()
		l.focusIndex--
		cmd = l.inputs()[l.focusIndex].Focus()
	}
	return cmd
}

func (l *LoginForm) IsValid() bool {
	for _, model := range l.inputs() {
		if model.Err != nil || model.Value() == "" {
			return false
		}
	}
	return true
}

func (l *LoginForm) inputs() []*textinput.Model {
	return []*textinput.Model{&l.Username, &l.Password}
}
