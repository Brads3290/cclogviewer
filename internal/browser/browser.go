package browser

import (
	"fmt"
	"os/exec"
	"runtime"
)

// OpenInBrowser opens a file in the system's default browser.
func OpenInBrowser(filename string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", filename)
	case "linux":
		cmd = exec.Command("xdg-open", filename)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", filename)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}
