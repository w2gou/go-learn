package sshclient

import (
	"fmt"
	"go-learn/cli/sshConnect/version2/config"
	"golang.org/x/crypto/ssh"
	"os"
)

func InteractiveShell(conn config.Connection) error {
	cfg := &ssh.ClientConfig{
		User: conn.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(conn.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	address := fmt.Sprintf("%s:%d", conn.Host, conn.Port)
	client, err := ssh.Dial("tcp", address, cfg)
	if err != nil {
		return err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	// 将本地输入输出绑定到远程
	session.Stdin = os.Stdin
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		return err
	}

	if err := session.Shell(); err != nil {
		return err
	}

	return session.Wait()
}
