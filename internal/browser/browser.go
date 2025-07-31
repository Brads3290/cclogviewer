package browser

import (
	"fmt"
	"github.com/brads3290/cclogviewer/internal/constants"
	"os/exec"
	"runtime"
)

// OpenInBrowser opens a file in the system's default browser.
func OpenInBrowser(filename string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case constants.PlatformDarwin:
		cmd = exec.Command(constants.MacOSOpenCommand, filename)
	case constants.PlatformLinux:
		cmd = exec.Command(constants.LinuxOpenCommand, filename)
	case constants.PlatformWindows:
		cmd = exec.Command(constants.WindowsCommand, constants.WindowsCmdFlag, constants.WindowsStartCommand, filename)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}
