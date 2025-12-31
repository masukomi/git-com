package output

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

var (
	redPrinter    = color.New(color.FgRed)
	yellowPrinter = color.New(color.FgYellow)
)

// PrintError prints an error message in red to stderr
func PrintError(msg string) {
	redPrinter.Fprintln(os.Stderr, msg)
}

// PrintWarning prints a warning message in yellow to stdout
func PrintWarning(msg string) {
	yellowPrinter.Println(msg)
}

// PrintWarningToStderr prints a warning message in yellow to stderr
func PrintWarningToStderr(msg string) {
	yellowPrinter.Fprintln(os.Stderr, msg)
}

// Print prints a message to stdout without coloring
func Print(msg string) {
	fmt.Println(msg)
}
