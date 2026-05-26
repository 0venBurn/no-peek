package main

import (
	"testing"
	"time"
)

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
