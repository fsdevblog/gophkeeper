package tui

import (
	"context"
	tea "github.com/charmbracelet/bubbletea"
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
			resp, err := m.loginForm.Submit(ctx)
			if err != nil {
				m.alertBanner = err.Error()
				return nil
			}

			errAuth := m.authenticateUser(&User{
				Username: resp.Username,
				ID:       resp.ID,
			}, resp.Token)

			if errAuth != nil {
				m.alertBanner = errAuth.Error()
				return nil
			}
			m.currentState = AppStateAuthorizedMain
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
		if errAuth := m.authenticateUser(&user, result.Token); errAuth != nil {
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
