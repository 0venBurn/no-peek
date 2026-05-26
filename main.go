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

func parseConfig() config {
	defaults := defaultConfig()
	modeValue := flag.String("mode", "", "initial mode: puzzle or deep")
	focusMinutes := flag.Int("focus", wholeMinutes(defaults.focusDuration), "puzzle focus round length in minutes")
	rescueMinutes := flag.Int("rescue", wholeMinutes(defaults.rescueDuration), "puzzle stuck/rescue round length in minutes")
	deepFocusMinutes := flag.Int("deep-focus", wholeMinutes(defaults.deepFocusDuration), "deep work focus length in minutes")
	shortBreakMinutes := flag.Int("short-break", wholeMinutes(defaults.shortBreakDuration), "deep work short break length in minutes")
	longBreakMinutes := flag.Int("long-break", wholeMinutes(defaults.longBreakDuration), "deep work long break length in minutes")
	noBreaks := flag.Bool("no-breaks", defaults.noBreaks, "deep work mode: continuous focus without breaks")
	deepCycles := flag.Int("cycles", defaults.deepCycles, "deep work cycles before returning to menu")
	flag.Parse()

	mode := appMode(strings.ToLower(strings.TrimSpace(*modeValue)))
	if mode != "" && !validMode(mode) {
		fmt.Fprintf(os.Stderr, "unknown mode %q: use --mode puzzle or --mode deep\n", *modeValue)
		flag.Usage()
		os.Exit(2)
	}

	return newConfig(
		mode,
		strings.TrimSpace(strings.Join(flag.Args(), " ")),
		time.Duration(*focusMinutes)*time.Minute,
		time.Duration(*rescueMinutes)*time.Minute,
		time.Duration(*deepFocusMinutes)*time.Minute,
		time.Duration(*shortBreakMinutes)*time.Minute,
		time.Duration(*longBreakMinutes)*time.Minute,
		*noBreaks,
		*deepCycles,
	)
}

func run(cfg config) error {
	_, err := tea.NewProgram(newAppModel(cfg), tea.WithAltScreen()).Run()
	return err
}
