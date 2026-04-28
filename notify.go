package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"
)

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
				_ = exec.Command(path, title, message).Run()
			}
		case "darwin":
			script := fmt.Sprintf(`display notification %q with title %q`, message, title)
			_ = exec.Command("osascript", "-e", script).Run()
		case "windows":
			ps := fmt.Sprintf(`[console]::beep(); Add-Type -AssemblyName System.Windows.Forms; [System.Windows.Forms.MessageBox]::Show(%q, %q) | Out-Null`, message, title)
			_ = exec.Command("powershell", "-NoProfile", "-Command", ps).Run()
		}
		return nil
	}
}
