package shell

import (
	"os/exec"
	"os"
	"strings"
	"github.com/gruntwork-io/gruntwork-cli/errors"
)

// Run the specified shell command with the specified arguments. Connect the command's stdin, stdout, and stderr to
// the currently running app.
func RunShellCommand(options *ShellOptions, command string, args ... string) error {
	options.Logger.Info("Running command: %s %s", command, strings.Join(args, " "))

	cmd := exec.Command(command, args...)

	// TODO: consider logging this via options.Logger
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Dir = options.WorkingDir

	return errors.WithStackTrace(cmd.Run())
}

// Run the specified shell command with the specified arguments. Return its stdout and stderr as a string
func RunShellCommandAndGetOutput(options *ShellOptions, command string, args ... string) (string, error) {
	options.Logger.Info("Running command: %s %s", command, strings.Join(args, " "))

	cmd := exec.Command(command, args...)

	cmd.Stdin = os.Stdin
	cmd.Dir = options.WorkingDir

	out, err := cmd.Output()
	if err != nil {
		return "", errors.WithStackTrace(err)
	}
	return string(out), nil
}

// Return true if the OS has the given command installed
func CommandInstalled(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}
