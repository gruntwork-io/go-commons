package shell

import (
	"github.com/sirupsen/logrus"
	"github.com/gruntwork-io/gruntwork-cli/logging"
)

type ShellOptions struct {
	NonInteractive bool
	Logger         *logrus.Logger
	WorkingDir     string
}

func NewShellOptions() *ShellOptions {
	return &ShellOptions{
		NonInteractive: false,
		Logger: logging.GetLogger(""),
		WorkingDir: ".",
	}
}
