package shell

import (
	"bytes"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/go-commons/logging"
	"github.com/stretchr/testify/assert"
)

func TestRunShellCommand(t *testing.T) {
	t.Parallel()

	assert.NoError(t, RunShellCommand(NewShellOptions(), "echo", "hi"))
}

func TestRunShellCommandInvalidCommand(t *testing.T) {
	t.Parallel()

	assert.Error(t, RunShellCommand(NewShellOptions(), "not-a-real-command"))
}

func TestRunShellCommandAndGetOutput(t *testing.T) {
	t.Parallel()

	out, err := RunShellCommandAndGetOutput(NewShellOptions(), "echo", "hi")
	assert.NoError(t, err)
	assert.Equal(t, "hi\n", out)
}

func TestRunShellCommandAndGetStdoutReturnsStdout(t *testing.T) {
	t.Parallel()

	out, err := RunShellCommandAndGetStdout(NewShellOptions(), "echo", "hi")
	assert.NoError(t, err)
	assert.Equal(t, "hi\n", out)
}

func TestRunShellCommandAndGetStdoutDoesNotReturnStderr(t *testing.T) {
	t.Parallel()

	out, err := RunShellCommandAndGetStdout(NewShellOptions(), filepath.Join("test-fixture", "echo_hi_stderr.sh"))
	assert.NoError(t, err)
	assert.Equal(t, "", out)
}

func TestRunShellCommandWithEnv(t *testing.T) {
	t.Parallel()

	envVars := map[string]string{
		"TEST_WITH_SPACES":  "test with spaces",
		"TEST_WITH_EQUALS":  "test=with=equals",
		"TEST_START_EQUALS": "=teststartequals",
		"TEST_BLANK":        "",
	}
	options := NewShellOptions()
	options.Env = envVars

	for k, v := range envVars {
		out, err := RunShellCommandAndGetOutput(options, "bash", "-c", fmt.Sprintf("echo $%s", k))
		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("%s\n", v), out)
	}
}

func TestCommandInstalledOnValidCommand(t *testing.T) {
	t.Parallel()

	assert.True(t, CommandInstalled("echo"))
}

func TestCommandInstalledOnInvalidCommand(t *testing.T) {
	t.Parallel()

	assert.False(t, CommandInstalled("not-a-real-command"))
}

func TestCommandInstalledEOnValidCommand(t *testing.T) {
	t.Parallel()

	assert.NoError(t, CommandInstalledE("echo"))
}

func TestCommandInstalledEOnInvalidCommand(t *testing.T) {
	t.Parallel()

	assert.Error(t, CommandInstalledE("not-a-real-command"))
}

// Test that when SensitiveArgs is true, do not log the args
func TestSensitiveArgsTrueHidesOnRunShellCommand(t *testing.T) {
	t.Parallel()

	buffer := bytes.NewBufferString("")
	logger := logging.GetLogger("")
	logger.Out = buffer
	options := NewShellOptions()
	options.SensitiveArgs = true
	options.Logger = logger

	assert.NoError(t, RunShellCommand(options, "echo", "hi"))
	assert.NotContains(t, buffer.String(), "hi")
	assert.Contains(t, buffer.String(), "echo")
}

// Test that when SensitiveArgs is false, log the args
func TestSensitiveArgsFalseShowsOnRunShellCommand(t *testing.T) {
	t.Parallel()

	buffer := bytes.NewBufferString("")
	logger := logging.GetLogger("")
	logger.Out = buffer
	options := NewShellOptions()
	options.Logger = logger

	assert.NoError(t, RunShellCommand(options, "echo", "hi"))
	assert.Contains(t, buffer.String(), "hi")
	assert.Contains(t, buffer.String(), "echo")
}

// Test that when SensitiveArgs is true, do not log the args
func TestSensitiveArgsTrueHidesOnRunShellCommandAndGetOutput(t *testing.T) {
	t.Parallel()

	buffer := bytes.NewBufferString("")
	logger := logging.GetLogger("")
	logger.Out = buffer
	options := NewShellOptions()
	options.SensitiveArgs = true
	options.Logger = logger

	_, err := RunShellCommandAndGetOutput(options, "echo", "hi")
	assert.NoError(t, err)
	assert.NotContains(t, buffer.String(), "hi")
	assert.Contains(t, buffer.String(), "echo")
}

// Test that when SensitiveArgs is false, log the args
func TestSensitiveArgsFalseShowsOnRunShellCommandAndGetOutput(t *testing.T) {
	t.Parallel()

	buffer := bytes.NewBufferString("")
	logger := logging.GetLogger("")
	logger.Out = buffer
	options := NewShellOptions()
	options.Logger = logger

	_, err := RunShellCommandAndGetOutput(options, "echo", "hi")
	assert.NoError(t, err)
	assert.Contains(t, buffer.String(), "hi")
	assert.Contains(t, buffer.String(), "echo")
}
