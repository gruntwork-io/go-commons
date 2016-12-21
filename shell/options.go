package shell

import (
	"github.com/Sirupsen/logrus"
	"github.com/gruntwork-io/gruntwork-cli/logging"
)

type ShellOptions struct {
	NonInteractive bool
	Logger         *logrus.Entry
	WorkingDir     string
}

func NewShellOptions() *ShellOptions {
	return &ShellOptions{
		NonInteractive: false,
		Logger: logging.GetLogger(""),
		WorkingDir: ".",
	}
}
