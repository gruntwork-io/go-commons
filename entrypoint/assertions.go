package entrypoint

import (
	"fmt"
	"github.com/urfave/cli"
	"os"
)

// StringFlagRequiredE checks if a required string flag is passed in on the CLI. This will return the set string, or an
// error if the flag is not passed in.
func StringFlagRequiredE(cliContext *cli.Context, flagName string) (string, error) {
	value := cliContext.String(flagName)
	if value == "" {
		return "", fmt.Errorf("--%s is required", flagName)
	}
	return value, nil
}

// EnvironmentVarRequiredE checks if a required environment variable is set. This will return the environment variable
// value, or an error if the environment variable is not set.
func EnvironmentVarRequiredE(varName string) (string, error) {
	value := os.Getenv(varName)
	if value == "" {
		return "", fmt.Errorf("The environment variable %s is required to be set", varName)
	}
	return value, nil
}
