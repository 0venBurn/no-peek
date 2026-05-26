package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestMenuModeControlsVisibleFields(t *testing.T) {
	menu := newMenuState(defaultConfig())
	if got := len(menu.fields()); got != 4 {
		t.Fatalf("puzzle fields = %d, want 4", got)
	}

	menu.toggleMode()
	if menu.cfg.mode != modeDeep {
		t.Fatalf("mode = %s, want deep", menu.cfg.mode)
	}
	if got := len(menu.fields()); got != 7 {
		t.Fatalf("deep fields = %d, want 7", got)
	}

	menu.toggleBreaks()
	fields := menu.fields()
	if got := len(fields); got != 5 {
		t.Fatalf("deep no-break fields = %d, want 5", got)
	}
	if fields[3].id != fieldBreaks || fields[4].id != fieldCycles {
		t.Fatalf("deep no-break trailing fields = %v, %v; want breaks, cycles", fields[3].id, fields[4].id)
	}
	if menu.cursor != 3 {
		t.Fatalf("cursor after hiding break durations = %d, want breaks field index 3", menu.cursor)
	}

	menu.toggleBreaks()
	menu.cursor = len(menu.fields()) - 1
	menu.toggleMode()
	if menu.cfg.mode != modePuzzle {
		t.Fatalf("mode = %s, want puzzle", menu.cfg.mode)
	}
	if menu.cursor != len(menu.fields())-1 {
		t.Fatalf("cursor = %d, want clamped to last puzzle field", menu.cursor)
	}
}

func TestMenuMinuteEditUsesSharedConfigNormalization(t *testing.T) {
	menu := newMenuState(defaultConfig())
	menu.editingField = fieldFocus
	menu.input = "0"
	menu.commitEdit()

	if got := menu.minuteField(fieldFocus); got != 1 {
		t.Fatalf("focus minutes = %d, want clamped to 1", got)
	}
}

func TestMenuCyclesAcceptsTypedNumber(t *testing.T) {
	menu := newMenuState(defaultConfig())
	menu.toggleMode()
	menu.editingField = fieldCycles
	menu.input = "12"
	menu.commitEdit()

	if menu.cfg.deepCycles != 12 {
		t.Fatalf("deep cycles = %d, want 12", menu.cfg.deepCycles)
	}
}

func TestEditingProblemNameClearsDefaultProblem(t *testing.T) {
	menu := newMenuState(defaultConfig())
	menu.cursor = 1

	outcome := menu.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if got := outcome.menu.cfg.problem; got != "" {
		t.Fatalf("problem after entering default text field = %q, want empty", got)
	}
}

func TestTypingProblemNameAcceptsHJKL(t *testing.T) {
	menu := newMenuState(defaultConfig())
	menu.editingField = fieldSession
	menu.cfg.problem = ""

	for _, r := range "hjkl" {
		menu = menu.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}).menu
	}
	if got := menu.cfg.problem; got != "hjkl" {
		t.Fatalf("problem after typing hjkl = %q, want %q", got, "hjkl")
	}
}
