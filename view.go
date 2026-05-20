package main

import (
	"fmt"
	"strings"
	"time"

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
	contentLineStyle = lipgloss.NewStyle().Background(lipgloss.Color(cuimhneBg1))
	titleStyle       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(cuimhneGreen)).Background(lipgloss.Color(cuimhneBg1))
	mutedStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color(cuimhneFg2)).Background(lipgloss.Color(cuimhneBg1))
	timeStyle        = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(cuimhneMist)).Background(lipgloss.Color(cuimhneBg1))
	warnStyle        = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(cuimhneGold)).Background(lipgloss.Color(cuimhneBg1))
	badStyle         = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(cuimhneTerra)).Background(lipgloss.Color(cuimhneBg1))

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
	buttonStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(cuimhneBg0)).
			Background(lipgloss.Color(cuimhneGreen)).
			Padding(0, 1)
	selectedLineStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(cuimhneBg0)).
				Background(lipgloss.Color(cuimhneGold)).
				Padding(0, 1)
	dangerButtonStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(cuimhneBg0)).
				Background(lipgloss.Color(cuimhneTerra)).
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
		return m.menuView()
	}

	var b strings.Builder
	b.WriteString(center(titleStyle.Render("NO PEEK")))
	b.WriteString("\n")
	b.WriteString(center(mutedStyle.Render(m.problem)))
	b.WriteString("\n\n")

	switch m.phase {
	case phaseFocus:
		b.WriteString(m.timerView("Focus round", m.focusDuration, "Stay with the problem. No hints yet."))
	case phaseDeepFocusOne:
		label := "Deep focus 1/2"
		if m.noBreaks {
			label = "Deep focus"
		}
		b.WriteString(m.timerView(label, m.deepFocusDuration, "No distractions."))
	case phaseDeepShortBreak:
		b.WriteString(m.timerView("Short break", m.shortBreakDuration, "Rest. Don't distract yourself."))
	case phaseDeepFocusTwo:
		b.WriteString(m.timerView("Deep focus 2/2", m.deepFocusDuration, "No distractions."))
	case phaseDeepLongBreak:
		b.WriteString(m.timerView("Long break", m.longBreakDuration, "Rest. Don't distract yourself."))
	case phaseCheckIn:
		b.WriteString(m.checkInView(
			"Time's up.",
			"Are you still generating new ideas?",
			"still thinking",
			fmt.Sprintf("another %s", formatDuration(m.focusDuration)),
			"stuck",
			fmt.Sprintf("%s rescue", formatDuration(m.rescueDuration)),
		))
	case phaseRescue:
		b.WriteString(m.timerView("Rescue round", m.rescueDuration, "Try examples, invariants, brute force, or a smaller case."))
	case phaseFinalCheckIn:
		b.WriteString(m.checkInView(
			"Rescue time is up.",
			"Did you find a new thread to pull on?",
			"yes",
			fmt.Sprintf("continue %s", formatDuration(m.focusDuration)),
			"no",
			"read editorial",
		))
	case phaseEditorial:
		b.WriteString(editorialView())
	}

	return b.String()
}

func (m model) menuView() string {
	var b strings.Builder
	b.WriteString(center(titleStyle.Render("NO PEEK")))
	b.WriteString("\n")
	b.WriteString(center(mutedStyle.Render("launch menu")))
	b.WriteString("\n\n")
	b.WriteString(m.menuRow("mode", "Mode", string(m.mode)))
	b.WriteString("\n")
	b.WriteString(m.menuRow("session", "Session", m.problem))
	b.WriteString("\n\n")
	if m.mode == modePuzzle {
		b.WriteString(m.menuRow("focus", "Focus", m.menuDurationValue("focus", m.focusDuration)))
		b.WriteString("\n")
		b.WriteString(m.menuRow("rescue", "Rescue", m.menuDurationValue("rescue", m.rescueDuration)))
	} else {
		b.WriteString(m.menuRow("deepFocus", "Focus", m.menuDurationValue("deepFocus", m.deepFocusDuration)))
		b.WriteString("\n")
		b.WriteString(m.menuRow("shortBreak", "Short break", m.menuDurationValue("shortBreak", m.shortBreakDuration)))
		b.WriteString("\n")
		b.WriteString(m.menuRow("longBreak", "Long break", m.menuDurationValue("longBreak", m.longBreakDuration)))
		b.WriteString("\n")
		breaks := "on"
		if m.noBreaks {
			breaks = "off"
		}
		b.WriteString(m.menuRow("breaks", "Breaks", breaks))
		b.WriteString("\n")
		b.WriteString(m.menuRow("cycles", "Cycles", fmt.Sprintf("%d", m.deepCycles)))
	}
	b.WriteString("\n\n")
	if m.menuEditingField == "" {
		b.WriteString(footer("j/k select · enter edit · s start · q quit"))
	} else if m.isMinuteField(m.menuEditingField) {
		b.WriteString(footer("type minutes, e.g. 25 · enter save · esc cancel"))
	} else {
		b.WriteString(footer("type value or h/l toggle · enter done · esc cancel"))
	}
	return b.String()
}

func (m model) menuDurationValue(field string, d time.Duration) string {
	if m.menuEditingField == field {
		if m.menuInput == "" {
			return ""
		}
		return m.menuInput + " min"
	}
	return fmt.Sprintf("%d min", int(d/time.Minute))
}

func (m model) menuRow(field, label, value string) string {
	prefix := "  "
	if m.selectedMenuField() == field {
		prefix = "> "
	}
	line := fmt.Sprintf("%s%-12s %s", prefix, label+":", value)
	if m.menuEditingField == field {
		line = selectedLineStyle.Width(innerWidth()).Render(line)
	} else {
		line = contentLineStyle.Width(innerWidth()).Render(line)
	}
	return center(line)
}

func (m model) timerView(label string, total time.Duration, subtitle string) string {
	var b strings.Builder
	b.WriteString(center(mutedStyle.Render(label)))
	b.WriteString("\n")
	b.WriteString(center(timeStyle.Render(formatDuration(m.remaining))))
	b.WriteString("\n")
	b.WriteString("\n")
	if m.paused {
		b.WriteString(center(warnStyle.Render("paused")))
		b.WriteString("\n")
	} else {
		b.WriteString(center(mutedStyle.Render(subtitle)))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(footer("p/space pause · q menu"))
	return b.String()
}

func (m model) checkInView(title, prompt, continueLabel, continueHint, stuckLabel, stuckHint string) string {
	var b strings.Builder
	b.WriteString(center(warnStyle.Render(title)))
	b.WriteString("\n\n")
	b.WriteString(center(prompt))
	b.WriteString("\n\n")
	b.WriteString(choice("t", continueLabel, continueHint))
	b.WriteString("\n")
	b.WriteString(choiceDanger("s", stuckLabel, stuckHint))
	b.WriteString("\n\n")
	b.WriteString(footer("q menu"))
	return b.String()
}

func editorialView() string {
	var b strings.Builder
	b.WriteString(center(badStyle.Render("READ THE EDITORIAL")))
	b.WriteString("\n\n")
	msg := "You gave the problem a real attempt. Learn the missing idea, then try to re-solve it without looking."
	b.WriteString(lipgloss.NewStyle().Width(innerWidth()).Align(lipgloss.Center).Render(msg))
	b.WriteString("\n\n")
	b.WriteString(footer("enter menu · r restart · q menu"))
	return b.String()
}

func choice(key, label, hint string) string {
	line := fmt.Sprintf("%s  %-16s %s", buttonStyle.Render(key), label, mutedStyle.Render(hint))
	return center(line)
}

func choiceDanger(key, label, hint string) string {
	line := fmt.Sprintf("%s  %-16s %s", dangerButtonStyle.Render(key), label, mutedStyle.Render(hint))
	return center(line)
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
