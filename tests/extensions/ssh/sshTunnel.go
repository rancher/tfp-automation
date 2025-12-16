package ssh

import (
	"fmt"
	"io"
	"net"

	"golang.org/x/crypto/ssh"
)

type BastionSSHTunnel struct {
	listener net.Listener
	client   *ssh.Client
}

// StartBastionSSHTunnel establishes an SSH tunnel through a bastion host. This function specifically is for airgap setups,
// but can be adapted for other use cases as well.
func StartBastionSSHTunnel(bastionAddr, sshUser string, sshKey []byte, localPort, remoteHost, remotePort string) (*BastionSSHTunnel, error) {
	signer, err := ssh.ParsePrivateKey(sshKey)
	if err != nil {
		return nil, err
	}

	auths := []ssh.AuthMethod{ssh.PublicKeys(signer)}
	cfg := &ssh.ClientConfig{
		User:            sshUser,
		Auth:            auths,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	cfg.SetDefaults()
	client, err := ssh.Dial("tcp", bastionAddr+":22", cfg)
	if err != nil {
		return nil, err
	}

	listener, err := net.Listen("tcp", "127.0.0.1:"+localPort)
	if err != nil {
		client.Close()
		return nil, err
	}

	tunnel := &BastionSSHTunnel{
		client:   client,
		listener: listener,
	}

	go func() {
		for {
			localConn, err := listener.Accept()
			if err != nil {
				break
			}

			go func() {
				remoteConn, err := client.Dial("tcp", fmt.Sprintf("%s:%s", remoteHost, remotePort))
				if err != nil {
					localConn.Close()
					return
				}
				go io.Copy(remoteConn, localConn)
				go io.Copy(localConn, remoteConn)
			}()
		}
	}()

	return tunnel, nil
}

// StopBastionSSHTunnel stops the SSH tunnel by closing the listener and client connections.
func (t *BastionSSHTunnel) StopBastionSSHTunnel() {
	if t.listener != nil {
		_ = t.listener.Close()
	}

	if t.client != nil {
		_ = t.client.Close()
	}
}
