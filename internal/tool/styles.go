package tool

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

var (
	// colors
	//
	// palette light
	// #e63946, #f1faee, #a8dadc, #457b9d, #1d3557
	// https://coolors.co/palette/e63946-f1faee-a8dadc-457b9d-1d3557
	//
	// palette dark
	// #f72585, #7209b7, #3a0ca3, #4361ee, #4cc9f0
	// https://coolors.co/palette/f72585-7209b7-3a0ca3-4361ee-4cc9f0
	colorHighlight = lipgloss.AdaptiveColor{Light: "#1D3557", Dark: "#e0aaff"}
	colorSubtle    = lipgloss.AdaptiveColor{Light: "#A8DADC", Dark: "#495057"}
	colorSpecial   = lipgloss.AdaptiveColor{Light: "#E63946", Dark: "#F72585"}
	colorBright    = lipgloss.AdaptiveColor{Light: "#F1FAEE", Dark: "#4CC9F0"}

	styleTitle = lipgloss.NewStyle().
			MarginLeft(5).MarginRight(5).
			Padding(0, 1, 0, 1).
			Bold(true).
			Foreground(colorBright).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorSubtle).
			BorderTop(true).
			BorderLeft(true).
			BorderRight(true).
			BorderBottom(true)

	styleListItem     = lipgloss.NewStyle().PaddingLeft(4)
	styleListSelected = lipgloss.NewStyle().PaddingLeft(2).Foreground(colorSpecial)
	styleListTitle    = lipgloss.NewStyle().Padding(1, 0, 0, 2).Bold(true)

	paginationStyle = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle       = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	//quitTextStyle   = lipgloss.NewStyle().Margin(1, 0, 2, 4)

	//errStyle      = lipgloss.NewStyle().PaddingLeft(4).PaddingBottom(1).Bold(true).Border(lipgloss.RoundedBorder())
	quitTextStyle = lipgloss.NewStyle().Margin(1, 0, 2, 4).Bold(true).Faint(true)

	defaultHuhTheme = huh.ThemeDracula()
	errStyle        = defaultHuhTheme.Focused.ErrorIndicator.
			Padding(0, 0, 1, 4).
			Bold(true).
			Border(lipgloss.RoundedBorder())
)
