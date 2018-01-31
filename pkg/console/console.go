package console

import (
	"fmt"
	"os"
	"sync"

	"path/filepath"
)

var (
	// DebugPrint enables/disables console debug printing.
	DebugPrint = false

	// Used by the caller to print multiple lines atomically. Exposed by Lock/Unlock methods.
	publicMutex = &sync.Mutex{}
	// Used internally by console.
	privateMutex = &sync.Mutex{}

	// Print prints a message.
	Print = func(data ...interface{}) {
		consolePrint("Print", data...)
		return
	}

	// PrintC prints a message.
	PrintC = func(data ...interface{}) {
		consolePrint("PrintC", data...)
		return
	}

	// Printf prints a formatted message.
	Printf = func(format string, data ...interface{}) {
		consolePrintf("Print", format, data...)
		return
	}

	// Println prints a message with a newline.
	Println = func(data ...interface{}) {
		consolePrintln("Print", data...)
		return
	}

	// Fatal print a error message and exit.
	Fatal = func(data ...interface{}) {
		consolePrint("Fatal", data...)
		os.Exit(1)
		return
	}

	// Fatalf print a error message with a format specified and exit.
	Fatalf = func(format string, data ...interface{}) {
		consolePrintf("Fatal", format, data...)
		os.Exit(1)
		return
	}

	// Fatalln print a error message with a new line and exit.
	Fatalln = func(data ...interface{}) {
		consolePrintln("Fatal", data...)
		os.Exit(1)
		return
	}

	// Error prints a error message.
	Error = func(data ...interface{}) {
		consolePrint("Error", data...)
		return
	}

	// Errorf print a error message with a format specified.
	Errorf = func(format string, data ...interface{}) {
		consolePrintf("Error", format, data...)
		return
	}

	// Errorln prints a error message with a new line.
	Errorln = func(data ...interface{}) {
		consolePrintln("Error", data...)
		return
	}

	// Info prints a informational message.
	Info = func(data ...interface{}) {
		consolePrint("Info", data...)
		return
	}

	// Infof prints a informational message in custom format.
	Infof = func(format string, data ...interface{}) {
		consolePrintf("Info", format, data...)
		return
	}

	// Infoln prints a informational message with a new line.
	Infoln = func(data ...interface{}) {
		consolePrintln("Info", data...)
		return
	}

	// Debug prints a debug message without a new line
	// Debug prints a debug message.
	Debug = func(data ...interface{}) {
		if DebugPrint {
			consolePrint("Debug", data...)
		}
	}

	// Debugf prints a debug message with a new line.
	Debugf = func(format string, data ...interface{}) {
		if DebugPrint {
			consolePrintf("Debug", format, data...)
		}
	}

	// Debugln prints a debug message with a new line.
	Debugln = func(data ...interface{}) {
		if DebugPrint {
			consolePrintln("Debug", data...)
		}
	}

	// Eraseline Print in new line and adjust to top so that we don't print over the ongoing progress bar.
	Eraseline = func() {
		consolePrintf("Print", "%c[2K\n", 27)
		consolePrintf("Print", "%c[A", 27)
	}
)

// wrap around standard fmt functions.
// consolePrint prints a message prefixed with message type and program name.
func consolePrint(tag string, a ...interface{}) {
	privateMutex.Lock()
	defer privateMutex.Unlock()

	/* #nosec */
	switch tag {
	case "Debug":
		// if no arguments are given do not invoke debug printer.
		if len(a) == 0 {
			return
		}
		fmt.Fprint(os.Stderr, ProgramName()+": <DEBUG> ")
		fmt.Fprint(os.Stderr, a...)
	case "Fatal":
		fallthrough
	case "Error":
		// if no arguments are given do not invoke fatal and error printer.
		if len(a) == 0 {
			return
		}
		fmt.Fprint(os.Stderr, ProgramName()+": <ERROR> ")
		fmt.Fprint(os.Stderr, a...)
	case "Info":
		// if no arguments are given do not invoke info printer.
		if len(a) == 0 {
			return
		}
		fmt.Fprint(os.Stdout, ProgramName()+": ")
		fmt.Fprint(os.Stdout, a...)
	default:
		fmt.Fprint(os.Stdout, a...)
	}
}

// consolePrintf - same as print with a new line.
func consolePrintf(tag string, format string, a ...interface{}) {
	privateMutex.Lock()
	defer privateMutex.Unlock()

	/* #nosec */
	switch tag {
	case "Debug":
		// if no arguments are given do not invoke debug printer.
		if len(a) == 0 {
			return
		}
		fmt.Fprint(os.Stderr, ProgramName()+": <DEBUG> ")
		fmt.Fprintf(os.Stderr, format, a...)
	case "Fatal":
		fallthrough
	case "Error":
		// if no arguments are given do not invoke fatal and error printer.
		if len(a) == 0 {
			return
		}
		fmt.Fprint(os.Stderr, ProgramName()+": <ERROR> ")
		fmt.Fprintf(os.Stderr, format, a...)
	case "Info":
		// if no arguments are given do not invoke info printer.
		if len(a) == 0 {
			return
		}
		fmt.Fprint(os.Stdout, ProgramName()+": ")
		fmt.Fprintf(os.Stdout, format, a...)
	default:
		fmt.Fprintf(os.Stdout, format, a...)
	}
}

// consolePrintln - same as print with a new line.
func consolePrintln(tag string, a ...interface{}) {
	privateMutex.Lock()
	defer privateMutex.Unlock()

	/* #nosec */
	switch tag {
	case "Debug":
		// if no arguments are given do not invoke debug printer.
		if len(a) == 0 {
			return
		}
		fmt.Fprint(os.Stderr, ProgramName()+": <DEBUG> ")
		fmt.Fprintln(os.Stderr, a...)
	case "Fatal":
		fallthrough
	case "Error":
		// if no arguments are given do not invoke fatal and error printer.
		if len(a) == 0 {
			return
		}
		fmt.Fprint(os.Stderr, ProgramName()+": <ERROR> ")
		fmt.Fprintln(os.Stderr, a...)
	case "Info":
		// if no arguments are given do not invoke info printer.
		if len(a) == 0 {
			return
		}
		fmt.Fprint(os.Stdout, ProgramName()+": ")
		fmt.Fprintln(os.Stdout, a...)
	default:
		fmt.Fprintln(os.Stdout, a...)
	}
}

// Lock console.
func Lock() {
	publicMutex.Lock()
}

// Unlock locked console.
func Unlock() {
	publicMutex.Unlock()
}

// ProgramName - return the name of the executable program.
func ProgramName() string {
	_, progName := filepath.Split(os.Args[0])
	return progName
}
