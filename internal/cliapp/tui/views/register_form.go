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

var _ tea.Model = &RegisterForm{}

type RegisterFormData struct {
	Username             string
	Password             string
	PasswordConfirmation string
}
type RegisterForm struct {
	form         *huh.Form
	authProvider AuthProvider
	formData     *RegisterFormData
	width        int
	height       int
	errorMessage string
}

func NewRegisterForm(authProvider AuthProvider) *RegisterForm {
	formData := new(RegisterFormData)
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
			huh.NewInput().
				Value(&formData.PasswordConfirmation).
				Validate(func(s string) error {
					if s != formData.Password {
						return fmt.Errorf("passwords do not match")
					}
					return nil
				}).
				Title("Confirm password").
				EchoMode(huh.EchoModePassword),
		),
	).WithShowErrors(true)

	return &RegisterForm{
		form:         form,
		authProvider: authProvider,
		formData:     formData,
		errorMessage: "",
	}
}

func (r *RegisterForm) Init() tea.Cmd {
	return r.form.Init()
}

func (r *RegisterForm) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typedMsg := msg.(type) {
	case tea.KeyMsg:
		switch typedMsg.Type {
		case tea.KeyEsc:
			return r, tui.ChangeStateCmd(tui.StatePublicMenu)
		case tea.KeyCtrlQ, tea.KeyCtrlC:
			return r, tea.Quit
		}
	case tea.WindowSizeMsg:
		r.width = typedMsg.Width
		var formWith = MinTerminalWidth
		if typedMsg.Width > MinTerminalWidth {
			formWith = typedMsg.Width / 2 //nolint:mnd
		}
		r.form = r.form.WithWidth(formWith)
	case tui.ErrorMessageMsg:
		r.errorMessage = typedMsg.Message
		return r, nil
	}
	var cmds []tea.Cmd

	form, cmd := r.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		r.form = f
		cmds = append(cmds, cmd)
	}

	if r.form.State == huh.StateCompleted {
		ctx, cancel := context.WithTimeout(context.Background(), DefaultRequestTimeout)
		defer cancel()
		err := spinner.New().Context(ctx).ActionWithErr(func(ctx context.Context) error {
			user, err := r.authProvider.Register(ctx, r.formData.Username, r.formData.Password)
			if err != nil {
				return err
			}
			cmds = append(cmds, tui.AuthenticateUserCmd(user))
			return nil
		}).Run()

		if err != nil {
			return r.resetForm(err.Error())
		}
	}
	return r, tea.Batch(cmds...)
}

func (r *RegisterForm) View() string {
	b := new(strings.Builder)
	b.WriteString(tui.Logo)
	if r.errorMessage != "" {
		b.WriteString("> " + r.errorMessage + "\n\n")
	}
	b.WriteString(r.form.View())

	output := lipgloss.JoinVertical(lipgloss.Center, b.String())
	return lipgloss.Place(r.width, r.height, lipgloss.Center, lipgloss.Center, output)
}

func (r *RegisterForm) resetForm(message string) (*RegisterForm, tea.Cmd) {
	r.form.State = huh.StateNormal
	newForm := NewRegisterForm(r.authProvider)
	newForm.errorMessage = message
	return newForm, newForm.Init()
}
