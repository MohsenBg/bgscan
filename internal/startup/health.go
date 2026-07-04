package startup

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"time"

	"bgscan/internal/ui/theme"

	"charm.land/lipgloss/v2"
)

const fastboot = false

// Style definitions for startup messages.
var (
	styleInfo = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00A7F7")).
			UnsetWidth().
			UnsetPadding().
			UnsetMargins()

	styleSuccess = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#32CD32")).
			UnsetWidth().
			UnsetPadding().
			UnsetMargins()

	styleWarning = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD700")).
			UnsetWidth().
			UnsetPadding().
			UnsetMargins()

	styleError = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555")).
			UnsetWidth().
			UnsetPadding().
			UnsetMargins()
)

// RunHealthChecks executes all startup validation steps in sequence.
//
// The following health checks are performed:
//  1. Logger initialization
//  2. Global theme setup
//  3. Configuration validation
//  4. Xray binary and template validation
//  5. DNSTT binary validation
//  6. Slipstream client validation
//
// Between steps, a short pause is introduced for improved UI readability.
// On completion, the user is prompted to press Enter.
func RunHealthChecks() {
	checkLoggerHealth()
	pause(500 * time.Millisecond)
	fmt.Println()

	theme.Init()

	checkConfigHealth()
	pause(500 * time.Millisecond)
	fmt.Println()

	checkXrayHealth()
	pause(500 * time.Millisecond)
	fmt.Println()

	checkDNSTTHealth()
	pause(500 * time.Millisecond)
	fmt.Println()

	checkSlipstreamHealth()
	pause(500 * time.Millisecond)
	fmt.Println()

	success("[SYSTEM] All startup checks completed ✅")
	fmt.Println()

	if err := pressEnterToContinue(); err != nil {
		log.Printf("failed to wait for Enter: %v", err)
	}
}

// binaryMissing logs a warning sequence indicating that a required binary
// could not be located. The component depending on it will be disabled.
func binaryMissing(name, binary string) {
	warn(fmt.Sprintf("[WARNING] %s scanner disabled.", name))
	warn(fmt.Sprintf("[WARNING] Could not find %s binary.", name))
	warn(fmt.Sprintf("[HINT] Ensure %s exists in assets/ or PATH.", binary))
}

// info logs an informational message using the info style.
func info(msg string) {
	fmt.Println(styleInfo.Render(msg))
	pause(100 * time.Millisecond)
}

// success logs a success message using the success style.
func success(msg string) {
	fmt.Println(styleSuccess.Render(msg))
	pause(100 * time.Millisecond)
}

// warn logs a warning message using the warning style.
func warn(msg string) {
	fmt.Println(styleWarning.Render(msg))
	pause(100 * time.Millisecond)
}

// errMsg logs an error message with its title and wrapped error.
func errMsg(title string, err error) {
	fmt.Println(styleError.Render(fmt.Sprintf("[ERROR] %s: %v", title, err)))
	pause(100 * time.Millisecond)
}

// infof formats and logs an informational message.
func infof(format string, a ...any) {
	info(fmt.Sprintf(format, a...))
}

// warnf formats and logs a warning message.
func warnf(format string, a ...any) {
	warn(fmt.Sprintf(format, a...))
}

// successf formats and logs a success message.
func successf(format string, a ...any) {
	success(fmt.Sprintf(format, a...))
}

// pressEnterToContinue pauses execution until the user presses Enter.
// Used for interactive CLI workflows.
func pressEnterToContinue() error {
	fmt.Print("Press Enter to continue...")
	_, err := bufio.NewReader(os.Stdin).ReadBytes('\n')
	return err
}

// pause adds a delay unless fastboot mode is enabled.
// Used to improve UI readability during sequential startup checks.
func pause(duration time.Duration) {
	if fastboot {
		return
	}
	time.Sleep(duration)
}
