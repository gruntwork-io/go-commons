package github

import (
	"context"
	"os"
	"strconv"
	"testing"

	"github.com/google/go-github/v44/github"
	"github.com/gruntwork-io/terratest/modules/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

const (
	ghAppIDEnv               = "GITHUB_APP_ID"
	ghAppInstallationIDEnv   = "GITHUB_APP_INSTALLATION_ID"
	ghAppPrivateKeyBase64Env = "GITHUB_APP_PRIVATE_KEY_PEM_BASE64"
)

func TestGithubAppConfig(t *testing.T) {
	t.Parallel()

	environment.RequireEnvVar(t, ghAppIDEnv)
	environment.RequireEnvVar(t, ghAppInstallationIDEnv)
	environment.RequireEnvVar(t, ghAppPrivateKeyBase64Env)

	ghAppConfig := getGitHubAppConfig(t)
	token, err := ghAppConfig.GetInstallationToken()
	require.NoError(t, err)

	gh := newGithubClient(token)
	user, _, err := gh.Users.Get(context.Background(), "")
	require.NoError(t, err)
	require.NotNil(t, user.Login)
	assert.Equal(t, "foo", *user.Login)
}

func getGitHubAppConfig(t *testing.T) *GithubAppConfig {
	appIDStr := os.Getenv(ghAppIDEnv)
	appID, err := strconv.ParseInt(appIDStr, 10, 64)
	require.NoError(t, err)

	appInstallationIDStr := os.Getenv(ghAppInstallationIDEnv)
	appInstallationID, err := strconv.ParseInt(appInstallationIDStr, 10, 64)
	require.NoError(t, err)

	return &GithubAppConfig{
		AppID:               appID,
		AppInstallationID:   appInstallationID,
		PrivateKeyPEMBase64: os.Getenv(ghAppPrivateKeyBase64Env),
	}
}

func newGithubClient(token string) *github.Client {
	tokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)

	tokenClient := oauth2.NewClient(context.Background(), tokenSource)
	client := github.NewClient(tokenClient)
	return client
}
