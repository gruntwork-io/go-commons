package entrypoint

import (
	"os"
	"github.com/urfave/cli"
	"github.com/gruntwork-io/gruntwork-cli/errors"
	"github.com/gruntwork-io/gruntwork-cli/logging"
)

const defaultSuccessExitCode = 0
const defaultErrorExitCode = 1

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
func checkForErrorsAndExit(err error) {
	exitCode := defaultSuccessExitCode

	if err != nil {
		logging.GetLogger("").WithError(err).Error(errors.PrintErrorWithStackTrace(err))

		errorWithExitCode, isErrorWithExitCode := err.(errors.ErrorWithExitCode)
		if isErrorWithExitCode {
			exitCode = errorWithExitCode.ExitCode
		} else {
			exitCode = defaultErrorExitCode
		}
	}

	os.Exit(exitCode)
}