package main

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func keyMsg(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func TestModelIgnoresStaleSessionTicks(t *testing.T) {
	m := newModel(config{
		mode:          modePuzzle,
		problem:       "Two Sum",
		focusDuration: time.Minute,
	})
	m.sessionID = 2
	m.session.remaining = time.Minute

	updated, _ := m.Update(tickMsg{sessionID: 1})
	got := updated.(model)
	if got.session.remaining != time.Minute {
		t.Fatalf("remaining after stale tick = %s, want %s", got.session.remaining, time.Minute)
	}
}

func TestModelSolvedKeyReturnsPuzzleSessionToMenu(t *testing.T) {
	m := newModel(config{
		mode:          modePuzzle,
		problem:       "Two Sum",
		focusDuration: time.Minute,
	})

	updated, cmd := m.Update(keyMsg("c"))
	got := updated.(model)
	if got.screen != screenMenu {
		t.Fatalf("screen after solved key = %v, want menu", got.screen)
	}
	if cmd == nil {
		t.Fatal("command after solved key = nil, want congrats notification command")
	}
}

func TestModelReturnsMenuAndNotificationCommandWhenSessionFinishes(t *testing.T) {
	m := newModel(config{
		mode:              modeDeep,
		problem:           "Write",
		deepFocusDuration: time.Minute,
		noBreaks:          true,
		deepCycles:        1,
	})
	m.session.remaining = time.Second

	updated, cmd := m.Update(tickMsg{sessionID: m.sessionID})
	got := updated.(model)
	if got.screen != screenMenu {
		t.Fatalf("screen after final tick = %v, want menu", got.screen)
	}
	if cmd == nil {
		t.Fatal("command after final tick = nil, want notification command")
	}
}
