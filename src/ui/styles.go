package ui

import "github.com/charmbracelet/lipgloss"

// Stylesheet encapsulates reusable styles built from a theme configuration.
type Stylesheet struct {
	Card       lipgloss.Style
	Title      lipgloss.Style
	Subtext    lipgloss.Style
	StatusBase lipgloss.Style
}

// NewStylesheet constructs a Stylesheet using the supplied Theme.
func NewStylesheet(t Theme) Stylesheet {
	return Stylesheet{
		Card: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(t.Border).
			Padding(1, 2).
			Background(t.Surface),
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(t.Primary),
		Subtext: lipgloss.NewStyle().
			Foreground(t.Muted),
		StatusBase: lipgloss.NewStyle().
			Bold(true).
			Padding(0, 1),
	}
}
