package github

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/gruntwork-io/go-commons/entrypoint"
	"github.com/gruntwork-io/go-commons/errors"
)

// GithubAppConfig represents configuration settings for a Github App that can be used to authenticate to the Github
// API.
type GithubAppConfig struct {
	// ID of the Github App to authenticate as.
	AppID int64 `json:"app_id"`

	// ID of the Github App installation to authenticate as. This ID determines the Github Org that the App can access.
	AppInstallationID int64 `json:"app_installation_id"`

	// The PEM encoded private key (base64 encoded) that can be used to obtain an installation token to access the
	// Github API.
	PrivateKeyPEMBase64 string `json:"private_key_pem"`
}

// GetInstallationToken uses the configured GitHub App credentials to obtain an installation token that can be used to
// access Github using both the API and Git CLI. This token works the same as an Oauth token, or Personal Access Token.
func (config *GithubAppConfig) GetInstallationToken() (string, error) {
	if config == nil {
		return "", errors.WithStackTrace(fmt.Errorf("GithubAppConfig is nil"))
	}

	// Decode private key PEM, which should be base64 encoded to allow easy handling of newlines.
	privateKeyPEM, err := base64.StdEncoding.DecodeString(config.PrivateKeyPEMBase64)
	if err != nil {
		return "", err
	}

	// Retrieve a valid installation token that can be used to authenticate requests to GitHub, or clone repos over
	// HTTPS.
	itr, err := ghinstallation.New(
		http.DefaultTransport,
		config.AppID,
		config.AppInstallationID,
		privateKeyPEM,
	)
	if err != nil {
		return "", errors.WithStackTrace(err)
	}
	token, err := itr.Token(context.Background())
	return token, errors.WithStackTrace(err)
}

// LoadGithubAppConfigFromEnv will load a Github App Configuration from the given environment variable, assuming it is
// encoded in JSON format.
func LoadGithubAppConfigFromEnv(envVarName string) (*GithubAppConfig, error) {
	githubAppConfigJSON, err := entrypoint.EnvironmentVarRequiredE(envVarName)
	if err != nil {
		return nil, err
	}

	var githubAppConfig *GithubAppConfig
	jsonLoadErr := json.Unmarshal([]byte(githubAppConfigJSON), githubAppConfig)
	return githubAppConfig, errors.WithStackTrace(jsonLoadErr)
}
