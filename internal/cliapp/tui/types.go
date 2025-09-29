package tui

import "github.com/charmbracelet/lipgloss"

//nolint:gochecknoglobals
var (
	primaryTextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("211"))
	mainStyle        = lipgloss.NewStyle().MarginLeft(2) //nolint:mnd
	errStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
	alertBannerStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#FF0000")).Padding(1)
)

type appState int

const (
	AppStateMain appState = iota
	AppStateLogin
	AppStateRegister

	AppStateAuthorizedMain
)

const (
	up       = "up"
	down     = "down"
	enter    = "enter"
	tab      = "tab"
	shiftTab = "shift+tab"
	ctrlC    = "ctrl+c"
	ctrlQ    = "ctrl+q"
	esc      = "esc"
)
