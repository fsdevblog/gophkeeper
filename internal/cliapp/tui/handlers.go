package tui

import (
	"context"
	"errors"
	"net/http"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsdevblog/gophkeeper/internal/cliapp/api"
)

func (m *Model) handleLogin(ctx context.Context, userInput string) tea.Cmd {
	var cmd tea.Cmd
	switch userInput {
	case shiftTab:
		cmd = m.loginForm.DecrementFocus()
	case tab:
		cmd = m.loginForm.IncrementFocus()
	case esc:
		m.currentState = AppStateMain
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
					return nil
				}
				m.alertBanner = err.Error()
			}
		}
	}
	return cmd
}

func (m *Model) handleRegister(ctx context.Context, userInput string) tea.Cmd {
	var cmd tea.Cmd
	switch userInput {
	case shiftTab, up:
		cmd = m.registerForm.DecrementFocus()
	case tab, down:
		cmd = m.registerForm.IncrementFocus()
	case esc:
		m.currentState = AppStateMain
	case enter:
		result, err := m.registerForm.Submit(ctx)
		if err != nil {
			m.alertBanner = err.Error()
			return nil
		}

		user := User{
			Username: result.Username,
			ID:       result.ID,
		}
		if errAuth := m.authenticateUser(user, result.Token); errAuth != nil {
			m.alertBanner = errAuth.Error()
			return nil
		}

		m.currentState = AppStateAuthorizedMain
	}
	return cmd
}

func (m *Model) handleMain(userInput string) {
	switch userInput {
	case up:
		m.mainMenu.Prev()
	case down:
		m.mainMenu.Next()
	case enter:
		switch m.mainMenu.GetSelected() {
		case MMChoiceLogin:
			m.currentState = AppStateLogin
		case MMChoiceRegister:
			m.currentState = AppStateRegister
		}
	}
}
