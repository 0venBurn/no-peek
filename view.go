package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	contentWidth      = 58
	horizontalPadding = 3
)

const (
	cuimhneBg0   = "#1A1714"
	cuimhneBg1   = "#242019"
	cuimhneBg3   = "#332E27"
	cuimhneFg0   = "#F0EBE1"
	cuimhneFg2   = "#9C9488"
	cuimhneGreen = "#7FA688"
	cuimhneTerra = "#C47A5A"
	cuimhneGold  = "#C4A882"
	cuimhneMist  = "#8FA89C"
)

var (
	contentLineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(cuimhneFg0)).Background(lipgloss.Color(cuimhneBg1))
	titleStyle       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(cuimhneGreen)).Background(lipgloss.Color(cuimhneBg1))
	mutedStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color(cuimhneFg2)).Background(lipgloss.Color(cuimhneBg1))
	timeStyle        = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(cuimhneMist)).Background(lipgloss.Color(cuimhneBg1))
	warnStyle        = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(cuimhneGold)).Background(lipgloss.Color(cuimhneBg1))

	appStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(cuimhneFg0)).
			Background(lipgloss.Color(cuimhneBg0))
	boxStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(cuimhneFg0)).
			Background(lipgloss.Color(cuimhneBg1)).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(cuimhneBg3)).
			Padding(1, horizontalPadding).
			Width(contentWidth)
	selectedLineStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(cuimhneBg0)).
				Background(lipgloss.Color(cuimhneGold)).
				Padding(0, 1)
)

func (m model) View() string {
	box := boxStyle.Render(m.content())

	if m.width <= 0 || m.height <= 0 {
		return appStyle.Render(box)
	}

	return appStyle.
		Width(m.width).
		Height(m.height).
		Render(lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			box,
			lipgloss.WithWhitespaceBackground(lipgloss.Color(cuimhneBg0)),
		))
}

func (m model) content() string {
	if m.screen == screenMenu {
		return menuView(m.menu.render())
	}
	return sessionView(m.session.render())
}

func sessionView(render sessionRender) string {
	var b strings.Builder
	b.WriteString(center(titleStyle.Render("NO PEEK")))
	b.WriteString("\n")
	b.WriteString(center(mutedStyle.Render(render.problem)))
	b.WriteString("\n\n")

	switch render.kind {
	case sessionRenderTimer:
		b.WriteString(timerView(render.timer))
	case sessionRenderCheckIn:
		b.WriteString(checkInView(render.checkIn))
	case sessionRenderEditorial:
		b.WriteString(editorialView(render.editorial))
	}

	return b.String()
}

func menuView(render menuRender) string {
	var b strings.Builder
	b.WriteString(center(titleStyle.Render("NO PEEK")))
	b.WriteString("\n")
	b.WriteString(center(mutedStyle.Render("launch menu")))
	b.WriteString("\n\n")

	for i, row := range render.rows {
		if i == 2 {
			b.WriteString("\n\n")
		} else if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(menuRow(row))
	}

	b.WriteString("\n\n")
	b.WriteString(footer(render.footer))
	return b.String()
}

func menuRow(row menuRowRender) string {
	prefix := "  "
	if row.selected {
		prefix = "> "
	}
	line := fmt.Sprintf("%s%-12s %s", prefix, row.label+":", row.value)
	if row.editing {
		line = selectedLineStyle.Width(innerWidth()).Render(line)
	} else {
		line = contentLineStyle.Width(innerWidth()).Render(line)
	}
	return center(line)
}

func timerView(timer timerRender) string {
	var b strings.Builder
	b.WriteString(center(mutedStyle.Render(timer.label)))
	b.WriteString("\n")
	b.WriteString(center(timeStyle.Render(formatDuration(timer.remaining))))
	b.WriteString("\n")
	b.WriteString("\n")
	if timer.paused {
		b.WriteString(center(warnStyle.Render("paused")))
		b.WriteString("\n")
	} else {
		b.WriteString(center(mutedStyle.Render(timer.subtitle)))
		b.WriteString("\n")
	}
	if timer.canSolve {
		b.WriteString("\n")
		b.WriteString(footer("c solved · p/space pause · q menu"))
	} else {
		b.WriteString("\n")
		b.WriteString(footer("p/space pause · q menu"))
	}
	return b.String()
}

func checkInView(checkIn checkInRender) string {
	var b strings.Builder
	b.WriteString(center(titleStyle.Render(checkIn.title)))
	b.WriteString("\n\n")
	b.WriteString(center(mutedStyle.Render(checkIn.prompt)))
	b.WriteString("\n\n")
	b.WriteString(choice("t", checkIn.continueLabel, checkIn.continueHint))
	b.WriteString("\n")
	b.WriteString(choice("s", checkIn.stuckLabel, checkIn.stuckHint))
	b.WriteString("\n\n")
	b.WriteString(footer("q menu"))
	return b.String()
}

func editorialView(editorial editorialRender) string {
	var b strings.Builder
	b.WriteString(center(titleStyle.Render(editorial.title)))
	b.WriteString("\n\n")
	b.WriteString(contentLineStyle.Width(innerWidth()).Align(lipgloss.Center).Render(editorial.message))
	b.WriteString("\n\n")
	b.WriteString(footer("enter menu · r restart · q menu"))
	return b.String()
}

func choice(key, label, hint string) string {
	line := fmt.Sprintf("%s   %-16s %s", key, label, hint)
	return center(mutedStyle.Render(line))
}

func footer(text string) string {
	return center(mutedStyle.Render(text))
}

func center(s string) string {
	return contentLineStyle.Render(lipgloss.PlaceHorizontal(
		innerWidth(),
		lipgloss.Center,
		s,
		lipgloss.WithWhitespaceBackground(lipgloss.Color(cuimhneBg1)),
	))
}

func innerWidth() int {
	return contentWidth - horizontalPadding*2
}
