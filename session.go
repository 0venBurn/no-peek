package main

import (
	"fmt"
	"time"
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

type sessionState struct {
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
}

type sessionTransition struct {
	session      sessionState
	nextTick     bool
	finished     bool
	notification notificationIntent
}

type timerPhaseDefinition struct {
	label    func(sessionState) string
	subtitle string
	duration func(sessionState) time.Duration
}

type checkInDefinition struct {
	title         string
	prompt        string
	continueLabel string
	continueHint  func(sessionState) string
	stuckLabel    string
	stuckHint     func(sessionState) string
}

type sessionRenderKind int

const (
	sessionRenderTimer sessionRenderKind = iota
	sessionRenderCheckIn
	sessionRenderEditorial
)

type sessionRender struct {
	problem   string
	kind      sessionRenderKind
	timer     timerRender
	checkIn   checkInRender
	editorial editorialRender
}

type timerRender struct {
	label     string
	remaining time.Duration
	paused    bool
	subtitle  string
}

type checkInRender struct {
	title         string
	prompt        string
	continueLabel string
	continueHint  string
	stuckLabel    string
	stuckHint     string
}

type editorialRender struct {
	title   string
	message string
}

var timerPhaseDefinitions = map[phase]timerPhaseDefinition{
	phaseFocus: {
		label:    staticLabel("Focus round"),
		subtitle: "Stay with the problem. No hints yet.",
		duration: func(s sessionState) time.Duration { return s.focusDuration },
	},
	phaseRescue: {
		label:    staticLabel("Rescue round"),
		subtitle: "Try examples, invariants, brute force, or a smaller case.",
		duration: func(s sessionState) time.Duration { return s.rescueDuration },
	},
	phaseDeepFocusOne: {
		label: func(s sessionState) string {
			if s.noBreaks {
				return "Deep focus"
			}
			return "Deep focus 1/2"
		},
		subtitle: "No distractions.",
		duration: func(s sessionState) time.Duration { return s.deepFocusDuration },
	},
	phaseDeepShortBreak: {
		label:    staticLabel("Short break"),
		subtitle: "Rest. Don't distract yourself.",
		duration: func(s sessionState) time.Duration { return s.shortBreakDuration },
	},
	phaseDeepFocusTwo: {
		label:    staticLabel("Deep focus 2/2"),
		subtitle: "No distractions.",
		duration: func(s sessionState) time.Duration { return s.deepFocusDuration },
	},
	phaseDeepLongBreak: {
		label:    staticLabel("Long break"),
		subtitle: "Rest. Don't distract yourself.",
		duration: func(s sessionState) time.Duration { return s.longBreakDuration },
	},
}

var checkInDefinitions = map[phase]checkInDefinition{
	phaseCheckIn: {
		title:         "Time's up.",
		prompt:        "Are you still generating new ideas?",
		continueLabel: "still thinking",
		continueHint:  func(s sessionState) string { return fmt.Sprintf("another %s", formatDuration(s.focusDuration)) },
		stuckLabel:    "stuck",
		stuckHint:     func(s sessionState) string { return fmt.Sprintf("%s rescue", formatDuration(s.rescueDuration)) },
	},
	phaseFinalCheckIn: {
		title:         "Rescue time is up.",
		prompt:        "Did you find a new thread to pull on?",
		continueLabel: "yes",
		continueHint:  func(s sessionState) string { return fmt.Sprintf("continue %s", formatDuration(s.focusDuration)) },
		stuckLabel:    "no",
		stuckHint:     staticHint("read editorial"),
	},
}

func staticLabel(label string) func(sessionState) string {
	return func(sessionState) string { return label }
}

func staticHint(hint string) func(sessionState) string {
	return func(sessionState) string { return hint }
}

func newSessionState(cfg config) sessionState {
	cfg = normalizeConfig(cfg)
	s := sessionState{
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
		return s.startDeepCycle()
	}
	return s.startFocusRound()
}

func (s sessionState) config() config {
	return newConfig(
		s.mode,
		s.problem,
		s.focusDuration,
		s.rescueDuration,
		s.deepFocusDuration,
		s.shortBreakDuration,
		s.longBreakDuration,
		s.noBreaks,
		s.deepCycles,
	)
}

func (s sessionState) render() sessionRender {
	render := sessionRender{problem: s.problem}
	if timer, ok := s.timerRender(); ok {
		render.kind = sessionRenderTimer
		render.timer = timer
		return render
	}
	if checkIn, ok := s.checkInRender(); ok {
		render.kind = sessionRenderCheckIn
		render.checkIn = checkIn
		return render
	}
	render.kind = sessionRenderEditorial
	render.editorial = editorialRender{
		title:   "READ THE EDITORIAL",
		message: "You gave the problem a real attempt. Learn the missing idea, then try to re-solve it without looking.",
	}
	return render
}

func (s sessionState) timerRender() (timerRender, bool) {
	def, ok := timerPhaseDefinitions[s.phase]
	if !ok {
		return timerRender{}, false
	}
	return timerRender{
		label:     def.label(s),
		remaining: s.remaining,
		paused:    s.paused,
		subtitle:  def.subtitle,
	}, true
}

func (s sessionState) checkInRender() (checkInRender, bool) {
	def, ok := checkInDefinitions[s.phase]
	if !ok {
		return checkInRender{}, false
	}
	return checkInRender{
		title:         def.title,
		prompt:        def.prompt,
		continueLabel: def.continueLabel,
		continueHint:  def.continueHint(s),
		stuckLabel:    def.stuckLabel,
		stuckHint:     def.stuckHint(s),
	}, true
}

func (s sessionState) togglePause() sessionState {
	if s.isTimerPhase() {
		s.paused = !s.paused
	}
	return s
}

func (s sessionState) continueThinking() sessionTransition {
	if s.phase != phaseCheckIn && s.phase != phaseFinalCheckIn {
		return s.noop()
	}
	return sessionTransition{session: s.startFocusRound(), nextTick: true}
}

func (s sessionState) restartFromEditorial() sessionTransition {
	if s.phase != phaseEditorial {
		return s.noop()
	}
	return sessionTransition{session: s.startFocusRound(), nextTick: true}
}

func (s sessionState) stuck() sessionTransition {
	switch s.phase {
	case phaseCheckIn:
		return sessionTransition{session: s.startTimerPhase(phaseRescue), nextTick: true}
	case phaseFinalCheckIn:
		s.phase = phaseEditorial
		s.paused = false
		return sessionTransition{session: s, notification: notifyIntent("No Peek", "Time to read the editorial.")}
	default:
		return s.noop()
	}
}

func (s sessionState) tick() sessionTransition {
	if !s.isTimerPhase() {
		return s.noop()
	}
	if s.paused {
		return sessionTransition{session: s, nextTick: true}
	}

	s.remaining -= time.Second
	if s.remaining > 0 {
		return sessionTransition{session: s, nextTick: true}
	}

	s.remaining = 0
	return s.advanceAfterTimerExpires()
}

func (s sessionState) advanceAfterTimerExpires() sessionTransition {
	switch s.phase {
	case phaseFocus:
		s.phase = phaseCheckIn
		return sessionTransition{
			session:      s,
			notification: notifyIntent("No Peek", fmt.Sprintf("%s are up. Are you still thinking or stuck?", formatDuration(s.focusDuration))),
		}
	case phaseRescue:
		s.phase = phaseFinalCheckIn
		return sessionTransition{
			session:      s,
			notification: notifyIntent("No Peek", fmt.Sprintf("%s rescue is up. Still stuck?", formatDuration(s.rescueDuration))),
		}
	case phaseDeepFocusOne:
		return s.finishDeepFocusBlock()
	case phaseDeepShortBreak:
		s = s.startTimerPhase(phaseDeepFocusTwo)
		return sessionTransition{
			session:      s,
			nextTick:     true,
			notification: notifyIntent("No Peek", fmt.Sprintf("%s short break complete. Time for another %s focus block.", formatDuration(s.shortBreakDuration), formatDuration(s.deepFocusDuration))),
		}
	case phaseDeepFocusTwo:
		s = s.startTimerPhase(phaseDeepLongBreak)
		return sessionTransition{
			session:      s,
			nextTick:     true,
			notification: notifyIntent("No Peek", fmt.Sprintf("%s focus block complete. Time for a %s long break.", formatDuration(s.deepFocusDuration), formatDuration(s.longBreakDuration))),
		}
	case phaseDeepLongBreak:
		s.completedCycles++
		if s.completedCycles >= s.deepCycles {
			return sessionTransition{session: s, finished: true}
		}
		s = s.startDeepCycle()
		return sessionTransition{
			session:      s,
			nextTick:     true,
			notification: notifyIntent("No Peek", fmt.Sprintf("%s long break complete. Starting a new %s focus block.", formatDuration(s.longBreakDuration), formatDuration(s.deepFocusDuration))),
		}
	default:
		return s.noop()
	}
}

func (s sessionState) finishDeepFocusBlock() sessionTransition {
	if s.noBreaks {
		s.completedCycles++
		if s.completedCycles >= s.deepCycles {
			return sessionTransition{session: s, finished: true}
		}
		s = s.startDeepCycle()
		return sessionTransition{
			session:      s,
			nextTick:     true,
			notification: notifyIntent("No Peek", fmt.Sprintf("%s focus block complete. Starting a new %s focus block.", formatDuration(s.deepFocusDuration), formatDuration(s.deepFocusDuration))),
		}
	}

	s = s.startTimerPhase(phaseDeepShortBreak)
	return sessionTransition{
		session:      s,
		nextTick:     true,
		notification: notifyIntent("No Peek", fmt.Sprintf("%s focus block complete. Time for a %s short break.", formatDuration(s.deepFocusDuration), formatDuration(s.shortBreakDuration))),
	}
}

func (s sessionState) startFocusRound() sessionState {
	return s.startTimerPhase(phaseFocus)
}

func (s sessionState) startDeepCycle() sessionState {
	return s.startTimerPhase(phaseDeepFocusOne)
}

func (s sessionState) startTimerPhase(phase phase) sessionState {
	def := timerPhaseDefinitions[phase]
	s.phase = phase
	s.remaining = def.duration(s)
	s.paused = false
	return s
}

func (s sessionState) isTimerPhase() bool {
	_, ok := timerPhaseDefinitions[s.phase]
	return ok
}

func (s sessionState) noop() sessionTransition {
	return sessionTransition{session: s}
}
