package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	cfg := parseConfig()
	if err := run(cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type appMode string

const (
	modePuzzle appMode = "puzzle"
	modeDeep   appMode = "deep"
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
}

func parseConfig() config {
	modeValue := flag.String("mode", "", "required mode to run: puzzle or deep")
	focusMinutes := flag.Int("focus", 30, "puzzle focus round length in minutes")
	rescueMinutes := flag.Int("rescue", 15, "puzzle stuck/rescue round length in minutes")
	deepFocusMinutes := flag.Int("deep-focus", 45, "deep work focus length in minutes")
	shortBreakMinutes := flag.Int("short-break", 5, "deep work short break length in minutes")
	longBreakMinutes := flag.Int("long-break", 20, "deep work long break length in minutes")
	noBreaks := flag.Bool("no-breaks", false, "deep work mode: continuous focus without breaks")
	flag.Parse()

	mode := appMode(strings.ToLower(strings.TrimSpace(*modeValue)))
	if mode == "" {
		fmt.Fprintln(os.Stderr, "missing required --mode flag: use --mode puzzle or --mode deep")
		flag.Usage()
		os.Exit(2)
	}
	if mode != modePuzzle && mode != modeDeep {
		fmt.Fprintf(os.Stderr, "unknown mode %q: use --mode puzzle or --mode deep\n", *modeValue)
		flag.Usage()
		os.Exit(2)
	}

	problem := strings.TrimSpace(strings.Join(flag.Args(), " "))
	if problem == "" {
		problem = "Untitled problem"
		if mode == modeDeep {
			problem = "Deep work"
		}
	}

	return config{
		mode:               mode,
		problem:            problem,
		focusDuration:      time.Duration(*focusMinutes) * time.Minute,
		rescueDuration:     time.Duration(*rescueMinutes) * time.Minute,
		deepFocusDuration:  time.Duration(*deepFocusMinutes) * time.Minute,
		shortBreakDuration: time.Duration(*shortBreakMinutes) * time.Minute,
		longBreakDuration:  time.Duration(*longBreakMinutes) * time.Minute,
		noBreaks:           *noBreaks,
	}
}

func run(cfg config) error {
	_, err := tea.NewProgram(newModel(cfg), tea.WithAltScreen()).Run()
	return err
}
