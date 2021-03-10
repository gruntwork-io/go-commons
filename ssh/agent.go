package ssh

import (
	"crypto/x509"
	"encoding/pem"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/agent"
)

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
func NewSSHAgent(socketDir string, socketFile string) (*SSHAgent, error) {
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
func SSHAgentWithPrivateKey(privateKey string) (*SSHAgent, error) {
	sshAgent, err := SSHAgentWithPrivateKeys([]string{privateKey})
	return sshAgent, err
}

// Instantiates and returns an in-memory ssh agent with the given private key(s) already added
// You should stop the agent to cleanup files afterwards by calling `defer sshAgent.Stop()`
func SSHAgentWithPrivateKeys(privateKeys []string) (*SSHAgent, error) {
	// Instantiate a temporary SSH agent
	socketDir, err := ioutil.TempDir("", "ssh-agent-")
	if err != nil {
		return nil, err
	}
	socketFile := filepath.Join(socketDir, "ssh_auth.sock")
	sshAgent, err := NewSSHAgent(socketDir, socketFile)
	if err != nil {
		return nil, err
	}

	// add given ssh keys to the newly created agent
	for _, privateKey := range privateKeys {
		// Create SSH key for the agent using the given SSH key pair(s)
		block, _ := pem.Decode([]byte(privateKey))
		decodedPrivateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		key := agent.AddedKey{PrivateKey: decodedPrivateKey}
		sshAgent.agent.Add(key)
	}

	return sshAgent, err
}
