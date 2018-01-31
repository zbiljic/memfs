package logger

import (
	"bytes"
	"io"
	"log"
	"strings"
	"testing"
)

func TestLevelFilter_impl(t *testing.T) {
	var _ io.Writer = new(LevelFilter)
}

func TestLevelFilter(t *testing.T) {
	buf := new(bytes.Buffer)
	filter := NewLevelFilter(LevelWarn)
	filter.SetLogOutput(buf)

	logger := log.New(filter, "", 0)
	logger.Print("WARN foo")
	logger.Println("ERROR bar")
	logger.Println("DEBUG baz")
	logger.Println("WARN buzz")

	result := buf.String()
	// remove time from buffer
	resultLines := strings.Split(result, "\n")
	result = ""
	for _, line := range resultLines {
		if len(line) > 21 {
			result = result + line[21:]
			result = result + "\n"
		}
	}
	expected := "WARN foo\nERROR bar\nWARN buzz\n"
	if result != expected {
		t.Fatalf("bad: %#v", result)
	}
}

func TestLevelFilterCheck(t *testing.T) {
	filter := NewLevelFilter(LevelWarn)
	filter.SetLogOutput(nil)

	testCases := []struct {
		line  string
		check bool
	}{
		{"WARN foo\n", true},
		{"ERROR bar\n", true},
		{"DEBUG baz\n", false},
		{"WARN buzz\n", true},
	}

	for _, testCase := range testCases {
		_, result := filter.Check([]byte(testCase.line))
		if result != testCase.check {
			t.Errorf("Fail: %s", testCase.line)
		}
	}
}

func TestLevelFilter_SetLogThreshold(t *testing.T) {
	filter := NewLevelFilter(LevelError)
	filter.SetLogOutput(nil)

	testCases := []struct {
		line        string
		checkBefore bool
		checkAfter  bool
	}{
		{"WARN foo\n", false, true},
		{"ERROR bar\n", true, true},
		{"DEBUG baz\n", false, false},
		{"WARN buzz\n", false, true},
	}

	for _, testCase := range testCases {
		_, result := filter.Check([]byte(testCase.line))
		if result != testCase.checkBefore {
			t.Errorf("Fail: %s", testCase.line)
		}
	}

	// Update the minimum log threshold to WARN
	filter.SetLevel(LevelWarn)

	for _, testCase := range testCases {
		_, result := filter.Check([]byte(testCase.line))
		if result != testCase.checkAfter {
			t.Errorf("Fail: %s", testCase.line)
		}
	}
}
