package ssh

import (
	"crypto/x509"
	"encoding/pem"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/agent"
)

// SSHAgent represents an instance of the ssh-agent process.
type SSHAgent struct {
	stop       chan bool
	stopped    chan bool
	socketDir  string
	socketFile string
	agent      agent.Agent
	ln         net.Listener

	logger *logrus.Entry
}

// Create SSH agent, start it in background and returns control back to the main thread
// You should stop the agent to cleanup files afterwards by calling `defer s.Stop()`
func NewSSHAgent(logger *logrus.Entry, socketDir string, socketFile string) (*SSHAgent, error) {
	var err error
	s := &SSHAgent{
		stop:       make(chan bool),
		stopped:    make(chan bool),
		socketDir:  socketDir,
		socketFile: socketFile,
		agent:      agent.NewKeyring(),
	}
	s.ln, err = net.Listen("unix", s.socketFile)
	if err != nil {
		return nil, err
	}
	go s.run()
	return s, nil
}

// expose socketFile variable
func (s *SSHAgent) SocketFile() string {
	return s.socketFile
}

// SSH Agent listener and handler
func (s *SSHAgent) run() {
	defer close(s.stopped)
	for {
		select {
		case <-s.stop:
			return
		default:
			c, err := s.ln.Accept()
			if err != nil {
				select {
				// When s.Stop() closes the listener, s.ln.Accept() returns an error that can be ignored
				// since the agent is in stopping process
				case <-s.stop:
					return
					// When s.ln.Accept() returns a legit error, we print it and continue accepting further requests
				default:
					if s.logger != nil {
						s.logger.Errorf("could not accept connection to agent %v", err)
					}
					continue
				}
			} else {
				defer c.Close()
				go func(c io.ReadWriter) {
					err := agent.ServeAgent(s.agent, c)
					if err != nil && s.logger != nil {
						s.logger.Errorf("could not serve ssh agent %v", err)
					}
				}(c)
			}
		}
	}
}

// Stop and clean up SSH agent
func (s *SSHAgent) Stop() {
	close(s.stop)
	s.ln.Close()
	<-s.stopped
	os.RemoveAll(s.socketDir)
}

// Instantiates and returns an in-memory ssh agent with the given private key already added
// You should stop the agent to cleanup files afterwards by calling `defer sshAgent.Stop()`
func SSHAgentWithPrivateKey(logger *logrus.Entry, privateKey string) (*SSHAgent, error) {
	sshAgent, err := SSHAgentWithPrivateKeys(logger, []string{privateKey})
	return sshAgent, err
}

// Instantiates and returns an in-memory ssh agent with the given private key(s) already added
// You should stop the agent to cleanup files afterwards by calling `defer sshAgent.Stop()`
func SSHAgentWithPrivateKeys(logger *logrus.Entry, privateKeys []string) (*SSHAgent, error) {
	// Instantiate a temporary SSH agent
	socketDir, err := ioutil.TempDir("", "ssh-agent-")
	if err != nil {
		return nil, err
	}
	socketFile := filepath.Join(socketDir, "ssh_auth.sock")
	sshAgent, err := NewSSHAgent(logger, socketDir, socketFile)
	if err != nil {
		return nil, err
	}

	// add given ssh keys to the newly created agent
	var allErrs *multierror.Error
	for _, privateKey := range privateKeys {
		// Create SSH key for the agent using the given SSH key pair(s)
		block, _ := pem.Decode([]byte(privateKey))
		decodedPrivateKey, err := decodePrivateKey(block.Bytes)
		if err != nil {
			logger.Error("Error decoding private key for adding to ssh-agent")
			allErrs = multierror.Append(allErrs, err)
		} else {
			key := agent.AddedKey{PrivateKey: decodedPrivateKey}
			if err := sshAgent.agent.Add(key); err != nil {
				logger.Error("Error adding private key ssh-agent")
				allErrs = multierror.Append(allErrs, err)
			}
		}
	}
	return sshAgent, allErrs.ErrorOrNil()
}

// decodePrivateKey first attempts to decode the key as PKCS8, and then fallsback to PKCS1 if that fails.
// This function returns a *rsa.PrivateKey, a *ecdsa.PrivateKey, or a ed25519.PrivateKey.
func decodePrivateKey(keyBytes []byte) (interface{}, error) {
	var allErrs *multierror.Error
	decodedPrivateKey, err := x509.ParsePKCS8PrivateKey(keyBytes)
	if err != nil {
		allErrs = multierror.Append(allErrs, err)
		decodedPrivateKey, err = x509.ParsePKCS1PrivateKey(keyBytes)
		if err != nil {
			allErrs = multierror.Append(allErrs, err)
			return nil, allErrs.ErrorOrNil()
		}
	}
	return decodedPrivateKey, nil
}
