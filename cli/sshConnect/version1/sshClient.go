package version1

import (
	"golang.org/x/crypto/ssh"
	"os"
	"strconv"
	"time"
)

func connect(conn connection) (*ssh.Client, error) {
	keyData, err := os.ReadFile(conn.Key)
	if err != nil {
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(keyData)
	if err != nil {
		return nil, err
	}

	config := &ssh.ClientConfig{
		User:            conn.User,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	addr := conn.Host + ":" + strconv.Itoa(conn.Port)
	return ssh.Dial("tcp", addr, config)
}

func execCommand(client *ssh.Client, cmd string) error {
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	return session.Run(cmd)
}
