package shell

import (
	"fmt"
	"bufio"
	"os"
	"github.com/gruntwork-io/gruntwork-cli/errors"
	"strings"
	"github.com/gruntwork-io/gruntwork-cli/logging"
	"github.com/fatih/color"
)

var BRIGHT_GREEN = color.New(color.FgHiGreen, color.Bold)

// Prompt the user for text in the CLI. Returns the text entered by the user.
func PromptUserForInput(prompt string, options *ShellOptions) (string, error) {
	loggerName := logging.GetLoggerName(options.Logger)
	if loggerName != "" {
		prompt = fmt.Sprintf("%s %s", loggerName, prompt)
	}
	BRIGHT_GREEN.Print(prompt)

	if options.NonInteractive {
		fmt.Println()
		options.Logger.Info("The non-interactive flag is set to true, so assuming 'yes' for all prompts")
		return "yes", nil
	}

	reader := bufio.NewReader(os.Stdin)

	text, err := reader.ReadString('\n')
	if err != nil {
		return "", errors.WithStackTrace(err)
	}

	return strings.TrimSpace(text), nil
}

// Prompt the user for a yes/no response and return true if they entered yes.
func PromptUserForYesNo(prompt string, options *ShellOptions) (bool, error) {
	resp, err := PromptUserForInput(fmt.Sprintf("%s (y/n) ", prompt), options)

	if err != nil {
		return false, errors.WithStackTrace(err)
	}

	switch strings.ToLower(resp) {
	case "y", "yes": return true, nil
	default: return false, nil
	}
}

