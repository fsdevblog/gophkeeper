package cliapp

import (
	"fmt"
	"github.com/fsdevblog/gophkeeper/internal/cliapp/api"
	"github.com/fsdevblog/gophkeeper/internal/cliapp/tui/auth"
	"github.com/fsdevblog/gophkeeper/internal/cliapp/tui/starter"
	"go.uber.org/zap"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsdevblog/gophkeeper/internal/cliapp/tui"
)

type CLIApp struct {
	baseURL string
}

type InitParams struct {
	BaseURL string
}

func New(p InitParams) *CLIApp {
	return &CLIApp{baseURL: p.BaseURL}
}

func (a *CLIApp) Run() error {
	apiClient := api.MustNew(a.baseURL)
	authProvider := auth.New(apiClient, zap.NewNop())
	state, errState := starter.New(tui.StatePublicMenu, authProvider)

	if errState != nil {
		return fmt.Errorf("run failed: %w", errState)
	}
	p := tea.NewProgram(state, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("cli app run failed: %w", err)
	}
	return nil
}
