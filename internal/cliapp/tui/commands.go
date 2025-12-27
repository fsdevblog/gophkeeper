package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsdevblog/gophkeeper/internal/cliapp/tui/models"
)

type StateMsg struct {
	NewState State
}

func ChangeStateCmd(newState State) tea.Cmd {
	return func() tea.Msg {
		return StateMsg{
			NewState: newState,
		}
	}
}

type ErrorMessageMsg struct {
	Message string
}

func ErrorMessageCmd(message string) tea.Cmd {
	return func() tea.Msg {
		return ErrorMessageMsg{
			Message: message,
		}
	}
}


type AuthenticateUserMsg struct {
	User *models.User
}

func AuthenticateUserCmd(user *models.User) tea.Cmd {
	return func() tea.Msg {
		return AuthenticateUserMsg{
			User: user,
		}
	}
}
