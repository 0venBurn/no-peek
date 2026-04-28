package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

const contentWidth = 58

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	mutedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	timeStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	warnStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220"))
	badStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("203"))

	appStyle = lipgloss.NewStyle().Padding(1, 2)
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 3).
			Width(contentWidth)
	buttonStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("62")).
			Padding(0, 1)
	dangerButtonStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("230")).
				Background(lipgloss.Color("203")).
				Padding(0, 1)
)

func (m model) View() string {
	box := boxStyle.Render(m.content())

	if m.width > 0 {
		box = lipgloss.PlaceHorizontal(m.width, lipgloss.Center, box)
	}
	if m.height > 0 {
		box = lipgloss.PlaceVertical(m.height, lipgloss.Center, box)
	}
	return appStyle.Render(box)
}

func (m model) content() string {
	var b strings.Builder
	b.WriteString(center(titleStyle.Render("NO PEEK")))
	b.WriteString("\n")
	b.WriteString(center(mutedStyle.Render(m.problem)))
	b.WriteString("\n\n")

	switch m.phase {
	case phaseFocus:
		b.WriteString(m.timerView("Focus round", m.focusDuration, "Stay with the problem. No hints yet."))
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

func (m model) timerView(label string, total time.Duration, subtitle string) string {
	var b strings.Builder
	b.WriteString(center(mutedStyle.Render(label)))
	b.WriteString("\n")
	b.WriteString(center(timeStyle.Render(formatDuration(m.remaining))))
	b.WriteString("\n")
	b.WriteString(center(progressBar(m.remaining, total, 36)))
	b.WriteString("\n\n")
	if m.paused {
		b.WriteString(center(warnStyle.Render("paused")))
		b.WriteString("\n")
	} else {
		b.WriteString(center(mutedStyle.Render(subtitle)))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(footer("p/space pause · q quit"))
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
	b.WriteString(footer("q quit"))
	return b.String()
}

func editorialView() string {
	var b strings.Builder
	b.WriteString(center(badStyle.Render("READ THE EDITORIAL")))
	b.WriteString("\n\n")
	msg := "You gave the problem a real attempt. Learn the missing idea, then try to re-solve it without looking."
	b.WriteString(lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center).Render(msg))
	b.WriteString("\n\n")
	b.WriteString(footer("r restart · q quit"))
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
	return lipgloss.PlaceHorizontal(contentWidth, lipgloss.Center, s)
}
