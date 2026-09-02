package components

import "github.com/charmbracelet/lipgloss"

// BadgeVariant defines styling presets for status indicators.
type BadgeVariant int

const (
	BadgeDefault BadgeVariant = iota
	BadgeSuccess
	BadgeWarning
	BadgeError
)

// RenderBadge outputs an isolated formatted status tag.
func RenderBadge(label string, variant BadgeVariant) string {
	base := lipgloss.NewStyle().Bold(true).Padding(0, 1)

	switch variant {
	case BadgeSuccess:
		return base.Background(lipgloss.Color("#2E7D32")).Foreground(lipgloss.Color("#FFFFFF")).Render(label)
	case BadgeWarning:
		return base.Background(lipgloss.Color("#ED6C02")).Foreground(lipgloss.Color("#FFFFFF")).Render(label)
	case BadgeError:
		return base.Background(lipgloss.Color("#D32F2F")).Foreground(lipgloss.Color("#FFFFFF")).Render(label)
	default:
		return base.Background(lipgloss.Color("#424242")).Foreground(lipgloss.Color("#E0E0E0")).Render(label)
	}
}
