// Package ssh allows to manage SSH connections and send commands through them.
package ssh

import (
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

// RunCommandAndGetOutput will run the given command on the host over SSH and return the stdout of the underlying
// command.
func RunCommandAndGetStdout(host Host, cmd string) (stdout string, returnErr error) {
	hostOptions, err := setUpSSHOptions(host)
	if err != nil {
		return "", err
	}
	hostOptions.Command = cmd
	return runSSHCommand(hostOptions)
}

func runSSHCommand(sshOptions *sshConnectionOptions) (stdout string, returnErr error) {
	closeStack := &sshCloseStack{stack: []Closeable{}}
	defer func() {
		if err := closeStack.CloseAll(); err != nil && returnErr == nil {
			returnErr = err
		}
	}()

	client, err := setUpSSHClient(sshOptions, closeStack)
	if err != nil {
		return "", err
	}

	session, err := client.NewSession()
	closeStack.Push(session)
	if err != nil {
		return "", err
	}

	bytes, err := session.Output(sshOptions.Command)
	return string(bytes), err
}

func setUpSSHOptions(host Host) (*sshConnectionOptions, error) {
	hostOptions, err := host.getSSHConnectionOptions()
	if err != nil {
		return nil, err
	}

	if host.JumpHost != nil {
		jumpHostOptions, err := setUpSSHOptions(*host.JumpHost)
		if err != nil {
			return nil, err
		}
		hostOptions.JumpHostOptions = jumpHostOptions
	}
	return hostOptions, nil
}

func setUpSSHClient(currentOpts *sshConnectionOptions, closeStack *sshCloseStack) (*ssh.Client, error) {
	if currentOpts.JumpHostOptions == nil {
		client, err := createSSHClient(currentOpts)
		closeStack.Push(client)
		return client, err
	}

	jumpHostClient, err := setUpSSHClient(currentOpts.JumpHostOptions, closeStack)
	if err != nil {
		return nil, err
	}

	hostVirtualConn, err := jumpHostClient.Dial(
		"tcp",
		currentOpts.ConnectionString(),
	)
	closeStack.Push(hostVirtualConn)
	if err != nil {
		return nil, err
	}

	hostConn, hostIncomingChannels, hostIncomingRequests, err := ssh.NewClientConn(
		hostVirtualConn,
		currentOpts.ConnectionString(),
		createSSHClientConfig(currentOpts),
	)
	closeStack.Push(hostConn)
	if err != nil {
		return nil, err
	}

	client := ssh.NewClient(hostConn, hostIncomingChannels, hostIncomingRequests)
	closeStack.Push(client)
	return client, nil
}

func createSSHClient(options *sshConnectionOptions) (*ssh.Client, error) {
	sshClientConfig := createSSHClientConfig(options)
	return ssh.Dial("tcp", options.ConnectionString(), sshClientConfig)
}

func createSSHClientConfig(hostOptions *sshConnectionOptions) *ssh.ClientConfig {
	clientConfig := &ssh.ClientConfig{
		User: hostOptions.Username,
		Auth: hostOptions.AuthMethods,
		// Do not do a host key check, as Terratest is only used for testing, not prod
		HostKeyCallback: NoOpHostKeyCallback,
		// By default, Go does not impose a timeout, so a SSH connection attempt can hang for a LONG time.
		Timeout: 10 * time.Second,
	}
	clientConfig.SetDefaults()
	return clientConfig
}

// NoOpHostKeyCallback is an ssh.HostKeyCallback that does nothing. Only use this when you're sure you don't want to check the host key at all
// (e.g., only for testing and non-production use cases).
func NoOpHostKeyCallback(hostname string, remote net.Addr, key ssh.PublicKey) error {
	return nil
}
