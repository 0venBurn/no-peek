package main

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type tickMsg struct {
	at        time.Time
	sessionID int
}

type screen int

const (
	screenMenu screen = iota
	screenSession
)

type model struct {
	screen    screen
	menu      menuState
	session   sessionState
	sessionID int
	width     int
	height    int
}

func newAppModel(cfg config) model {
	return model{
		screen: screenMenu,
		menu:   newMenuState(cfg),
	}
}

func newModel(cfg config) model {
	cfg = normalizeConfig(cfg)
	return model{
		screen:    screenSession,
		menu:      newMenuState(cfg),
		session:   newSessionState(cfg),
		sessionID: 1,
	}
}

func (m model) Init() tea.Cmd {
	if m.screen == screenSession {
		return tick(m.sessionID)
	}
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		return m.handleKey(msg)
	case tickMsg:
		if m.screen == screenSession && msg.sessionID == m.sessionID {
			return m.applySessionTransition(m.session.tick())
		}
	}
	return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.screen == screenMenu {
		return m.handleMenuKey(msg)
	}
	return m.handleSessionKey(msg)
}

func (m model) handleMenuKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	outcome := m.menu.handleKey(msg)
	m.menu = outcome.menu
	if outcome.quit {
		return m, tea.Quit
	}
	if outcome.start {
		m.sessionID++
		m.session = newSessionState(m.menu.config())
		m.screen = screenSession
		return m, tick(m.sessionID)
	}
	return m, nil
}

func (m model) handleSessionKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "q":
		return m.returnToMenu()
	case "p", " ":
		m.session = m.session.togglePause()
	case "t":
		return m.applySessionTransition(m.session.continueThinking())
	case "s":
		return m.applySessionTransition(m.session.stuck())
	case "c":
		return m.applySessionTransition(m.session.solved())
	case "r":
		return m.applySessionTransition(m.session.restartFromEditorial())
	case "enter":
		if m.session.phase == phaseEditorial {
			return m.returnToMenu()
		}
	}
	return m, nil
}

func (m model) applySessionTransition(transition sessionTransition) (tea.Model, tea.Cmd) {
	m.session = transition.session

	var cmds []tea.Cmd
	if transition.nextTick {
		cmds = append(cmds, tick(m.sessionID))
	}
	if transition.notification.enabled {
		cmds = append(cmds, transition.notification.cmd())
	}
	if transition.finished {
		next, cmd := m.returnToMenu()
		cmds = append(cmds, cmd)
		return next, tea.Batch(cmds...)
	}
	return m, tea.Batch(cmds...)
}

func (m model) returnToMenu() (tea.Model, tea.Cmd) {
	m.menu = newMenuState(m.session.config())
	m.screen = screenMenu
	m.session = sessionState{}
	return m, nil
}

func tick(sessionID int) tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg{at: t, sessionID: sessionID}
	})
}
