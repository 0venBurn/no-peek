package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type phase int

const (
	phaseFocus phase = iota
	phaseCheckIn
	phaseRescue
	phaseFinalCheckIn
	phaseEditorial
)

type tickMsg time.Time

type model struct {
	problem        string
	phase          phase
	remaining      time.Duration
	focusDuration  time.Duration
	rescueDuration time.Duration
	paused         bool
	width          int
	height         int
}

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
			Width(58)
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

func main() {
	focusMinutes := flag.Int("focus", 30, "focus round length in minutes")
	rescueMinutes := flag.Int("rescue", 15, "stuck/rescue round length in minutes")
	flag.Parse()

	problem := strings.TrimSpace(strings.Join(flag.Args(), " "))
	if problem == "" {
		problem = "Untitled problem"
	}

	m := model{
		problem:        problem,
		phase:          phaseFocus,
		focusDuration:  time.Duration(*focusMinutes) * time.Minute,
		rescueDuration: time.Duration(*rescueMinutes) * time.Minute,
	}
	m.remaining = m.focusDuration

	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func (m model) Init() tea.Cmd {
	return tick()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "p", " ":
			if m.isTimerPhase() {
				m.paused = !m.paused
			}
		case "t":
			if m.phase == phaseCheckIn || m.phase == phaseFinalCheckIn {
				m.phase = phaseFocus
				m.remaining = m.focusDuration
				m.paused = false
				return m, tick()
			}
		case "s":
			if m.phase == phaseCheckIn {
				m.phase = phaseRescue
				m.remaining = m.rescueDuration
				m.paused = false
				return m, tick()
			}
			if m.phase == phaseFinalCheckIn {
				m.phase = phaseEditorial
				m.paused = false
				return m, notify("No Peek", "Time to read the editorial.")
			}
		case "r":
			if m.phase == phaseEditorial {
				m.phase = phaseFocus
				m.remaining = m.focusDuration
				return m, tick()
			}
		}

	case tickMsg:
		if !m.isTimerPhase() {
			return m, nil
		}
		if m.paused {
			return m, tick()
		}

		m.remaining -= time.Second
		if m.remaining <= 0 {
			m.remaining = 0
			if m.phase == phaseFocus {
				m.phase = phaseCheckIn
				return m, notify("No Peek", "30 minutes are up. Are you still thinking or stuck?")
			}
			if m.phase == phaseRescue {
				m.phase = phaseFinalCheckIn
				return m, notify("No Peek", "15 minute rescue is up. Still stuck?")
			}
		}
		return m, tick()
	}

	return m, nil
}

func (m model) View() string {
	content := m.content()
	box := boxStyle.Render(content)

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
	b.WriteString(lipgloss.PlaceHorizontal(58, lipgloss.Center, titleStyle.Render("NO PEEK")))
	b.WriteString("\n")
	b.WriteString(lipgloss.PlaceHorizontal(58, lipgloss.Center, mutedStyle.Render(m.problem)))
	b.WriteString("\n\n")

	switch m.phase {
	case phaseFocus:
		b.WriteString(m.timerView("Focus round", m.focusDuration, "Stay with the problem. No hints yet."))
	case phaseCheckIn:
		b.WriteString(lipgloss.PlaceHorizontal(58, lipgloss.Center, warnStyle.Render("Time's up.")))
		b.WriteString("\n\n")
		b.WriteString(lipgloss.PlaceHorizontal(58, lipgloss.Center, "Are you still generating new ideas?"))
		b.WriteString("\n\n")
		b.WriteString(choice("t", "still thinking", fmt.Sprintf("another %s", formatDuration(m.focusDuration))))
		b.WriteString("\n")
		b.WriteString(choiceDanger("s", "stuck", fmt.Sprintf("%s rescue", formatDuration(m.rescueDuration))))
		b.WriteString("\n\n")
		b.WriteString(footer("q quit"))
	case phaseRescue:
		b.WriteString(m.timerView("Rescue round", m.rescueDuration, "Try examples, invariants, brute force, or a smaller case."))
	case phaseFinalCheckIn:
		b.WriteString(lipgloss.PlaceHorizontal(58, lipgloss.Center, warnStyle.Render("Rescue time is up.")))
		b.WriteString("\n\n")
		b.WriteString(lipgloss.PlaceHorizontal(58, lipgloss.Center, "Did you find a new thread to pull on?"))
		b.WriteString("\n\n")
		b.WriteString(choice("t", "yes", fmt.Sprintf("continue %s", formatDuration(m.focusDuration))))
		b.WriteString("\n")
		b.WriteString(choiceDanger("s", "no", "read editorial"))
		b.WriteString("\n\n")
		b.WriteString(footer("q quit"))
	case phaseEditorial:
		b.WriteString(lipgloss.PlaceHorizontal(58, lipgloss.Center, badStyle.Render("READ THE EDITORIAL")))
		b.WriteString("\n\n")
		msg := "You gave the problem a real attempt. Learn the missing idea, then try to re-solve it without looking."
		b.WriteString(lipgloss.NewStyle().Width(58).Align(lipgloss.Center).Render(msg))
		b.WriteString("\n\n")
		b.WriteString(footer("r restart · q quit"))
	}

	return b.String()
}

func (m model) timerView(label string, total time.Duration, subtitle string) string {
	var b strings.Builder
	b.WriteString(lipgloss.PlaceHorizontal(58, lipgloss.Center, mutedStyle.Render(label)))
	b.WriteString("\n")
	b.WriteString(lipgloss.PlaceHorizontal(58, lipgloss.Center, timeStyle.Render(formatDuration(m.remaining))))
	b.WriteString("\n")
	b.WriteString(lipgloss.PlaceHorizontal(58, lipgloss.Center, progressBar(m.remaining, total, 36)))
	b.WriteString("\n\n")
	if m.paused {
		b.WriteString(lipgloss.PlaceHorizontal(58, lipgloss.Center, warnStyle.Render("paused")))
		b.WriteString("\n")
	} else {
		b.WriteString(lipgloss.PlaceHorizontal(58, lipgloss.Center, mutedStyle.Render(subtitle)))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(footer("p/space pause · q quit"))
	return b.String()
}

func choice(key, label, hint string) string {
	line := fmt.Sprintf("%s  %-16s %s", buttonStyle.Render(key), label, mutedStyle.Render(hint))
	return lipgloss.PlaceHorizontal(58, lipgloss.Center, line)
}

func choiceDanger(key, label, hint string) string {
	line := fmt.Sprintf("%s  %-16s %s", dangerButtonStyle.Render(key), label, mutedStyle.Render(hint))
	return lipgloss.PlaceHorizontal(58, lipgloss.Center, line)
}

func footer(text string) string {
	return lipgloss.PlaceHorizontal(58, lipgloss.Center, mutedStyle.Render(text))
}

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

func (m model) isTimerPhase() bool {
	return m.phase == phaseFocus || m.phase == phaseRescue
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
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

func notify(title, message string) tea.Cmd {
	return func() tea.Msg {
		// Always ring the terminal bell. This is dependency-free and works anywhere
		// the terminal has audible/visual bell enabled.
		_, _ = os.Stdout.Write([]byte("\a"))

		// Best-effort native desktop notification, with no Go dependencies.
		// If the command is unavailable, silently fall back to the bell.
		switch runtime.GOOS {
		case "linux":
			if path, err := exec.LookPath("notify-send"); err == nil {
				_ = exec.Command(path, title, message).Run()
			}
		case "darwin":
			script := fmt.Sprintf(`display notification %q with title %q`, message, title)
			_ = exec.Command("osascript", "-e", script).Run()
		case "windows":
			ps := fmt.Sprintf(`[console]::beep(); Add-Type -AssemblyName System.Windows.Forms; [System.Windows.Forms.MessageBox]::Show(%q, %q) | Out-Null`, message, title)
			_ = exec.Command("powershell", "-NoProfile", "-Command", ps).Run()
		}
		return nil
	}
}
