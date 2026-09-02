package ui

import "github.com/charmbracelet/lipgloss"

// Theme defines the color tokens used across all UI components.
type Theme struct {
	Primary   lipgloss.Color
	Secondary lipgloss.Color
	Accent    lipgloss.Color
	Muted     lipgloss.Color
	Success   lipgloss.Color
	Warning   lipgloss.Color
	Error     lipgloss.Color
	Critical  lipgloss.Color
	Surface   lipgloss.Color
	Border    lipgloss.Color
}

// DefaultTheme provides standard accessible colors.
var DefaultTheme = Theme{
	Primary:   lipgloss.Color("#af59ff"),
	Secondary: lipgloss.Color("#7f56e0"),
	Accent:    lipgloss.Color("#13ad75"),
	Muted:     lipgloss.Color("#6b6b6b"),
	Success:   lipgloss.Color("#62ca42"),
	Warning:   lipgloss.Color("#ff9100"),
	Error:     lipgloss.Color("#db1c1c"),
	Critical:  lipgloss.Color("#b30220"),
	Surface:   lipgloss.Color("#1E1E2E"),
	Border:    lipgloss.Color("#4e4961"),
}
