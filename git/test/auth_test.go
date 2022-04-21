//go:build gittest

package gittest

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/go-commons/files"
	"github.com/gruntwork-io/go-commons/git"
	"github.com/gruntwork-io/go-commons/logging"
	"github.com/gruntwork-io/terratest/modules/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	gitPATEnvName = "GITHUB_OAUTH_TOKEN"
)

var (
	logger = logging.GetLogger("testlogger")
)

// NOTE: All these tests should be run in the provided docker environment to avoid polluting the local git configuration
// settings. The tests will assert that it is running in the docker environment, and will fail if it is not.

func TestHTTPSAuth(t *testing.T) {
	t.Parallel()

	currentDir, err := os.Getwd()
	require.NoError(t, err)
	require.Equal(t, "/workspace/go-commons/git/test", currentDir)

	environment.RequireEnvVar(t, gitPATEnvName)
	gitPAT := os.Getenv(gitPATEnvName)
	require.NoError(t, git.ConfigureHTTPSAuth(logger, "git", gitPAT, "github.com"))

	tmpDir, err := ioutil.TempDir("", "git-test")
	require.NoError(t, err)
	require.NoError(t, git.Clone(logger, "https://github.com/gruntwork-io/terraform-aws-lambda.git", tmpDir))
	assert.True(t, files.IsDir(filepath.Join(tmpDir, "modules/lambda")))
}

func TestForceHTTPS(t *testing.T) {
	t.Parallel()

	currentDir, err := os.Getwd()
	require.NoError(t, err)
	require.Equal(t, "/workspace/go-commons/git/test", currentDir)

	environment.RequireEnvVar(t, gitPATEnvName)
	gitPAT := os.Getenv(gitPATEnvName)
	require.NoError(t, git.ConfigureHTTPSAuth(logger, "git", gitPAT, "github.com"))
	require.NoError(t, git.ConfigureForceHTTPS(logger))

	tmpDir, err := ioutil.TempDir("", "git-test")
	require.NoError(t, err)
	require.NoError(t, git.Clone(logger, "git@github.com:gruntwork-io/terraform-aws-lambda.git", tmpDir))
	assert.True(t, files.IsDir(filepath.Join(tmpDir, "modules/lambda")))
}
