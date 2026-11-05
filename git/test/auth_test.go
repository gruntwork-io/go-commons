//go:build gittest

package gittest

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

	terratestGitURL = "https://github.com/gruntwork-io/terratest.git"
	cloudNukeGitURL = "https://github.com/gruntwork-io/cloud-nuke.git"
)

var (
	logger = logging.GetLogger("testlogger", "")
)

// NOTE: All these tests should be run in the provided docker environment to avoid polluting the local git configuration
// settings. The tests will assert that it is running in the docker environment, and will fail if it is not.
// All these tests are also run in serial to avoid race conditions on the git config file.
// The repositories cloned here are public, so GitHub accepts the clone with any token; tests that check which token
// was selected read it back from the credential helper instead.

func TestHTTPSAuth(t *testing.T) {
	defer cleanupGitConfig(t)

	currentDir, err := os.Getwd()
	require.NoError(t, err)
	require.Equal(t, "/workspace/go-commons/git/test", currentDir)

	environment.RequireEnvVar(t, gitPATEnvName)
	gitPAT := os.Getenv(gitPATEnvName)
	require.NoError(t, git.ConfigureHTTPSAuth(logger, "git", gitPAT, "github.com"))
	assert.Equal(t, gitPAT, credentialFill(t, terratestGitURL))

	tmpDir := t.TempDir()
	require.NoError(t, git.Clone(logger, terratestGitURL, tmpDir))
	assert.True(t, files.IsDir(filepath.Join(tmpDir, "modules")))
}

func TestHTTPSAuthWithPath(t *testing.T) {
	defer cleanupGitConfig(t)

	currentDir, err := os.Getwd()
	require.NoError(t, err)
	require.Equal(t, "/workspace/go-commons/git/test", currentDir)

	environment.RequireEnvVar(t, gitPATEnvName)
	gitPAT := os.Getenv(gitPATEnvName)

	terratestOpts := git.CacheCredentialOptions{
		Host:            terratestGitURL,
		DefaultUsername: "git",
		IncludeHTTPPath: true,
		SocketPath:      "",
		Timeout:         3600,
	}
	require.NoError(
		t,
		git.ConfigureCacheCredentialsHelper(logger, terratestOpts),
	)
	require.NoError(
		t,
		git.StoreCacheCredentials(logger, "git", gitPAT, "github.com", "gruntwork-io/terratest.git", ""),
	)
	assert.Equal(t, gitPAT, credentialFill(t, terratestGitURL))

	tmpDir := t.TempDir()
	require.NoError(t, git.Clone(logger, terratestGitURL, tmpDir))
	assert.True(t, files.IsDir(filepath.Join(tmpDir, "modules")))
}

func TestHTTPSAuthMixed(t *testing.T) {
	defer cleanupGitConfig(t)

	currentDir, err := os.Getwd()
	require.NoError(t, err)
	require.Equal(t, "/workspace/go-commons/git/test", currentDir)

	// Make sure the directory for git credential sockets exist
	socketPath := "/tmp/git-credential-sockets"
	require.NoError(t, os.MkdirAll(socketPath, 0o700))
	terratestSocketPath := filepath.Join(socketPath, "terratest")
	githubSocketPath := filepath.Join(socketPath, "github")

	environment.RequireEnvVar(t, gitPATEnvName)
	gitPAT := os.Getenv(gitPATEnvName)

	terratestOpts := git.CacheCredentialOptions{
		Host:            terratestGitURL,
		DefaultUsername: "git",
		IncludeHTTPPath: true,
		SocketPath:      terratestSocketPath,
		Timeout:         3600,
	}
	require.NoError(
		t,
		git.ConfigureCacheCredentialsHelper(logger, terratestOpts),
	)
	require.NoError(
		t,
		git.StoreCacheCredentials(logger, "git", gitPAT, "github.com", "gruntwork-io/terratest.git", terratestSocketPath),
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

	assert.Equal(t, gitPAT, credentialFill(t, terratestGitURL))
	assert.Equal(t, "wrong-pat", credentialFill(t, cloudNukeGitURL))

	terratestDir := filepath.Join(t.TempDir(), "terratest")
	require.NoError(t, os.Mkdir(terratestDir, 0o755))
	require.NoError(t, git.Clone(logger, terratestGitURL, terratestDir))
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

	tmpDir := t.TempDir()
	require.NoError(t, git.Clone(logger, "git@github.com:gruntwork-io/terratest.git", tmpDir))
	assert.True(t, files.IsDir(filepath.Join(tmpDir, "modules")))
}

// credentialFill returns the password git's configured credential helpers supply for the given URL.
func credentialFill(t *testing.T, url string) string {
	cmd := exec.Command("git", "credential", "fill")
	cmd.Stdin = strings.NewReader("url=" + url + "\n\n")
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.Output()
	require.NoError(t, err)

	for _, line := range strings.Split(string(out), "\n") {
		if password, ok := strings.CutPrefix(line, "password="); ok {
			return password
		}
	}
	t.Fatalf("git credential fill returned no password for %s", url)
	return ""
}

// cleanupGitConfig will reset the git credential cache and git config
func cleanupGitConfig(t *testing.T) {
	data, err := os.ReadFile("/root/.gitconfig")
	require.NoError(t, err)
	ttlogger.Log(t, string(data))

	require.NoError(t, os.Remove("/root/.gitconfig"))

	cmd := shell.Command{
		Command: "git",
		Args:    []string{"credential-cache", "exit"},
	}
	shell.RunCommand(t, cmd)
}
