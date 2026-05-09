package main

import (
	"fmt"
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
	phaseDeepFocusOne
	phaseDeepShortBreak
	phaseDeepFocusTwo
	phaseDeepLongBreak
)

type tickMsg time.Time

type model struct {
	mode               appMode
	problem            string
	phase              phase
	remaining          time.Duration
	focusDuration      time.Duration
	rescueDuration     time.Duration
	deepFocusDuration  time.Duration
	shortBreakDuration time.Duration
	longBreakDuration  time.Duration
	paused             bool
	width              int
	height             int
}

func newModel(cfg config) model {
	m := model{
		mode:               cfg.mode,
		problem:            cfg.problem,
		focusDuration:      cfg.focusDuration,
		rescueDuration:     cfg.rescueDuration,
		deepFocusDuration:  cfg.deepFocusDuration,
		shortBreakDuration: cfg.shortBreakDuration,
		longBreakDuration:  cfg.longBreakDuration,
	}
	if cfg.mode == modeDeep {
		m.phase = phaseDeepFocusOne
		m.remaining = cfg.deepFocusDuration
	} else {
		m.phase = phaseFocus
		m.remaining = cfg.focusDuration
	}
	return m
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
		return m, notify("No Peek", fmt.Sprintf("%s are up. Are you still thinking or stuck?", formatDuration(m.focusDuration)))
	case phaseRescue:
		m.phase = phaseFinalCheckIn
		return m, notify("No Peek", fmt.Sprintf("%s rescue is up. Still stuck?", formatDuration(m.rescueDuration)))
	case phaseDeepFocusOne:
		m.phase = phaseDeepShortBreak
		m.remaining = m.shortBreakDuration
		return m, tea.Batch(tick(), notify("No Peek", fmt.Sprintf("%s focus block complete. Time for a %s short break.", formatDuration(m.deepFocusDuration), formatDuration(m.shortBreakDuration))))
	case phaseDeepShortBreak:
		m.phase = phaseDeepFocusTwo
		m.remaining = m.deepFocusDuration
		return m, tea.Batch(tick(), notify("No Peek", fmt.Sprintf("%s short break complete. Time for another %s focus block.", formatDuration(m.shortBreakDuration), formatDuration(m.deepFocusDuration))))
	case phaseDeepFocusTwo:
		m.phase = phaseDeepLongBreak
		m.remaining = m.longBreakDuration
		return m, tea.Batch(tick(), notify("No Peek", fmt.Sprintf("%s focus block complete. Time for a %s long break.", formatDuration(m.deepFocusDuration), formatDuration(m.longBreakDuration))))
	case phaseDeepLongBreak:
		m = m.startDeepCycle()
		return m, tea.Batch(tick(), notify("No Peek", fmt.Sprintf("%s long break complete. Starting a new %s focus block.", formatDuration(m.longBreakDuration), formatDuration(m.deepFocusDuration))))
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

func (m model) startDeepCycle() model {
	m.phase = phaseDeepFocusOne
	m.remaining = m.deepFocusDuration
	m.paused = false
	return m
}

func (m model) isTimerPhase() bool {
	return m.phase == phaseFocus || m.phase == phaseRescue ||
		m.phase == phaseDeepFocusOne || m.phase == phaseDeepShortBreak ||
		m.phase == phaseDeepFocusTwo || m.phase == phaseDeepLongBreak
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}
