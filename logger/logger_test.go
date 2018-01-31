package logger

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func resetEnv(t *testing.T) {
	t.Helper()
	logLevelFromEnv = false
	logFileFromEnv = false
}

func TestWriteLogToFile(t *testing.T) {
	resetEnv(t)

	tmpfile, err := ioutil.TempFile("", "")
	assert.NoError(t, err)
	defer func() {
		os.Remove(tmpfile.Name())
	}()

	SetupLogging(false, false, tmpfile.Name())
	log.Printf("INFO TEST")
	log.Printf("DEBUG TEST") // <- should be ignored

	f, err := ioutil.ReadFile(tmpfile.Name())
	assert.NoError(t, err)
	assert.Equal(t, []byte("Z INFO TEST\n"), f[19:])
}

func TestDebugWriteLogToFile(t *testing.T) {
	resetEnv(t)

	tmpfile, err := ioutil.TempFile("", "")
	assert.NoError(t, err)
	defer func() {
		os.Remove(tmpfile.Name())
	}()

	SetupLogging(true, false, tmpfile.Name())
	log.Printf("DEBUG TEST")

	f, err := ioutil.ReadFile(tmpfile.Name())
	assert.NoError(t, err)
	assert.Equal(t, []byte("Z DEBUG TEST\n"), f[19:])
}

func TestErrorWriteLogToFile(t *testing.T) {
	resetEnv(t)

	tmpfile, err := ioutil.TempFile("", "")
	assert.NoError(t, err)
	defer func() {
		os.Remove(tmpfile.Name())
	}()

	SetupLogging(false, true, tmpfile.Name())
	log.Printf("ERROR TEST")
	log.Printf("INFO TEST") // <- should be ignored

	f, err := ioutil.ReadFile(tmpfile.Name())
	assert.NoError(t, err)
	assert.Equal(t, []byte("Z ERROR TEST\n"), f[19:])
}

func TestAddDefaultLogLevel(t *testing.T) {
	resetEnv(t)

	tmpfile, err := ioutil.TempFile("", "")
	assert.NoError(t, err)
	defer func() {
		os.Remove(tmpfile.Name())
	}()

	SetupLogging(true, false, tmpfile.Name())
	log.Printf("TEST")

	f, err := ioutil.ReadFile(tmpfile.Name())
	assert.NoError(t, err)
	assert.Equal(t, []byte("Z INFO TEST\n"), f[19:])
}
