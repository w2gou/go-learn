package sftpclient

import (
    "fmt"
    "github.com/pkg/sftp"
    "golang.org/x/crypto/ssh"
    "io"
    "os"
    "path"
    "path/filepath"
    "strings"
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

// ReadDir 列出远端目录（用于补全）。
func (c *Client) ReadDir(p string) ([]os.FileInfo, error) {
    return c.raw.ReadDir(p)
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
    info, err := c.raw.Stat(remote)
    if err != nil {
        return fmt.Errorf("stat remote: %w", err)
    }

	// 如果本地是目录或以分隔符结尾，将目标视为目录
	localIsDir := false
    if st, err := os.Stat(local); err == nil && st.IsDir() {
        localIsDir = true
    }
	if strings.HasSuffix(local, string(os.PathSeparator)) || strings.HasSuffix(local, "/") || strings.HasSuffix(local, "\\") {
		localIsDir = true
	}

    if info.IsDir() {
        // 递归下载目录，包含远端顶层目录名
        if !localIsDir {
            localIsDir = true
        }
        // 目标根 = local/<basename(remote)>
        base := filepath.Base(filepath.FromSlash(remote))
        destRoot := local
        if localIsDir {
            destRoot = filepath.Join(local, base)
        }
        if err := os.MkdirAll(destRoot, 0o755); err != nil && !os.IsExist(err) {
            return fmt.Errorf("mkdir local dir: %w", err)
        }
        return c.downloadDir(remote, destRoot)
    }

    // 进一步启发：若 local 不存在且看起来像目录名（无扩展名），按目录处理
    if !localIsDir {
        if _, err := os.Stat(local); os.IsNotExist(err) {
            base := filepath.Base(local)
            if !strings.Contains(base, ".") { // 简单启发：无扩展名视为目录
                localIsDir = true
            }
        }
    }
    // 下载文件：若 local 是目录，则拼上远端文件名
    if localIsDir {
        local = filepath.Join(local, filepath.Base(remote))
    }
	if err := os.MkdirAll(filepath.Dir(local), 0o755); err != nil && !os.IsExist(err) {
		return fmt.Errorf("mkdir local: %w", err)
	}
	rf, err := c.raw.Open(remote)
	if err != nil {
		return fmt.Errorf("open remote: %w", err)
	}
	defer rf.Close()
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

func (c *Client) downloadDir(remoteDir, localDir string) error {
    // 规范化远端路径为 POSIX 风格
    base := path.Clean(strings.ReplaceAll(remoteDir, "\\", "/"))
    w := c.raw.Walk(remoteDir)
    for w.Step() {
        if w.Err() != nil {
            return w.Err()
        }
        p := path.Clean(strings.ReplaceAll(w.Path(), "\\", "/"))
        var rel string
        if p == base {
            rel = ""
        } else if strings.HasPrefix(p, base+"/") {
            rel = strings.TrimPrefix(p, base+"/")
        } else {
            // 回退策略，尽量避免错误嵌套
            rel = strings.TrimPrefix(p, base)
            rel = strings.TrimPrefix(rel, "/")
        }

        target := filepath.Join(localDir, filepath.FromSlash(rel))
        if w.Stat().IsDir() {
            if err := os.MkdirAll(target, 0o755); err != nil {
                return err
            }
            continue
        }
        if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
            return err
        }
        rf, err := c.raw.Open(w.Path())
        if err != nil {
            return err
        }
        lf, err := os.Create(target)
        if err != nil {
            rf.Close()
            return err
        }
        if _, err := io.Copy(lf, rf); err != nil {
            lf.Close(); rf.Close()
            return err
        }
        lf.Close(); rf.Close()
        fmt.Printf("下载: %s -> %s\n", w.Path(), target)
    }
    fmt.Printf("目录下载完成: %s -> %s\n", remoteDir, localDir)
    return nil
}

// UploadAuto: 根据本地路径类型（文件/目录）与远端目标（目录/文件）自动选择上传策略。
// - 本地是文件：若远端存在为目录或以分隔符结尾，上传到该目录下；否则按给定远端文件路径上传。
// - 本地是目录：将整个目录上传为 远端目录/<basename(本地)>，保留目录结构。
func (c *Client) UploadAuto(local, remote string) error {
    st, err := os.Stat(local)
    if err != nil {
        return fmt.Errorf("stat local: %w", err)
    }
    remote = toPosix(remote)
    if st.IsDir() {
        base := filepath.Base(local)
        remoteRoot := path.Join(remote, base)
        if err := c.raw.MkdirAll(remoteRoot); err != nil {
            return fmt.Errorf("mkdir remote: %w", err)
        }
        return c.uploadDir(local, remoteRoot)
    }
    // 文件
    isDirHint := strings.HasSuffix(remote, "/") || strings.HasSuffix(remote, "\\")
    if isDir, _ := c.isRemoteDir(remote); isDir || isDirHint {
        remote = path.Join(strings.TrimRight(remote, "/\\"), filepath.Base(local))
    }
    if err := c.raw.MkdirAll(path.Dir(remote)); err != nil {
        return fmt.Errorf("mkdir remote: %w", err)
    }
    lf, err := os.Open(local)
    if err != nil {
        return fmt.Errorf("open local: %w", err)
    }
    defer lf.Close()
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

func (c *Client) uploadDir(localDir, remoteRoot string) error {
    return filepath.WalkDir(localDir, func(p string, d os.DirEntry, err error) error {
        if err != nil {
            return err
        }
        rel, _ := filepath.Rel(localDir, p)
        rel = filepath.ToSlash(rel)
        if rel == "." {
            return nil
        }
        remotePath := path.Clean(path.Join(remoteRoot, rel))
        if d.IsDir() {
            if err := c.raw.MkdirAll(remotePath); err != nil {
                return err
            }
            return nil
        }
        lf, err := os.Open(p)
        if err != nil {
            return err
        }
        defer lf.Close()
        if err := c.raw.MkdirAll(path.Dir(remotePath)); err != nil {
            return err
        }
        rf, err := c.raw.Create(remotePath)
        if err != nil {
            return err
        }
        if _, err := io.Copy(rf, lf); err != nil {
            rf.Close()
            return err
        }
        rf.Close()
        fmt.Printf("上传: %s -> %s\n", p, remotePath)
        return nil
    })
}

func (c *Client) isRemoteDir(p string) (bool, error) {
    fi, err := c.raw.Stat(toPosix(p))
    if err != nil {
        return false, err
    }
    return fi.IsDir(), nil
}

func toPosix(p string) string {
    return path.Clean(strings.ReplaceAll(p, "\\", "/"))
}
