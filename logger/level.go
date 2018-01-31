package logger

import (
	"io"
	"log"
	"time"
)

// Level is an enumeration of log levels.
type Level int

// Possible values for the Level enum.
const (
	_ Level = iota
	LevelTrace
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
	LevelCritical
	LevelFatal
)

func (t Level) String() string {
	return prefixes[t]
}

var prefixes = map[Level]string{
	LevelTrace:    "TRACE",
	LevelDebug:    "DEBUG",
	LevelInfo:     "INFO",
	LevelWarn:     "WARN",
	LevelError:    "ERROR",
	LevelCritical: "CRITICAL",
	LevelFatal:    "FATAL",
}

var infoThresholdByteSlice = []byte(LevelInfo.String())

// LevelFilter is an io.Writer that can be used with a logger that will filter
// out log messages that aren't at least a certain threshold.
type LevelFilter struct {
	// logThreshold is the minimum level allowed through.
	logThreshold Level

	// The underlying io.Writer where log messages that pass the filter will be
	// send to.
	writer io.Writer

	validLevels map[Level]struct{}
}

// NewLevelFilter creates a new LevelFilter.
func NewLevelFilter(logThreshold Level) *LevelFilter {
	f := &LevelFilter{}

	f.logThreshold = logThreshold

	f.init()
	return f
}

func (f *LevelFilter) init() {
	log.SetFlags(0)

	validLevels := make(map[Level]struct{})
	for level := range prefixes {
		if level < f.logThreshold {
			continue
		}
		validLevels[level] = struct{}{}
	}
	f.validLevels = validLevels
}

// SetLevel changes the threshold above which messages are written to the log
// writer.
func (f *LevelFilter) SetLevel(threshold Level) {
	f.logThreshold = threshold
	f.init()
}

// SetLogOutput changes the file where log messages are written.
func (f *LevelFilter) SetLogOutput(handle io.Writer) {
	f.writer = handle
	f.init()
}

// GetLevel returns the defined Level for the logger.
func (f *LevelFilter) GetLevel() Level {
	return f.logThreshold
}

// Check checks whether a given line if it should be included in the level
// filter.
func (f *LevelFilter) Check(line []byte) (hasThreshold, isValid bool) {
	// extract threshold
	var threshold Level
	threshold, hasThreshold = getLevel(line)
	if !hasThreshold {
		// line without threshold will default to INFO level
		threshold = LevelInfo
	}
	_, isValid = f.validLevels[threshold]

	return
}

func (f *LevelFilter) Write(p []byte) (int, error) {
	// Note in general that io.Writer can receive any byte sequence to write, but
	// the "log" package always guarantees that we only get a single line. We use
	// that as a slight optimization within this method, assuming we're dealing
	// with a single, complete line of log data.

	tok, valid := f.Check(p)

	if !valid {
		return len(p), nil
	}

	line := []byte(time.Now().UTC().Format(time.RFC3339))
	line = append(line, ' ')
	if !tok {
		line = append(line, infoThresholdByteSlice...)
		line = append(line, ' ')
	}
	line = append(line, p...)

	return f.writer.Write(line)
}

// getLevel returns appropriate Level for the given line.
//
// It checks bytes using custom switch.
func getLevel(line []byte) (Level, bool) {
	l := len(line)
	if l > 3 {
		switch line[0] {
		case 'T':
			if l > 4 && line[4] == 'E' {
				return LevelTrace, true
			}
			break
		case 'D':
			if l > 4 && line[4] == 'G' {
				return LevelDebug, true
			}
			break
		case 'I':
			if line[3] == 'O' {
				return LevelInfo, true
			}
			break
		case 'W':
			if line[3] == 'N' {
				return LevelWarn, true
			}
			break
		case 'E':
			if l > 4 && line[4] == 'R' {
				return LevelError, true
			}
			break
		case 'C':
			if l > 7 && line[7] == 'L' {
				return LevelCritical, true
			}
			break
		case 'F':
			if l > 4 && line[4] == 'L' {
				return LevelFatal, true
			}
			break
		}
	}
	return LevelTrace, false
}
