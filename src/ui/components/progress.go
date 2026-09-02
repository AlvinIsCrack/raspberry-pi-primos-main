package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ProgressBarConfig define los parámetros de renderizado de la barra de progreso.
type ProgressBarConfig struct {
	Percent     int
	TotalWidth  int
	FgColor     lipgloss.Color
	EmptyColor  lipgloss.Color
	LabelFormat string
	FilledChar  string // Opcional
	EmptyChar   string // Opcional
}

// formatProgressLabel reemplaza los placeholders '%p' por el porcentaje numérico.
func formatProgressLabel(template string, percent int) string {
	if template == "" {
		return ""
	}
	const escapedToken = "\x00PERCENT\x00"
	s := strings.ReplaceAll(template, "%%", escapedToken)
	s = strings.ReplaceAll(s, "%p", fmt.Sprintf("%d%%", percent))
	s = strings.ReplaceAll(s, escapedToken, "%")
	return s
}

// RenderProgressBar dibuja una barra de progreso basada en caracteres según la configuración provista.
func RenderProgressBar(cfg ProgressBarConfig) string {
	percent := min(100, max(0, cfg.Percent))
	if cfg.TotalWidth <= 0 {
		return ""
	}

	filledChar := cfg.FilledChar
	if filledChar == "" {
		filledChar = "#"
	}

	emptyChar := cfg.EmptyChar
	if emptyChar == "" {
		emptyChar = " "
	}

	filledLen := (percent * cfg.TotalWidth) / 100
	emptyLen := cfg.TotalWidth - filledLen

	filled := lipgloss.NewStyle().Foreground(cfg.FgColor).Render(strings.Repeat(filledChar, filledLen))
	empty := lipgloss.NewStyle().Foreground(cfg.EmptyColor).Render(strings.Repeat(emptyChar, emptyLen))
	bar := fmt.Sprintf("%s%s", filled, empty)

	label := formatProgressLabel(cfg.LabelFormat, percent)
	if label == "" {
		return bar
	}

	renderedLabel := lipgloss.NewStyle().
		Width(cfg.TotalWidth).
		Align(lipgloss.Center).
		Foreground(cfg.EmptyColor).
		Render(label)

	return lipgloss.JoinVertical(lipgloss.Center, bar, renderedLabel)
}
