// Package ssh allows to manage SSH connections and send commands through them.
package ssh

import (
	"fmt"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

type outputReturnType int

const (
	stdout outputReturnType = iota
	combined
)

// RunCommandAndGetStdout will run the given command on the host over SSH and return the stdout of the underlying
// command.
func RunCommandAndGetStdout(host Host, cmd string) (string, error) {
	return runSSHCommand(host, cmd, stdout)
}

// RunCommandAndGetOutput will run the given command on the host over SSH and return the combined (stdout + stderr)
// output.
func RunCommandAndGetOutput(host Host, cmd string) (string, error) {
	return runSSHCommand(host, cmd, combined)
}

// runSSHCommand will setup a new SSH connection to the given host, run the command, and return the stdout or combined
// output depending on which output is requested.
func runSSHCommand(host Host, cmd string, retType outputReturnType) (output string, returnErr error) {
	hostOptions, err := setUpSSHOptions(host)
	if err != nil {
		return "", err
	}
	hostOptions.Command = cmd

	closeStack := &sshCloseStack{stack: []Closeable{}}
	defer func() {
		if err := closeStack.CloseAll(); err != nil && returnErr == nil {
			returnErr = err
		}
	}()

	client, err := setUpSSHClient(hostOptions, closeStack)
	if err != nil {
		return "", err
	}

	session, err := client.NewSession()
	closeStack.Push(session)
	if err != nil {
		return "", err
	}

	switch retType {
	case stdout:
		bytes, err := session.Output(hostOptions.Command)
		return string(bytes), err
	case combined:
		bytes, err := session.CombinedOutput(hostOptions.Command)
		return string(bytes), err
	}
	return "", fmt.Errorf("runSSHCommand - This condition is impossible: UNKNOWN RETURN TYPE")
}

// setUpSSHOptions will configure SSHConnectionOptions for the given host, recursively configuring jump host connections
// as needed.
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

// setUpSSHClient will configure SSH client connections for the given host, recursively configuring jump host
// connections as needed.
func setUpSSHClient(currentOpts *sshConnectionOptions, closeStack *sshCloseStack) (*ssh.Client, error) {
	// Base case: no more jump hosts defined
	// Since there is no jump host, we directly configure the connection to the server.
	if currentOpts.JumpHostOptions == nil {
		client, err := createSSHClient(currentOpts)
		closeStack.Push(client)
		return client, err
	}

	// Here, we know there is a jump host connection that is needed to connect to this host, so we recursively set up
	// the jump host connection and use that to connect to the host.
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

// createSSHClient configures a direct SSH connection to the given host connection option.
func createSSHClient(options *sshConnectionOptions) (*ssh.Client, error) {
	sshClientConfig := createSSHClientConfig(options)
	return ssh.Dial("tcp", options.ConnectionString(), sshClientConfig)
}

// createSSHClientConfig returns the SSH client configuration to use when setting up the connection to the host.
func createSSHClientConfig(hostOptions *sshConnectionOptions) *ssh.ClientConfig {
	clientConfig := &ssh.ClientConfig{
		User:            hostOptions.Username,
		Auth:            hostOptions.AuthMethods,
		HostKeyCallback: hostOptions.HostKeyCallback,
		// By default, Go does not impose a timeout, so a SSH connection attempt can hang for a LONG time.
		Timeout: 10 * time.Second,
	}
	clientConfig.SetDefaults()
	return clientConfig
}

// NoOpHostKeyCallback is an ssh.HostKeyCallback that does nothing. Only use this when you're sure you don't want to
// check the host key at all (e.g., only for testing and non-production use cases).
func NoOpHostKeyCallback(hostname string, remote net.Addr, key ssh.PublicKey) error {
	return nil
}
