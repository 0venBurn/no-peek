package main

import "testing"

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
