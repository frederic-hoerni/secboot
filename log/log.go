package log

import (
	"fmt"
	"os"
)

var active bool = false

func Activate() {
	active = true
}

func Info(format string, v ...any) {
	if !active {
		return
	}
	fmt.Printf(format, v...)
	fmt.Printf("\n")
}

func Error(format string, v ...any) {
	if !active {
		return
	}
	fmt.Fprintf(os.Stderr, "error: ")
	fmt.Fprintf(os.Stderr, format, v...)
	fmt.Fprintf(os.Stderr, "\n")
}
