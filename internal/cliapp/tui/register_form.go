package tui

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fsdevblog/gophkeeper/internal/cliapp/api"
)

const (
	minPasswordLength = 8
	maxPasswordLength = 72 // bcrypt max password length
)

type RegisterForm struct {
	Username             textinput.Model
	Password             textinput.Model
	PasswordConfirmation textinput.Model

	focusIndex   int
	api          *api.Client
	isSubmitting bool

	spinner spinner.Model
}

func NewRegisterForm(apiClient *api.Client) *RegisterForm {
	username := textinput.New()
	username.Placeholder = "Username"
	username.CharLimit = 15
	username.Width = 15
	username.Focus()

	password := textinput.New()
	password.Width = 15
	password.Placeholder = "Password"
	password.EchoMode = textinput.EchoPassword

	passwordConfirmation := textinput.New()
	passwordConfirmation.Width = 15
	passwordConfirmation.Placeholder = "Password Confirmation"
	passwordConfirmation.EchoMode = textinput.EchoPassword

	s := spinner.New()
	s.Spinner = spinner.Points
	s.Tick()
	return &RegisterForm{
		Username:             username,
		Password:             password,
		PasswordConfirmation: passwordConfirmation,
		focusIndex:           0,
		api:                  apiClient,
		spinner:              s,
	}
}

type RegisterResult struct {
	Username string
	ID       uuid.UUID
	Token    string
}

func (r *RegisterForm) Submit(ctx context.Context) (*RegisterResult, error) {
	if !r.IsValid() {
		return nil, ErrInvalidForm
	}
	if r.isSubmitting {
		return nil, ErrAlreadySubmitting
	}
	r.isSubmitting = true
	defer func() {
		r.isSubmitting = false
	}()

	resp, token, err := r.api.Register(ctx, api.RegisterParams{
		Username: r.Username.Value(),
		Password: r.Password.Value(),
	})
	if err != nil {
		var errResponse *api.UnexpectedHTTPStatusCodeError
		if errors.As(err, &errResponse) {
			switch errResponse.StatusCode {
			case http.StatusConflict:
				return nil, ErrUserAlreadyExists
			default:
				return nil, fmt.Errorf("register: %w", errResponse)
			}
		}
		return nil, fmt.Errorf("register: %w", err)
	}
	return &RegisterResult{
		Username: resp.Username,
		ID:       resp.ID,
		Token:    token,
	}, nil
}

func (r *RegisterForm) UpdateValues(msg tea.Msg) tea.Cmd {
	var cmds = make([]tea.Cmd, 3) //nolint:mnd

	r.Username, cmds[0] = r.Username.Update(msg)
	r.Password, cmds[1] = r.Password.Update(msg)

	r.Password.Validate = func(s string) error {
		if len(s) < minPasswordLength {
			return errors.New("too short password")
		}
		b := []byte(s)
		if len(b) > maxPasswordLength {
			return errors.New("too long password")
		}
		return nil
	}

	r.PasswordConfirmation, cmds[2] = r.PasswordConfirmation.Update(msg)

	r.PasswordConfirmation.Validate = func(s string) error {
		if s != r.Password.Value() {
			return errors.New("passwords do not match")
		}
		return nil
	}
	return tea.Batch(cmds...)
}

func (r *RegisterForm) IsValid() bool {
	for _, model := range r.inputs() {
		if model.Err != nil || model.Value() == "" {
			return false
		}
	}
	return true
}

func (r *RegisterForm) IncrementFocus() tea.Cmd {
	var cmd tea.Cmd
	if r.focusIndex < len(r.inputs())-1 {
		r.inputs()[r.focusIndex].Blur()
		r.focusIndex++
		cmd = r.inputs()[r.focusIndex].Focus()
	}
	return cmd
}

func (r *RegisterForm) DecrementFocus() tea.Cmd {
	var cmd tea.Cmd
	if r.focusIndex > 0 {
		r.inputs()[r.focusIndex].Blur()
		r.focusIndex--
		cmd = r.inputs()[r.focusIndex].Focus()
	}
	return cmd
}

func (r *RegisterForm) Render() string {
	var b strings.Builder
	title := lipgloss.NewStyle().
		Bold(true).
		Render("Registration")
	b.WriteString(title + "\n")

	b.WriteString(r.Username.View() + "\n")
	b.WriteString(r.Password.View() + "\n")
	if r.Password.Err != nil {
		b.WriteString(errStyle.Render(r.Password.Err.Error()) + "\n")
	}
	b.WriteString(r.PasswordConfirmation.View() + "\n")
	if r.PasswordConfirmation.Err != nil {
		b.WriteString(errStyle.Render(r.PasswordConfirmation.Err.Error()) + "\n")
	}

	b.WriteString("\nUsing:\n")
	b.WriteString(primaryTextStyle.Render("Tab/Shift+Tab:") + " switching form fields\n")
	b.WriteString(primaryTextStyle.Render("Enter:") + " submit the form\n")
	b.WriteString(primaryTextStyle.Render("Esc:") + " back to main\n")

	b.WriteString(r.spinner.View() + "\n")
	return b.String()
}
func (r *RegisterForm) inputs() []*textinput.Model {
	return []*textinput.Model{&r.Username, &r.Password, &r.PasswordConfirmation}
}
