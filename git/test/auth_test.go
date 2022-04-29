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
	ttlogger "github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/shell"
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
// All these tests are also run in serial to avoid race conditions on the git config file.

func TestHTTPSAuth(t *testing.T) {
	defer cleanupGitConfig(t)

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

func TestHTTPSAuthWithPath(t *testing.T) {
	defer cleanupGitConfig(t)

	currentDir, err := os.Getwd()
	require.NoError(t, err)
	require.Equal(t, "/workspace/go-commons/git/test", currentDir)

	environment.RequireEnvVar(t, gitPATEnvName)
	gitPAT := os.Getenv(gitPATEnvName)

	lambdaGitURL := "https://github.com/gruntwork-io/terraform-aws-lambda.git"
	lambdaOpts := git.CacheCredentialOptions{
		Host:            lambdaGitURL,
		DefaultUsername: "git",
		IncludeHTTPPath: true,
		SocketPath:      "",
		Timeout:         3600,
	}
	require.NoError(
		t,
		git.ConfigureCacheCredentialsHelper(logger, lambdaOpts),
	)
	require.NoError(
		t,
		git.StoreCacheCredentials(logger, "git", gitPAT, "github.com", "gruntwork-io/terraform-aws-lambda.git", ""),
	)

	tmpDir, err := ioutil.TempDir("", "git-test")
	require.NoError(t, err)
	require.NoError(t, git.Clone(logger, lambdaGitURL, tmpDir))
	assert.True(t, files.IsDir(filepath.Join(tmpDir, "modules/lambda")))
}

func TestHTTPSAuthMixed(t *testing.T) {
	defer cleanupGitConfig(t)

	currentDir, err := os.Getwd()
	require.NoError(t, err)
	require.Equal(t, "/workspace/go-commons/git/test", currentDir)

	// Make sure the directory for git credential sockets exist
	socketPath := "/tmp/git-credential-sockets"
	require.NoError(t, os.MkdirAll(socketPath, 0o700))
	lambdaSocketPath := filepath.Join(socketPath, "lambda")
	githubSocketPath := filepath.Join(socketPath, "github")

	environment.RequireEnvVar(t, gitPATEnvName)
	gitPAT := os.Getenv(gitPATEnvName)

	lambdaGitURL := "https://github.com/gruntwork-io/terraform-aws-lambda.git"
	lambdaOpts := git.CacheCredentialOptions{
		Host:            lambdaGitURL,
		DefaultUsername: "git",
		IncludeHTTPPath: true,
		SocketPath:      lambdaSocketPath,
		Timeout:         3600,
	}
	require.NoError(
		t,
		git.ConfigureCacheCredentialsHelper(logger, lambdaOpts),
	)
	require.NoError(
		t,
		git.StoreCacheCredentials(logger, "git", gitPAT, "github.com", "gruntwork-io/terraform-aws-lambda.git", lambdaSocketPath),
	)

	githubOpts := git.CacheCredentialOptions{
		Host:            "https://github.com",
		DefaultUsername: "git",
		IncludeHTTPPath: false,
		SocketPath:      githubSocketPath,
		Timeout:         3600,
	}
	require.NoError(
		t,
		git.ConfigureCacheCredentialsHelper(logger, githubOpts),
	)
	require.NoError(
		t,
		git.StoreCacheCredentials(logger, "git", "wrong-pat", "github.com", "", githubSocketPath),
	)

	tmpDir, err := ioutil.TempDir("", "git-test")
	require.NoError(t, err)
	lambdaDir := filepath.Join(tmpDir, "terraform-aws-lambda")
	ciDir := filepath.Join(tmpDir, "terraform-aws-ci")
	require.NoError(t, os.Mkdir(lambdaDir, 0o755))
	require.NoError(t, os.Mkdir(ciDir, 0o755))

	require.NoError(
		t,
		git.Clone(logger, lambdaGitURL, lambdaDir),
	)
	require.Error(
		t,
		git.Clone(logger, "https://github.com/gruntwork-io/terraform-aws-ci.git", ciDir),
	)

}

func TestForceHTTPS(t *testing.T) {
	defer cleanupGitConfig(t)

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

// cleanupGitConfig will reset the git credential cache and git config
func cleanupGitConfig(t *testing.T) {
	data, err := ioutil.ReadFile("/root/.gitconfig")
	require.NoError(t, err)
	ttlogger.Logf(t, string(data))

	require.NoError(t, os.Remove("/root/.gitconfig"))

	cmd := shell.Command{
		Command: "git",
		Args:    []string{"credential-cache", "exit"},
	}
	shell.RunCommand(t, cmd)
}
