package sqsbuddy

import (
	"errors"
	"golang.org/x/crypto/ssh"
	"os"
)

type SSHBuddy struct {
	User          string
	Address       string
	Command       string
	KeyPath       string
	KeyPassphrase string
	InsecureHosts bool
}

func (b *SSHBuddy) Run() ([]byte, error) {
	cfg := &ssh.ClientConfig{
		User: b.User,
		Auth: []ssh.AuthMethod{},
	}
	if b.InsecureHosts {
		cfg.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	} else {
		return nil, errors.New("unsupported")
	}
	if b.KeyPath == "" {
		return nil, errors.New("unsupported: need key")
	}

	data, err := os.ReadFile(b.KeyPath)
	if err != nil {
		return nil, err
	}
	signer, err := b.parse(data)
	if err != nil {
		return nil, err
	}
	cfg.Auth = append(cfg.Auth, ssh.PublicKeys(signer))

	conn, err := ssh.Dial("tcp", b.Address, cfg)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	sess, err := conn.NewSession()
	if err != nil {
		return nil, err
	}
	defer sess.Close()
	return sess.CombinedOutput(b.Command)
}

func (b *SSHBuddy) parse(in []byte) (ssh.Signer, error) {
	if b.KeyPassphrase == "" {
		return ssh.ParsePrivateKey(in)
	}
	return ssh.ParsePrivateKeyWithPassphrase(in, []byte(b.KeyPassphrase))
}
