package sshclient

import (
    "fmt"
    "go-learn/cli/sshConnect/version2/config"
    "golang.org/x/crypto/ssh"
    "golang.org/x/term"
    "os"
    "runtime"
    "syscall"
    "unsafe"
)

const (
    defaultTermWidth  = 80
    defaultTermHeight = 24
)

// InteractiveShell 建立 SSH 连接并进入交互式 Shell。
func InteractiveShell(conn config.Connection) error {
    cfg := buildClientConfig(conn)

    // 终端置为 raw 模式，便于交互体验（Ctrl-C、Backspace 等直传）
    fd := int(os.Stdin.Fd())
    restore, err := makeStdinRaw(fd)
    if err != nil {
        return err
    }
    defer restore()

    client, err := dial(conn, cfg)
    if err != nil {
        return fmt.Errorf("SSH 连接失败: %w", err)
    }
    defer client.Close()

    session, err := client.NewSession()
    if err != nil {
        return fmt.Errorf("创建会话失败: %w", err)
    }
    defer session.Close()

    // 申请 PTY（带颜色和正确尺寸）
    termName := detectTERM()
    w, h := currentSizeOrDefault(fd)
    if err := requestPTY(session, termName, w, h); err != nil {
        return fmt.Errorf("请求 PTY 失败: %w", err)
    }

    // 绑定本地 IO 与远端会话
    bindSessionIO(session)

    // 尝试设置远端 UTF-8，降低中文乱码概率
    setRemoteUTF8(session)

    // 在 Windows 控制台启用 ANSI 渲染（颜色/样式）
    enableANSIOnWindows()

    if err := session.Shell(); err != nil {
        return fmt.Errorf("启动远程 Shell 失败: %w", err)
    }
    return session.Wait()
}

// --- helpers ---

func buildClientConfig(conn config.Connection) *ssh.ClientConfig {
    return &ssh.ClientConfig{
        User: conn.User,
        Auth: []ssh.AuthMethod{
            ssh.Password(conn.Password),
        },
        HostKeyCallback: ssh.InsecureIgnoreHostKey(),
    }
}

func makeStdinRaw(fd int) (restore func(), err error) {
    oldState, err := term.MakeRaw(fd)
    if err != nil {
        return nil, fmt.Errorf("不能将 stdin 设为 raw: %w", err)
    }
    return func() { _ = term.Restore(fd, oldState) }, nil
}

func dial(conn config.Connection, cfg *ssh.ClientConfig) (*ssh.Client, error) {
    address := fmt.Sprintf("%s:%d", conn.Host, conn.Port)
    return ssh.Dial("tcp", address, cfg)
}

func detectTERM() string {
    if t := os.Getenv("TERM"); t != "" {
        return t
    }
    return "xterm-256color"
}

func currentSizeOrDefault(fd int) (w, h int) {
    w, h, err := term.GetSize(fd)
    if err != nil {
        // 获取失败则使用默认大小
        return defaultTermWidth, defaultTermHeight
    }
    return w, h
}

func requestPTY(s *ssh.Session, termName string, width, height int) error {
    modes := ssh.TerminalModes{
        ssh.ECHO:          1,
        ssh.TTY_OP_ISPEED: 14400,
        ssh.TTY_OP_OSPEED: 14400,
    }
    // 注意顺序: height, width
    return s.RequestPty(termName, height, width, modes)
}

func bindSessionIO(s *ssh.Session) {
    s.Stdin = os.Stdin
    s.Stdout = os.Stdout
    s.Stderr = os.Stderr
}

func setRemoteUTF8(s *ssh.Session) {
    _ = s.Setenv("LANG", "C.UTF-8")
    _ = s.Setenv("LC_ALL", "C.UTF-8")
}

// enableANSIOnWindows 启用 Windows 控制台 ANSI 转义处理（颜色、样式等）
func enableANSIOnWindows() {
    if runtime.GOOS != "windows" {
        return
    }
    // Windows 10+ 需要开启虚拟终端处理
    const enableVirtualTerminalProcessing = 0x0004

    kernel32 := syscall.NewLazyDLL("kernel32.dll")
    getStdHandle := kernel32.NewProc("GetStdHandle")
    getConsoleMode := kernel32.NewProc("GetConsoleMode")
    setConsoleMode := kernel32.NewProc("SetConsoleMode")

    // -11: STD_OUTPUT_HANDLE, -12: STD_ERROR_HANDLE
    for _, std := range []int32{-11, -12} {
        h, _, _ := getStdHandle.Call(uintptr(std))
        if h == 0 || h == ^uintptr(0) {
            continue
        }
        var mode uint32
        r1, _, _ := getConsoleMode.Call(h, uintptr(unsafe.Pointer(&mode)))
        if r1 == 0 {
            continue
        }
        _, _, _ = setConsoleMode.Call(h, uintptr(mode|enableVirtualTerminalProcessing))
    }
}

