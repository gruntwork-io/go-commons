package shell

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/gruntwork-io/go-commons/errors"
	"github.com/sirupsen/logrus"
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

// Run the specified shell command with the specified arguments. Return its stdout, stderr, and interleaved output as
// separate strings in a struct.
func RunShellCommandAndGetOutputStruct(options *ShellOptions, command string, args ...string) (*Output, error) {
	return runShellCommand(options, false, command, args...)
}

// Run the specified shell command with the specified arguments. Return its stdout and stderr as a string
func RunShellCommandAndGetOutput(options *ShellOptions, command string, args ...string) (string, error) {
	out, err := runShellCommand(options, false, command, args...)
	return out.Combined(), err
}

// Run the specified shell command with the specified arguments. Return its interleaved stdout and stderr as a string
// and also stream stdout and stderr to the OS stdout/stderr
func RunShellCommandAndGetAndStreamOutput(options *ShellOptions, command string, args ...string) (string, error) {
	out, err := runShellCommand(options, true, command, args...)
	return out.Combined(), err
}

// Run the specified shell command with the specified arguments. Return its stdout as a string
func RunShellCommandAndGetStdout(options *ShellOptions, command string, args ...string) (string, error) {
	out, err := runShellCommand(options, false, command, args...)
	return out.Stdout(), err
}

// Run the specified shell command with the specified arguments. Return its stdout as a string and also stream stdout
// and stderr to the OS stdout/stderr
func RunShellCommandAndGetStdoutAndStreamOutput(options *ShellOptions, command string, args ...string) (string, error) {
	out, err := runShellCommand(options, true, command, args...)
	return out.Stdout(), err
}

// Run the specified shell command with the specified arguments. Return its stdout, stderr, and interleaved output as a
// struct and also stream stdout and stderr to the OS stdout/stderr
func RunShellCommandAndGetOutputStructAndStreamOutput(options *ShellOptions, command string, args ...string) (*Output, error) {
	return runShellCommand(options, true, command, args...)
}

// Run the specified shell command with the specified arguments. Return its stdout and stderr as a string and also
// stream stdout and stderr to the OS stdout/stderr
func runShellCommand(options *ShellOptions, streamOutput bool, command string, args ...string) (*Output, error) {
	logCommand(options, command, args...)
	cmd := exec.Command(command, args...)

	setCommandOptions(options, cmd)

	cmd.Stdin = os.Stdin

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, errors.WithStackTrace(err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, errors.WithStackTrace(err)
	}

	if err := cmd.Start(); err != nil {
		return nil, errors.WithStackTrace(err)
	}

	output, err := readStdoutAndStderr(
		options.Logger,
		streamOutput,
		stdout,
		stderr,
	)
	if err != nil {
		return output, err
	}

	err = cmd.Wait()
	return output, errors.WithStackTrace(err)
}

// This function captures stdout and stderr into the given variables while still printing it to the stdout and stderr
// of this Go program
func readStdoutAndStderr(log *logrus.Logger, streamOutput bool, stdout, stderr io.ReadCloser) (*Output, error) {
	out := newOutput()
	stdoutReader := bufio.NewReader(stdout)
	stderrReader := bufio.NewReader(stderr)

	wg := &sync.WaitGroup{}

	wg.Add(2)
	var stdoutErr, stderrErr error
	go func() {
		defer wg.Done()
		stdoutErr = readData(log, streamOutput, stdoutReader, out.stdout)
	}()
	go func() {
		defer wg.Done()
		stderrErr = readData(log, streamOutput, stderrReader, out.stderr)
	}()
	wg.Wait()

	if stdoutErr != nil {
		return out, stdoutErr
	}
	if stderrErr != nil {
		return out, stderrErr
	}

	return out, nil
}

func readData(log *logrus.Logger, streamOutput bool, reader *bufio.Reader, writer io.StringWriter) error {
	var line string
	var readErr error
	for {
		line, readErr = reader.ReadString('\n')

		// only return early if the line does not have
		// any contents. We could have a line that does
		// not not have a newline before io.EOF, we still
		// need to add it to the output.
		if len(line) == 0 && readErr == io.EOF {
			break
		}

		if streamOutput {
			log.Println(line)
		}
		if _, err := writer.WriteString(line); err != nil {
			return err
		}

		if readErr != nil {
			break
		}
	}
	if readErr != io.EOF {
		return readErr
	}
	return nil
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
