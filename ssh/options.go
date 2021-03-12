package ssh

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"reflect"
	"strconv"

	multierror "github.com/hashicorp/go-multierror"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"

	"github.com/gruntwork-io/go-commons/collections"
)

// Host is a remote host.
type Host struct {
	Hostname    string // host name or ip address
	SSHUserName string // user name
	CustomPort  int    // port number to use to connect to the host (port 22 will be used if unset)

	JumpHost *Host // Another host to use as a jump host to reach this host.

	HostKeyCallback ssh.HostKeyCallback // Callback function for handling host key checks.

	// set one or more authentication methods,
	// the first valid method will be used
	PrivateKey       string    // ssh private key to use as authentication method (disabled by default)
	SSHAgent         bool      // enable authentication using your existing local SSH agent (disabled by default)
	OverrideSSHAgent *SSHAgent // enable an in process `SSHAgent` for connections to this host (disabled by default)
	Password         string    // plain text password (blank by default)
}

// getSSHConnectionOptions converts the host configuration into a set of options that can be used for managing the SSH
// connection.
func (host *Host) getSSHConnectionOptions() (*sshConnectionOptions, error) {
	if host == nil {
		return nil, nil
	}

	authMethods, err := host.createAuthMethods()
	if err != nil {
		return nil, err
	}

	hostOptions := sshConnectionOptions{
		Username:        host.SSHUserName,
		Address:         host.Hostname,
		Port:            host.getPort(),
		HostKeyCallback: host.HostKeyCallback,
		AuthMethods:     authMethods,
	}
	return &hostOptions, nil
}

// getPort gets the port that should be used to communicate with the host
func (h Host) getPort() int {

	//If a CustomPort is not set use standard ssh port
	if h.CustomPort == 0 {
		return 22
	} else {
		return h.CustomPort
	}
}

// createAuthMethods returns an array of authentication methods
func (host Host) createAuthMethods() ([]ssh.AuthMethod, error) {
	var methods []ssh.AuthMethod

	// override local ssh agent with given sshAgent instance
	if host.OverrideSSHAgent != nil {
		conn, err := net.Dial("unix", host.OverrideSSHAgent.socketFile)
		if err != nil {
			fmt.Print("Failed to dial in memory ssh agent")
			return methods, err
		}
		agentClient := agent.NewClient(conn)
		methods = append(methods, []ssh.AuthMethod{ssh.PublicKeysCallback(agentClient.Signers)}...)
	}

	// use existing ssh agent socket
	// if agent authentication is enabled and no agent is set up, returns an error
	if host.SSHAgent {
		socket := os.Getenv("SSH_AUTH_SOCK")
		conn, err := net.Dial("unix", socket)
		if err != nil {
			return methods, err
		}
		agentClient := agent.NewClient(conn)
		methods = append(methods, []ssh.AuthMethod{ssh.PublicKeysCallback(agentClient.Signers)}...)
	}

	// use provided ssh key pair
	if host.PrivateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(host.PrivateKey))
		if err != nil {
			return methods, err
		}
		methods = append(methods, []ssh.AuthMethod{ssh.PublicKeys(signer)}...)
	}

	// Use given password
	if len(host.Password) > 0 {
		methods = append(methods, []ssh.AuthMethod{ssh.Password(host.Password)}...)
	}

	// no valid authentication method was provided
	if len(methods) < 1 {
		return methods, errors.New("no authentication method defined")
	}

	return methods, nil
}

// sshConnectionOptions are the options for an SSH connection.
type sshConnectionOptions struct {
	Username        string
	Address         string
	Port            int
	AuthMethods     []ssh.AuthMethod
	HostKeyCallback ssh.HostKeyCallback
	Command         string
	JumpHostOptions *sshConnectionOptions
}

// ConnectionString returns the connection string for an SSH connection.
func (options *sshConnectionOptions) ConnectionString() string {
	return net.JoinHostPort(options.Address, strconv.Itoa(options.Port))
}

// sshCloseStack is a LIFO (stack) data structure for tracking all the resources that need to be closed at the end of an
// SSH connection. This is useful for having a single defer call in a top-level method to clean up resources that are
// recursively created across multiple jump hosts.
type sshCloseStack struct {
	stack []Closeable
}

// Push will push an item on the close stack by prepending the item to the top of the array.
func (this *sshCloseStack) Push(item Closeable) {
	this.stack = append([]Closeable{item}, this.stack...)
}

// CloseAll iterates over all the closeable items and closes the connection one by one. This will attempt to close
// everything in the stack regardless of errors, and return a single multierror at the end that aggregates all
// encountered errors.
func (this *sshCloseStack) CloseAll() error {
	allErrs := &multierror.Error{}
	for _, closeable := range this.stack {
		// Closing a connection may result in an EOF error if it's already closed (e.g. due to hitting CTRL + D), so
		// don't report those errors, as there is nothing actually wrong in that case.
		allErrs = multierror.Append(allErrs, Close(closeable, io.EOF.Error()))
	}
	return allErrs.ErrorOrNil()
}

// Closeable can be closed.
type Closeable interface {
	Close() error
}

// Close closes a Closeable.
func Close(closeable Closeable, ignoreErrors ...string) error {
	if interfaceIsNil(closeable) {
		return nil
	}

	if err := closeable.Close(); err != nil && !collections.ListContainsElement(ignoreErrors, err.Error()) {
		return err
	}
	return nil
}

// Checking an interface directly against nil does not work, and if you don't know the exact types the interface may be
// ahead of time, the only way to know if you're dealing with nil is to use reflection.
// http://stackoverflow.com/questions/13476349/check-for-nil-and-nil-interface-in-go
func interfaceIsNil(i interface{}) bool {
	return i == nil || reflect.ValueOf(i).IsNil()
}
