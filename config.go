package main

import "time"

type appMode string

const (
	modePuzzle appMode = "puzzle"
	modeDeep   appMode = "deep"
)

func validMode(mode appMode) bool {
	return mode == modePuzzle || mode == modeDeep
}

func defaultProblem(mode appMode) string {
	if mode == modeDeep {
		return "Deep work"
	}
	return "Untitled problem"
}

const (
	defaultPuzzleFocusMinutes = 30
	defaultRescueMinutes      = 15
	defaultDeepFocusMinutes   = 45
	defaultShortBreakMinutes  = 5
	defaultLongBreakMinutes   = 20
	defaultDeepCycles         = 1
	maxDeepCycles             = 99
)

type config struct {
	mode               appMode
	problem            string
	focusDuration      time.Duration
	rescueDuration     time.Duration
	deepFocusDuration  time.Duration
	shortBreakDuration time.Duration
	longBreakDuration  time.Duration
	noBreaks           bool
	deepCycles         int
}

func newConfig(mode appMode, problem string, focusDuration, rescueDuration, deepFocusDuration, shortBreakDuration, longBreakDuration time.Duration, noBreaks bool, deepCycles int) config {
	return normalizeConfig(config{
		mode:               mode,
		problem:            problem,
		focusDuration:      focusDuration,
		rescueDuration:     rescueDuration,
		deepFocusDuration:  deepFocusDuration,
		shortBreakDuration: shortBreakDuration,
		longBreakDuration:  longBreakDuration,
		noBreaks:           noBreaks,
		deepCycles:         deepCycles,
	})
}

func defaultConfig() config {
	return config{
		mode:               modePuzzle,
		problem:            defaultProblem(modePuzzle),
		focusDuration:      time.Duration(defaultPuzzleFocusMinutes) * time.Minute,
		rescueDuration:     time.Duration(defaultRescueMinutes) * time.Minute,
		deepFocusDuration:  time.Duration(defaultDeepFocusMinutes) * time.Minute,
		shortBreakDuration: time.Duration(defaultShortBreakMinutes) * time.Minute,
		longBreakDuration:  time.Duration(defaultLongBreakMinutes) * time.Minute,
		deepCycles:         defaultDeepCycles,
	}
}

func normalizeConfig(cfg config) config {
	defaults := defaultConfig()

	if !validMode(cfg.mode) {
		cfg.mode = defaults.mode
	}
	if cfg.problem == "" {
		cfg.problem = defaultProblem(cfg.mode)
	}
	if cfg.focusDuration <= 0 {
		cfg.focusDuration = defaults.focusDuration
	}
	if cfg.rescueDuration <= 0 {
		cfg.rescueDuration = defaults.rescueDuration
	}
	if cfg.deepFocusDuration <= 0 {
		cfg.deepFocusDuration = defaults.deepFocusDuration
	}
	if cfg.shortBreakDuration <= 0 {
		cfg.shortBreakDuration = defaults.shortBreakDuration
	}
	if cfg.longBreakDuration <= 0 {
		cfg.longBreakDuration = defaults.longBreakDuration
	}

	cfg.focusDuration = clampDuration(cfg.focusDuration)
	cfg.rescueDuration = clampDuration(cfg.rescueDuration)
	cfg.deepFocusDuration = clampDuration(cfg.deepFocusDuration)
	cfg.shortBreakDuration = clampDuration(cfg.shortBreakDuration)
	cfg.longBreakDuration = clampDuration(cfg.longBreakDuration)
	cfg.deepCycles = clampDeepCycles(cfg.deepCycles)

	return cfg
}

func clampDuration(d time.Duration) time.Duration {
	if d < time.Minute {
		return time.Minute
	}
	return d
}

func clampDeepCycles(cycles int) int {
	if cycles < 1 {
		return 1
	}
	if cycles > maxDeepCycles {
		return maxDeepCycles
	}
	return cycles
}

func wholeMinutes(d time.Duration) int {
	return int(d / time.Minute)
}
