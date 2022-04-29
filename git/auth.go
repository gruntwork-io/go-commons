package git

import (
	"fmt"

	"github.com/gruntwork-io/go-commons/shell"
	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
)

// CredentialOptions are the possible configurations options for configuring the git credential-cache helper.
type CacheCredentialOptions struct {
	// Host is the VCS host where the cache credential helper should be triggered.
	// Set to "" if you want to apply the credential cache helper to all hosts.
	Host string

	// DefaultUsername is the default username to use when authenticating to the VCS host.
	DefaultUsername string

	// IncludeHTTPPath indicates whether to path through the git http path to the credential helper, enabling matching
	// with the path (e.g., the org/repo.git component of https://github.com/org/repo.git.
	IncludeHTTPPath bool

	// SocketPath configures the path to the Unix socket file to use to interact with the cache credential daemon. When
	// blank, uses the default path baked into the command:
	// https://git-scm.com/docs/git-credential-cache#_options
	// This is useful when you are configuring the cache for the same host across multiple paths, as the cache
	// credential helper is known to break when you have an entry for a specific path and the generic all hosts.
	SocketPath string

	// Timeout is the timeout in seconds for credentials in the cache.
	Timeout int
}

// ConfigureForceHTTPS configures git to force usage of https endpoints instead of SSH based endpoints for the three
// primary VCS platforms (GitHub, GitLab, BitBucket).
func ConfigureForceHTTPS(logger *logrus.Logger) error {
	opts := shell.NewShellOptions()
	if logger != nil {
		opts.Logger = logger
	}

	var allErr error

	for _, host := range []string{"github.com", "gitlab.com", "bitbucket.org"} {
		if err := shell.RunShellCommand(
			opts,
			"git", "config", "--global",
			fmt.Sprintf("url.https://%s.insteadOf", host),
			fmt.Sprintf("ssh://git@%s", host),
		); err != nil {
			allErr = multierror.Append(allErr, err)
		}

		if err := shell.RunShellCommand(
			opts,
			"git", "config", "--global",
			fmt.Sprintf("url.https://%s/.insteadOf", host),
			fmt.Sprintf("git@%s:", host),
		); err != nil {
			allErr = multierror.Append(allErr, err)
		}
	}
	return allErr
}

// ConfigureHTTPSAuth configures git with username and password to authenticate with the given VCS host when interacting
// with git over HTTPS. This uses the cache credentials store to configure the credentials. Refer to the git
// documentation on credentials storage for more information:
// https://git-scm.com/book/en/v2/Git-Tools-Credential-Storage
// NOTE: this configures the cache credential helper globally, with a default timeout of 1 hour. If you want more
// control over the configuration, use the ConfigureCacheCredentialsHelper and StoreCacheCredentials functions directly.
func ConfigureHTTPSAuth(
	logger *logrus.Logger,
	gitUsername string,
	gitOauthToken string,
	vcsHost string,
) error {
	// Legacy options that were in use when function was first introduced.
	cacheOpts := CacheCredentialOptions{
		Host:            "",
		DefaultUsername: "",
		IncludeHTTPPath: false,
		SocketPath:      "",
		Timeout:         3600,
	}
	if err := ConfigureCacheCredentialsHelper(logger, cacheOpts); err != nil {
		return err
	}
	return StoreCacheCredentials(logger, gitUsername, gitOauthToken, vcsHost, "", "")
}

// ConfigureCacheCredentialsHelper configures git globally to use the cache credentials helper for authentication based
// on the provided options configuration.
func ConfigureCacheCredentialsHelper(logger *logrus.Logger, options CacheCredentialOptions) error {
	shellOpts := shell.NewShellOptions()
	if logger != nil {
		shellOpts.Logger = logger
	}

	credentialConfigPrefix := "credential"
	if options.Host != "" {
		credentialConfigPrefix += fmt.Sprintf(".%s", options.Host)
	}

	helperOpts := fmt.Sprintf("--timeout %d", options.Timeout)
	if options.SocketPath != "" {
		helperOpts += fmt.Sprintf(" --socket %s", options.SocketPath)
	}
	if err := shell.RunShellCommand(
		shellOpts,
		"git", "config", "--global",
		credentialConfigPrefix+".helper",
		"cache "+helperOpts,
	); err != nil {
		return err
	}

	if options.DefaultUsername != "" {
		if err := shell.RunShellCommand(
			shellOpts,
			"git", "config", "--global",
			credentialConfigPrefix+".username",
			options.DefaultUsername,
		); err != nil {
			return err
		}
	}

	if options.IncludeHTTPPath {
		if err := shell.RunShellCommand(
			shellOpts,
			"git", "config", "--global",
			credentialConfigPrefix+".useHttpPath",
			"true",
		); err != nil {
			return err
		}
	}

	return nil
}

// StoreCacheCredentials stores the given git credentials for the vcs host and path pair to the git credential-cache
// helper.
func StoreCacheCredentials(
	logger *logrus.Logger,
	gitUsername string,
	gitOauthToken string,
	vcsHost string,
	vcsPath string,
	socketPath string,
) error {
	opts := shell.NewShellOptions()
	if logger != nil {
		opts.Logger = logger
	}

	if gitUsername == "" {
		gitUsername = "git"
	}
	credentialsStoreInput := fmt.Sprintf(`protocol=https
host=%s
username=%s
password=%s`, vcsHost, gitUsername, gitOauthToken)
	if vcsPath != "" {
		credentialsStoreInput += fmt.Sprintf("\npath=%s", vcsPath)
	}

	cmdArgs := []string{"credential-cache"}
	if socketPath != "" {
		cmdArgs = append(cmdArgs, "--socket", socketPath)
	}
	cmdArgs = append(cmdArgs, "store")
	return shell.RunShellCommandWithInput(opts, credentialsStoreInput, "git", cmdArgs...)
}
