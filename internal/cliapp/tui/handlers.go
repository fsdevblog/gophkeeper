package tui

import (
	"context"
	"errors"
	"net/http"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsdevblog/gophkeeper/internal/cliapp/api"
)

func (m *Model) handleLogin(ctx context.Context, userInput string) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch userInput {
	case shiftTab:
		cmd = m.loginForm.DecrementFocus()
	case tab:
		cmd = m.loginForm.IncrementFocus()
	case esc:
		m.CurrentState = AppStateMain
	case enter:
		if m.loginForm.IsValid() {
			_, err := m.api.Login(ctx, api.LoginParams{
				Username: m.loginForm.Username.Value(),
				Password: m.loginForm.Password.Value(),
			})
			if err != nil {
				var errResponse *api.UnexpectedHTTPStatusCodeError
				if errors.As(err, &errResponse) && errResponse.StatusCode == http.StatusUnauthorized {
					m.alertBanner = "wrong username or password"
					return m, nil
				}
				m.alertBanner = err.Error()
			}
		}
	}
	return m, cmd
}

func (m *Model) handleRegister(ctx context.Context, userInput string) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch userInput {
	case shiftTab:
		cmd = m.registerForm.DecrementFocus()
	case tab:
		cmd = m.registerForm.IncrementFocus()
	case esc:
		m.CurrentState = AppStateMain
	case enter:
		if err := m.registerForm.Submit(ctx); err != nil {
			m.alertBanner = err.Error()
			return m, nil
		}
	}
	return m, cmd
}

func (m *Model) handleMain(userInput string) (tea.Model, tea.Cmd) {
	switch userInput {
	case up:
		m.mainMenu.Prev()
	case down:
		m.mainMenu.Next()
	case enter:
		switch m.mainMenu.GetSelected() {
		case MMChoiceLogin:
			m.CurrentState = AppStateLogin
		case MMChoiceRegister:
			m.CurrentState = AppStateRegister
		}
	}
	return m, nil
}
