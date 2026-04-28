package main

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
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

func newModel(cfg config) model {
	return model{
		problem:        cfg.problem,
		phase:          phaseFocus,
		remaining:      cfg.focusDuration,
		focusDuration:  cfg.focusDuration,
		rescueDuration: cfg.rescueDuration,
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
		return m.handleKey(msg)

	case tickMsg:
		return m.handleTick()
	}

	return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "p", " ":
		if m.isTimerPhase() {
			m.paused = !m.paused
		}
	case "t":
		if m.phase == phaseCheckIn || m.phase == phaseFinalCheckIn {
			return m.startFocusRound(), tick()
		}
	case "s":
		return m.handleStuck()
	case "r":
		if m.phase == phaseEditorial {
			return m.startFocusRound(), tick()
		}
	}

	return m, nil
}

func (m model) handleStuck() (tea.Model, tea.Cmd) {
	switch m.phase {
	case phaseCheckIn:
		m.phase = phaseRescue
		m.remaining = m.rescueDuration
		m.paused = false
		return m, tick()
	case phaseFinalCheckIn:
		m.phase = phaseEditorial
		m.paused = false
		return m, notify("No Peek", "Time to read the editorial.")
	default:
		return m, nil
	}
}

func (m model) handleTick() (tea.Model, tea.Cmd) {
	if !m.isTimerPhase() {
		return m, nil
	}
	if m.paused {
		return m, tick()
	}

	m.remaining -= time.Second
	if m.remaining > 0 {
		return m, tick()
	}

	m.remaining = 0
	return m.advanceAfterTimerExpires()
}

func (m model) advanceAfterTimerExpires() (tea.Model, tea.Cmd) {
	switch m.phase {
	case phaseFocus:
		m.phase = phaseCheckIn
		return m, notify("No Peek", "30 minutes are up. Are you still thinking or stuck?")
	case phaseRescue:
		m.phase = phaseFinalCheckIn
		return m, notify("No Peek", "15 minute rescue is up. Still stuck?")
	default:
		return m, nil
	}
}

func (m model) startFocusRound() model {
	m.phase = phaseFocus
	m.remaining = m.focusDuration
	m.paused = false
	return m
}

func (m model) isTimerPhase() bool {
	return m.phase == phaseFocus || m.phase == phaseRescue
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}
