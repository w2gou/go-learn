package sftpclient

import (
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
	"os"
	"path/filepath"
)

type Client struct {
	raw *sftp.Client
}

func New(sshClient *ssh.Client) (*Client, error) {
	s, err := sftp.NewClient(sshClient)
	if err != nil {
		return nil, fmt.Errorf("sftp new: %w", err)
	}
	return &Client{raw: s}, nil
}

func (c *Client) Close() error {
	if c == nil || c.raw == nil {
		return nil
	}
	return c.raw.Close()
}

// Upload: 本地 -> 远端
func (c *Client) Upload(local, remote string) error {
	lf, err := os.Open(local)
	if err != nil {
		return fmt.Errorf("open local: %w", err)
	}
	defer lf.Close()

	_ = c.raw.MkdirAll(filepath.Dir(remote)) // 忽略已存在错误

	rf, err := c.raw.Create(remote)
	if err != nil {
		return fmt.Errorf("create remote: %w", err)
	}
	defer rf.Close()

	if _, err := io.Copy(rf, lf); err != nil {
		return fmt.Errorf("copy to remote: %w", err)
	}
	fmt.Printf("上传成功: %s -> %s\n", local, remote)
	return nil
}

// Download: 远端 -> 本地
func (c *Client) Download(remote, local string) error {
	rf, err := c.raw.Open(remote)
	if err != nil {
		return fmt.Errorf("open remote: %w", err)
	}
	defer rf.Close()

	if err := os.MkdirAll(filepath.Dir(local), 0o755); err != nil && !os.IsExist(err) {
		return fmt.Errorf("mkdir local: %w", err)
	}

	lf, err := os.Create(local)
	if err != nil {
		return fmt.Errorf("create local: %w", err)
	}
	defer lf.Close()

	if _, err := io.Copy(lf, rf); err != nil {
		return fmt.Errorf("copy to local: %w", err)
	}
	fmt.Printf("下载成功: %s -> %s\n", remote, local)
	return nil
}
