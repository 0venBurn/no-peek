package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"
)

func startNotification(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdin = nil
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	_ = cmd.Start()
}

func notify(title, message string) tea.Cmd {
	return func() tea.Msg {
		// Always ring the terminal bell. This is dependency-free and works anywhere
		// the terminal has audible/visual bell enabled.
		_, _ = os.Stdout.Write([]byte("\a"))

		// Best-effort native desktop notification, with no Go dependencies.
		// If the command is unavailable, silently fall back to the bell.
		switch runtime.GOOS {
		case "linux":
			if path, err := exec.LookPath("notify-send"); err == nil {
				startNotification(path, "--app-name=No Peek", "--urgency=critical", title, message)
			}
		case "darwin":
			script := fmt.Sprintf(`display notification %q with title %q sound name "Glass"`, message, title)
			startNotification("osascript", "-e", script)
		case "windows":
			ps := fmt.Sprintf(`[console]::beep(); Add-Type -AssemblyName System.Windows.Forms; [System.Windows.Forms.MessageBox]::Show(%q, %q) | Out-Null`, message, title)
			startNotification("powershell", "-NoProfile", "-Command", ps)
		}
		return nil
	}
}
