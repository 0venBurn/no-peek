package main

import (
	"testing"
	"time"
)

func TestPuzzleSessionFlow(t *testing.T) {
	s := newSessionState(config{
		mode:           modePuzzle,
		problem:        "Two Sum",
		focusDuration:  time.Minute,
		rescueDuration: time.Minute,
	})
	s.remaining = time.Second

	tr := s.tick()
	if tr.session.phase != phaseCheckIn || !tr.notification.enabled {
		t.Fatalf("focus expiry = phase %v notification %v, want check-in notification", tr.session.phase, tr.notification.enabled)
	}

	tr = tr.session.stuck()
	if tr.session.phase != phaseRescue || !tr.nextTick {
		t.Fatalf("stuck from check-in = phase %v nextTick %v, want rescue next tick", tr.session.phase, tr.nextTick)
	}

	s = tr.session
	s.remaining = time.Second
	tr = s.tick()
	if tr.session.phase != phaseFinalCheckIn || !tr.notification.enabled {
		t.Fatalf("rescue expiry = phase %v notification %v, want final check-in notification", tr.session.phase, tr.notification.enabled)
	}

	tr = tr.session.stuck()
	if tr.session.phase != phaseEditorial || !tr.notification.enabled {
		t.Fatalf("stuck from final check-in = phase %v notification %v, want editorial notification", tr.session.phase, tr.notification.enabled)
	}
}

func TestDeepNoBreakCyclesFinishAfterConfiguredFocusBlocks(t *testing.T) {
	s := newSessionState(config{
		mode:              modeDeep,
		problem:           "Write",
		deepFocusDuration: time.Minute,
		noBreaks:          true,
		deepCycles:        2,
	})
	s.remaining = time.Second

	tr := s.tick()
	if tr.finished || !tr.nextTick || tr.session.completedCycles != 1 || tr.session.phase != phaseDeepFocusOne {
		t.Fatalf("first focus block = finished %v nextTick %v cycles %d phase %v, want next cycle", tr.finished, tr.nextTick, tr.session.completedCycles, tr.session.phase)
	}

	s = tr.session
	s.remaining = time.Second
	tr = s.tick()
	if !tr.finished || tr.session.completedCycles != 2 {
		t.Fatalf("second focus block = finished %v cycles %d, want finished after 2", tr.finished, tr.session.completedCycles)
	}
}
