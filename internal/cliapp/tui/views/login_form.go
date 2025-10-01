package views

import (
	"context"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/fsdevblog/gophkeeper/internal/cliapp/tui"
	"github.com/fsdevblog/gophkeeper/internal/cliapp/tui/utils"
	"strings"
)

var _ tea.Model = &LoginForm{}

type LoginForm struct {
	form         *huh.Form
	authProvider AuthProvider
	formData     *LoginFormData
	width        int
	height       int
	errorMessage string
}

type LoginFormData struct {
	Username string
	Password string
}

func NewLoginForm(authProvider AuthProvider) *LoginForm {
	formData := new(LoginFormData)
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Value(&formData.Username).
				Validate(utils.ValidateUsernameField).
				Title("Username"),
			huh.NewInput().
				Value(&formData.Password).
				Validate(utils.ValidatePasswordField).
				Title("Password").
				EchoMode(huh.EchoModePassword),
		),
	).WithShowErrors(true)
	return &LoginForm{
		formData:     formData,
		form:         form,
		width:        0,
		height:       0,
		authProvider: authProvider,
	}
}
func (m *LoginForm) Init() tea.Cmd {
	return m.form.Init()
}

func (m *LoginForm) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typedMsg := msg.(type) {
	case tea.KeyMsg:
		switch typedMsg.Type {
		case tea.KeyEsc:
			return m, tui.ChangeStateCmd(tui.StatePublicMenu)
		case tea.KeyCtrlQ, tea.KeyCtrlC:
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = typedMsg.Width
		var formWith = MinTerminalWidth
		if typedMsg.Width > MinTerminalWidth {
			formWith = typedMsg.Width / 2 //nolint:mnd
		}
		m.form = m.form.WithWidth(formWith)
	case tui.ErrorMessageMsg:
		m.errorMessage = typedMsg.Message
		return m, nil
	}
	var cmds []tea.Cmd

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		cmds = append(cmds, cmd)
	}

	if m.form.State == huh.StateCompleted {
		ctx, cancel := context.WithTimeout(context.Background(), DefaultRequestTimeout)
		defer cancel()
		err := spinner.New().Context(ctx).ActionWithErr(func(ctx context.Context) error {
			cmd, err := m.submit(ctx)
			if err != nil {
				return err
			}
			cmds = append(cmds, cmd)
			return nil
		}).Run()

		if err != nil {
			return m.resetForm(err.Error())
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *LoginForm) resetForm(message string) (*LoginForm, tea.Cmd) {
	m.form.State = huh.StateNormal
	newForm := NewLoginForm(m.authProvider)
	newForm.errorMessage = message
	return newForm, newForm.Init()
}

func (m *LoginForm) submit(ctx context.Context) (tea.Cmd, error) {
	user, err := m.authProvider.Login(ctx, m.formData.Username, m.formData.Password)
	if err != nil {
		return nil, fmt.Errorf("login: %w", err)
	}
	return tui.AuthenticateUserCmd(user), nil
}

func (m *LoginForm) View() string {
	b := new(strings.Builder)
	b.WriteString(tui.Logo)
	if m.errorMessage != "" {
		b.WriteString("> " + m.errorMessage + "\n\n")
	}
	formView := m.form.View()
	b.WriteString(formView)

	output := lipgloss.JoinVertical(lipgloss.Center, b.String())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, output)
}
