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

type screen int

const (
	screenMenu screen = iota
	screenSession
)

type model struct {
	screen             screen
	mode               appMode
	problem            string
	phase              phase
	remaining          time.Duration
	focusDuration      time.Duration
	rescueDuration     time.Duration
	deepFocusDuration  time.Duration
	shortBreakDuration time.Duration
	longBreakDuration  time.Duration
	noBreaks           bool
	deepCycles         int
	completedCycles    int
	paused             bool
	menuCursor         int
	menuEditingField   string
	menuInput          string
	width              int
	height             int
}

func newAppModel(cfg config) model {
	m := model{screen: screenMenu}
	m.applyConfig(cfg)
	return m
}

func newModel(cfg config) model {
	m := model{}
	m.applyConfig(cfg)
	m.screen = screenSession
	return m
}

func (m *model) applyConfig(cfg config) {
	screen := m.screen
	menuCursor := m.menuCursor
	menuEditingField := m.menuEditingField
	menuInput := m.menuInput
	width := m.width
	height := m.height
	*m = model{
		screen:             screen,
		menuCursor:         menuCursor,
		menuEditingField:   menuEditingField,
		menuInput:          menuInput,
		width:              width,
		height:             height,
		mode:               cfg.mode,
		problem:            cfg.problem,
		focusDuration:      cfg.focusDuration,
		rescueDuration:     cfg.rescueDuration,
		deepFocusDuration:  cfg.deepFocusDuration,
		shortBreakDuration: cfg.shortBreakDuration,
		longBreakDuration:  cfg.longBreakDuration,
		noBreaks:           cfg.noBreaks,
		deepCycles:         cfg.deepCycles,
	}
	if cfg.mode == modeDeep {
		m.phase = phaseDeepFocusOne
		m.remaining = cfg.deepFocusDuration
	} else {
		m.phase = phaseFocus
		m.remaining = cfg.focusDuration
	}
}

func (m model) startSession() (tea.Model, tea.Cmd) {
	cfg := m.config()
	m.applyConfig(cfg)
	m.screen = screenSession
	return m, tick()
}

func (m model) config() config {
	return config{mode: m.mode, problem: m.problem, focusDuration: m.focusDuration, rescueDuration: m.rescueDuration, deepFocusDuration: m.deepFocusDuration, shortBreakDuration: m.shortBreakDuration, longBreakDuration: m.longBreakDuration, noBreaks: m.noBreaks, deepCycles: m.deepCycles}
}

func (m model) returnToMenu() (tea.Model, tea.Cmd) {
	m.screen = screenMenu
	m.paused = false
	m.completedCycles = 0
	return m, nil
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
		if m.screen == screenSession {
			return m.handleTick()
		}
	}

	return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.screen == screenMenu {
		return m.handleMenuKey(msg)
	}
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "q":
		return m.returnToMenu()
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
	case "enter":
		if m.phase == phaseEditorial {
			return m.returnToMenu()
		}
	}

	return m, nil
}

func (m model) handleMenuKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	activeField := m.menuEditingField
	if activeField == "" {
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "s":
			return m.startSession()
		case "enter":
			m.menuEditingField = m.selectedMenuField()
			m.menuInput = ""
		case "up", "k", "shift+tab":
			m.moveMenuCursor(-1)
		case "down", "j", "tab":
			m.moveMenuCursor(1)
		}
		return m, nil
	}

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "enter":
		m.commitMenuEdit()
	case "esc":
		m.menuEditingField = ""
		m.menuInput = ""
	case "left", "h":
		if activeField == "cycles" || activeField == "mode" || activeField == "breaks" {
			m.adjustSelected(-1)
		}
	case "right", "l":
		if activeField == "cycles" || activeField == "mode" || activeField == "breaks" {
			m.adjustSelected(1)
		}
	case " ":
		if activeField == "mode" {
			m.toggleMode()
		} else if activeField == "breaks" {
			m.noBreaks = !m.noBreaks
		} else if activeField == "session" {
			m.problem += " "
		}
	case "backspace":
		if activeField == "session" && len(m.problem) > 0 {
			m.problem = m.problem[:len(m.problem)-1]
		} else if m.isMinuteField(activeField) && len(m.menuInput) > 0 {
			m.menuInput = m.menuInput[:len(m.menuInput)-1]
		}
	default:
		s := msg.String()
		if activeField == "session" && len(s) == 1 && s[0] >= 32 && s[0] <= 126 {
			m.problem += s
		} else if m.isMinuteField(activeField) && len(s) == 1 && s[0] >= '0' && s[0] <= '9' && len(m.menuInput) < 3 {
			m.menuInput += s
		}
	}
	return m, nil
}

func (m *model) toggleMode() {
	if m.mode == modePuzzle {
		m.mode = modeDeep
	} else {
		m.mode = modePuzzle
	}
	if m.menuCursor >= len(m.menuFields()) {
		m.menuCursor = len(m.menuFields()) - 1
	}
}

func (m model) menuFields() []string {
	if m.mode == modePuzzle {
		return []string{"mode", "session", "focus", "rescue"}
	}
	return []string{"mode", "session", "deepFocus", "shortBreak", "longBreak", "breaks", "cycles"}
}

func (m model) selectedMenuField() string {
	fields := m.menuFields()
	if m.menuCursor < 0 || m.menuCursor >= len(fields) {
		return fields[0]
	}
	return fields[m.menuCursor]
}

func (m *model) moveMenuCursor(delta int) {
	fields := m.menuFields()
	m.menuCursor = (m.menuCursor + delta + len(fields)) % len(fields)
	m.menuEditingField = ""
	m.menuInput = ""
}

func (m *model) commitMenuEdit() {
	if m.isMinuteField(m.menuEditingField) && m.menuInput != "" {
		minutes := 0
		for _, r := range m.menuInput {
			minutes = minutes*10 + int(r-'0')
		}
		if minutes < 1 {
			minutes = 1
		}
		m.setMinuteFieldValue(m.menuEditingField, minutes)
	}
	m.menuEditingField = ""
	m.menuInput = ""
}

func (m model) isMinuteField(field string) bool {
	switch field {
	case "focus", "rescue", "deepFocus", "shortBreak", "longBreak":
		return true
	default:
		return false
	}
}

func (m model) minuteFieldValue(field string) int {
	switch field {
	case "focus":
		return int(m.focusDuration / time.Minute)
	case "rescue":
		return int(m.rescueDuration / time.Minute)
	case "deepFocus":
		return int(m.deepFocusDuration / time.Minute)
	case "shortBreak":
		return int(m.shortBreakDuration / time.Minute)
	case "longBreak":
		return int(m.longBreakDuration / time.Minute)
	default:
		return 1
	}
}

func (m *model) setMinuteFieldValue(field string, minutes int) {
	d := time.Duration(minutes) * time.Minute
	switch field {
	case "focus":
		m.focusDuration = d
	case "rescue":
		m.rescueDuration = d
	case "deepFocus":
		m.deepFocusDuration = d
	case "shortBreak":
		m.shortBreakDuration = d
	case "longBreak":
		m.longBreakDuration = d
	}
}

func (m *model) adjustSelected(delta int) {
	step := time.Duration(delta) * time.Minute
	switch m.selectedMenuField() {
	case "mode":
		m.toggleMode()
	case "focus":
		m.focusDuration = clampDuration(m.focusDuration + step)
	case "rescue":
		m.rescueDuration = clampDuration(m.rescueDuration + step)
	case "deepFocus":
		m.deepFocusDuration = clampDuration(m.deepFocusDuration + step)
	case "shortBreak":
		m.shortBreakDuration = clampDuration(m.shortBreakDuration + step)
	case "longBreak":
		m.longBreakDuration = clampDuration(m.longBreakDuration + step)
	case "breaks":
		m.noBreaks = !m.noBreaks
	case "cycles":
		m.deepCycles += delta
		if m.deepCycles < 1 {
			m.deepCycles = 1
		}
		if m.deepCycles > 99 {
			m.deepCycles = 99
		}
	}
}

func clampDuration(d time.Duration) time.Duration {
	if d < time.Minute {
		return time.Minute
	}
	return d
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
		if m.noBreaks {
			m.completedCycles++
			if m.completedCycles >= m.deepCycles {
				return m.returnToMenu()
			}
			m = m.startDeepCycle()
			return m, tea.Batch(tick(), notify("No Peek", fmt.Sprintf("%s focus block complete. Starting a new %s focus block.", formatDuration(m.deepFocusDuration), formatDuration(m.deepFocusDuration))))
		}
		m.phase = phaseDeepShortBreak
		m.remaining = m.shortBreakDuration
		return m, tea.Batch(tick(), notify("No Peek", fmt.Sprintf("%s focus block complete. Time for a %s short break.", formatDuration(m.deepFocusDuration), formatDuration(m.shortBreakDuration))))
	case phaseDeepShortBreak:
		m.phase = phaseDeepFocusTwo
		m.remaining = m.deepFocusDuration
		return m, tea.Batch(tick(), notify("No Peek", fmt.Sprintf("%s short break complete. Time for another %s focus block.", formatDuration(m.shortBreakDuration), formatDuration(m.deepFocusDuration))))
	case phaseDeepFocusTwo:
		if m.noBreaks {
			m.completedCycles++
			if m.completedCycles >= m.deepCycles {
				return m.returnToMenu()
			}
			m = m.startDeepCycle()
			return m, tea.Batch(tick(), notify("No Peek", fmt.Sprintf("%s focus block complete. Starting a new %s focus block.", formatDuration(m.deepFocusDuration), formatDuration(m.deepFocusDuration))))
		}
		m.phase = phaseDeepLongBreak
		m.remaining = m.longBreakDuration
		return m, tea.Batch(tick(), notify("No Peek", fmt.Sprintf("%s focus block complete. Time for a %s long break.", formatDuration(m.deepFocusDuration), formatDuration(m.longBreakDuration))))
	case phaseDeepLongBreak:
		m.completedCycles++
		if m.completedCycles >= m.deepCycles {
			return m.returnToMenu()
		}
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
