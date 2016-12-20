package shell

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestRunShellCommand(t *testing.T) {
	t.Parallel()

	assert.Nil(t, RunShellCommand(NewShellOptions(), "echo", "hi"))
}

func TestRunShellCommandInvalidCommand(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, RunShellCommand(NewShellOptions(), "not-a-real-command"))
}

func TestRunShellCommandAndGetOutput(t *testing.T) {
	t.Parallel()

	out, err := RunShellCommandAndGetOutput(NewShellOptions(), "echo", "hi")
	assert.Nil(t, err, "Unexpected error: %v", err)
	assert.Equal(t, "hi\n", out)
}

func TestCommandInstalledOnValidCommand(t *testing.T) {
	t.Parallel()

	assert.True(t, CommandInstalled("echo"))
}

func TestCommandInstalledOnInvalidCommand(t *testing.T) {
	t.Parallel()

	assert.False(t, CommandInstalled("not-a-real-command"))
}
