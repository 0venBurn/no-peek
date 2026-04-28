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

type config struct {
	problem        string
	focusDuration  time.Duration
	rescueDuration time.Duration
}

func parseConfig() config {
	focusMinutes := flag.Int("focus", 30, "focus round length in minutes")
	rescueMinutes := flag.Int("rescue", 15, "stuck/rescue round length in minutes")
	flag.Parse()

	problem := strings.TrimSpace(strings.Join(flag.Args(), " "))
	if problem == "" {
		problem = "Untitled problem"
	}

	return config{
		problem:        problem,
		focusDuration:  time.Duration(*focusMinutes) * time.Minute,
		rescueDuration: time.Duration(*rescueMinutes) * time.Minute,
	}
}

func run(cfg config) error {
	_, err := tea.NewProgram(newModel(cfg), tea.WithAltScreen()).Run()
	return err
}
