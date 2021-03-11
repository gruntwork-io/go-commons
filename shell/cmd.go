package shell

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/gruntwork-io/go-commons/errors"
)

// Run the specified shell command with the specified arguments. Connect the command's stdin, stdout, and stderr to
// the currently running app.
func RunShellCommand(options *ShellOptions, command string, args ...string) error {
	logCommand(options, command, args...)
	cmd := exec.Command(command, args...)

	// TODO: consider logging this via options.Logger
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	setCommandOptions(options, cmd)

	return errors.WithStackTrace(cmd.Run())
}

// Run the specified shell command with the specified arguments. Return its stdout and stderr as a string
func RunShellCommandAndGetOutput(options *ShellOptions, command string, args ...string) (string, error) {
	logCommand(options, command, args...)
	cmd := exec.Command(command, args...)

	cmd.Stdin = os.Stdin

	setCommandOptions(options, cmd)

	out, err := cmd.CombinedOutput()
	return string(out), errors.WithStackTrace(err)
}

// Run the specified shell command with the specified arguments. Return its stdout and stderr as a string and also
// stream stdout and stderr to the OS stdout/stderr
func RunShellCommandAndGetAndStreamOutput(options *ShellOptions, command string, args ...string) (string, error) {
	logCommand(options, command, args...)
	cmd := exec.Command(command, args...)

	setCommandOptions(options, cmd)

	cmd.Stdin = os.Stdin

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", errors.WithStackTrace(err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", errors.WithStackTrace(err)
	}

	if err := cmd.Start(); err != nil {
		return "", errors.WithStackTrace(err)
	}

	output, err := readStdoutAndStderr(
		stdout,
		true,
		stderr,
		true,
		options,
	)
	if err != nil {
		return output, err
	}

	err = cmd.Wait()
	return output, errors.WithStackTrace(err)
}

// Run the specified shell command with the specified arguments. Return its stdout as a string
func RunShellCommandAndGetStdout(options *ShellOptions, command string, args ...string) (string, error) {
	logCommand(options, command, args...)
	cmd := exec.Command(command, args...)

	cmd.Stdin = os.Stdin

	setCommandOptions(options, cmd)

	out, err := cmd.Output()
	return string(out), errors.WithStackTrace(err)
}

// Run the specified shell command with the specified arguments. Return its stdout as a string and also stream stdout
// and stderr to the OS stdout/stderr
func RunShellCommandAndGetStdoutAndStreamOutput(options *ShellOptions, command string, args ...string) (string, error) {
	logCommand(options, command, args...)
	cmd := exec.Command(command, args...)

	setCommandOptions(options, cmd)

	cmd.Stdin = os.Stdin

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", errors.WithStackTrace(err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", errors.WithStackTrace(err)
	}

	if err := cmd.Start(); err != nil {
		return "", errors.WithStackTrace(err)
	}

	output, err := readStdoutAndStderr(
		stdout,
		true,
		stderr,
		false,
		options,
	)
	if err != nil {
		return output, err
	}

	err = cmd.Wait()
	return output, errors.WithStackTrace(err)
}

// This function captures stdout and stderr while still printing it to the stdout and stderr of this Go program
func readStdoutAndStderr(
	stdout io.ReadCloser,
	includeStdout bool,
	stderr io.ReadCloser,
	includeStderr bool,
	options *ShellOptions,
) (string, error) {
	allOutput := []string{}

	stdoutScanner := bufio.NewScanner(stdout)
	stderrScanner := bufio.NewScanner(stderr)

	for {
		if stdoutScanner.Scan() {
			text := stdoutScanner.Text()
			options.Logger.Println(text)
			if includeStdout {
				allOutput = append(allOutput, text)
			}
		} else if stderrScanner.Scan() {
			text := stderrScanner.Text()
			options.Logger.Println(text)
			if includeStderr {
				allOutput = append(allOutput, text)
			}
		} else {
			break
		}
	}

	if err := stdoutScanner.Err(); err != nil {
		return "", errors.WithStackTrace(err)
	}

	if err := stderrScanner.Err(); err != nil {
		return "", errors.WithStackTrace(err)
	}

	return strings.Join(allOutput, "\n"), nil
}

func logCommand(options *ShellOptions, command string, args ...string) {
	if options.SensitiveArgs {
		options.Logger.Infof("Running command: %s (args redacted)", command)
	} else {
		options.Logger.Infof("Running command: %s %s", command, strings.Join(args, " "))
	}
}

// Return true if the OS has the given command installed
func CommandInstalled(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

// CommandInstalledE returns an error if command is not installed
func CommandInstalledE(command string) error {
	if commandExists := CommandInstalled(command); !commandExists {
		err := fmt.Errorf("Command %s is not installed", command)
		return errors.WithStackTrace(err)
	}
	return nil
}

// setCommandOptions takes the shell options and maps them to the configurations for the exec.Cmd object, applying them
// to the passed in Cmd object.
func setCommandOptions(options *ShellOptions, cmd *exec.Cmd) {
	cmd.Dir = options.WorkingDir
	cmd.Env = formatEnvVars(options)
}

// formatEnvVars takes environment variables encoded into ShellOptions and converts them to a format understood by
// exec.Command
func formatEnvVars(options *ShellOptions) []string {
	env := os.Environ()
	for key, value := range options.Env {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}
	return env
}
