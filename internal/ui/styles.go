package ui

import (
	"charm.land/bubbles/v2/help"
	"charm.land/lipgloss/v2"
)

var (
	primary    = lipgloss.Color("#7AA2F7")
	highlight  = lipgloss.Color("#9ECE6A")
	dim        = lipgloss.Color("#565F89")
	bg         = lipgloss.Color("#1A1B26")
	selectedBg = lipgloss.Color("#292E42")
	accent     = lipgloss.Color("#BB9AF7")
)

var listTitleStyle = lipgloss.NewStyle().
	Background(selectedBg).
	Foreground(accent).
	Padding(0, 1)

var footerStyle = lipgloss.NewStyle().
	Padding(0, 1).
	Border(lipgloss.NormalBorder(), true, false, false).
	BorderForeground(dim)

var searchStyle = lipgloss.NewStyle().
	Foreground(accent).
	Border(lipgloss.RoundedBorder()).
	BorderForeground(accent).
	Padding(0, 1)

var progressStyle = lipgloss.NewStyle().
	Foreground(primary).
	MarginRight(10).
	Background(bg)

var selectedItem = lipgloss.NewStyle().
	Foreground(highlight).
	Bold(true).
	Background(selectedBg).
	Padding(0, 1)

var normalItem = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#C0CAF5"))

var dimItem = lipgloss.NewStyle().
	Foreground(dim)

var helpStyle = help.Styles{
	ShortKey:       lipgloss.NewStyle().Foreground(primary).Bold(true),
	ShortDesc:      lipgloss.NewStyle().Foreground(dim),
	ShortSeparator: lipgloss.NewStyle().Foreground(dim),

	FullKey:       lipgloss.NewStyle().Foreground(primary).Bold(true),
	FullDesc:      lipgloss.NewStyle().Foreground(dim),
	FullSeparator: lipgloss.NewStyle().Foreground(dim),

	Ellipsis: lipgloss.NewStyle().Foreground(dim),
}

var thumbnailStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(dim).
	Padding(1).
	Align(lipgloss.Center, lipgloss.Center).
	Foreground(lipgloss.Color("#7AA2F7"))
