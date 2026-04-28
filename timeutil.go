package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func progressBar(remaining, total time.Duration, width int) string {
	if total <= 0 {
		total = time.Second
	}
	elapsed := total - remaining
	if elapsed < 0 {
		elapsed = 0
	}
	if elapsed > total {
		elapsed = total
	}

	filled := int(float64(elapsed) / float64(total) * float64(width))
	if filled > width {
		filled = width
	}

	bar := strings.Repeat("━", filled) + strings.Repeat("─", width-filled)
	return lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Render(bar)
}

func formatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	total := int(d.Round(time.Second).Seconds())
	minutes := total / 60
	seconds := total % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}
