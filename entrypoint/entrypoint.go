package entrypoint

import (
	"fmt"
	"os"

	"github.com/urfave/cli"

	"github.com/gruntwork-io/go-commons/errors"
	"github.com/gruntwork-io/go-commons/logging"
)

const defaultSuccessExitCode = 0
const defaultErrorExitCode = 1
const debugEnvironmentVarName = "GRUNTWORK_DEBUG"

// Wrapper around cli.NewApp that sets the help text printer.
func NewApp() *cli.App {
	cli.HelpPrinter = WrappedHelpPrinter
	cli.AppHelpTemplate = CLI_APP_HELP_TEMPLATE
	cli.CommandHelpTemplate = CLI_COMMAND_HELP_TEMPLATE
	cli.SubcommandHelpTemplate = CLI_APP_HELP_TEMPLATE
	app := cli.NewApp()
	return app
}

// Run the given app, handling errors, panics, and stack traces where possible
func RunApp(app *cli.App) {
	cli.OsExiter = func(exitCode int) {
		// Do nothing. We just need to override this function, as the default value calls os.Exit, which
		// kills the app (or any automated test) dead in its tracks.
	}

	defer errors.Recover(checkForErrorsAndExit)
	err := app.Run(os.Args)
	checkForErrorsAndExit(err)
}

// If there is an error, display it in the console and exit with a non-zero exit code. Otherwise, exit 0.
// Note that if the GRUNTWORK_DEBUG environment variable is set, this will print out the stack trace.
func checkForErrorsAndExit(err error) {
	logError(err)
	exitCode := getExitCode(err)
	os.Exit(exitCode)
}

// logError will output an error message to stderr. This will output the stack trace if we are in debug mode.
func logError(err error) {
	isDebugMode := os.Getenv(debugEnvironmentVarName) != ""
	if err != nil {
		errWithoutStackTrace := errors.Unwrap(err)
		if isDebugMode {
			logging.GetLogger("").WithError(err).Error(errors.PrintErrorWithStackTrace(err))
		} else {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", errWithoutStackTrace)
		}
	}
}

// getExitCode will return an exit code to use for the CLI app. This will either be:
// - defaultSuccessExitCode if there is no error.
// - defaultErrorExitCode if there is a standard error.
// - error exit code if there is an error that indicates an exit code.
func getExitCode(err error) int {
	exitCode := defaultSuccessExitCode
	if err != nil {
		errWithoutStackTrace := errors.Unwrap(err)
		errorWithExitCode, isErrorWithExitCode := errWithoutStackTrace.(errors.ErrorWithExitCode)
		if isErrorWithExitCode {
			exitCode = errorWithExitCode.ExitCode
		} else {
			exitCode = defaultErrorExitCode
		}
	}
	return exitCode
}
