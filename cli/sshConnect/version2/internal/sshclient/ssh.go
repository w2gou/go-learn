package sshclient

import (
    "bufio"
    "bytes"
    "fmt"
    "go-learn/cli/sshConnect/version2/config"
    "go-learn/cli/sshConnect/version2/internal/sftpclient"
    "golang.org/x/crypto/ssh"
    "golang.org/x/term"
    "io"
    "path"
    "os"
    "runtime"
    "strings"
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

    // 仅绑定输出到本地；输入我们自己转发以拦截本地命令
    session.Stdout = os.Stdout
    session.Stderr = os.Stderr

    stdinPipe, err := session.StdinPipe()
    if err != nil {
        return fmt.Errorf("获取会话输入管道失败: %w", err)
    }

    // 尝试设置远端 UTF-8，降低中文乱码概率
    setRemoteUTF8(session)

    // 在 Windows 控制台启用 ANSI 渲染（颜色/样式）
    enableANSIOnWindows()

    // 提示本地命令用法
    fmt.Fprintln(os.Stderr, "本地命令: 在行首输入 :upload <本地路径> <远端路径> 或 :download <远端路径> <本地路径>；输入 :help 查看帮助")

    // 启动本地->远端的输入转发（含本地命令拦截）
    go interceptAndForwardInput(os.Stdin, stdinPipe, client)

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

// interceptAndForwardInput 在行首拦截以 ':' 开头的本地命令，其余字节透传到远端 shell。
func interceptAndForwardInput(localIn io.Reader, remoteIn io.Writer, sshClient *ssh.Client) {
    r := bufio.NewReader(localIn)
    atLineStart := true
    for {
        b, err := r.ReadByte()
        if err != nil {
            return
        }

        if atLineStart && b == ':' {
            // 本地命令：回显 ':'，并进入本地行编辑模式（带回显、退格处理）
            _, _ = os.Stdout.Write([]byte{':'})
            cmdline, canceled := readLocalCommandInteractive(r, sshClient)
            if !canceled {
                handleLocalCommand(cmdline, sshClient)
            }
            atLineStart = true
            continue
        }

        // 透传到远端 shell
        if _, err := remoteIn.Write([]byte{b}); err != nil {
            return
        }
        if b == '\n' || b == '\r' {
            atLineStart = true
        } else {
            atLineStart = false
        }
    }
}

// readLineRaw 按字节读取一行，支持以 \n 或 \r 或 \r\n 结尾。
func readLineRaw(r *bufio.Reader) (string, error) {
    var buf []byte
    for {
        b, err := r.ReadByte()
        if err != nil {
            return string(buf), err
        }
        if b == '\n' || b == '\r' {
            if b == '\r' {
                if next, _ := r.Peek(1); len(next) == 1 && next[0] == '\n' {
                    _, _ = r.ReadByte()
                }
            }
            return string(buf), nil
        }
        buf = append(buf, b)
    }
}

// readLocalCommandInteractive 读取并回显一行本地命令，支持退格与回车，Ctrl-C 取消。
// 返回 (命令行, 是否取消)。
func readLocalCommandInteractive(r *bufio.Reader, sshClient *ssh.Client) (string, bool) {
    var buf []byte
    for {
        b, err := r.ReadByte()
        if err != nil {
            // 输入结束，按完成处理
            _, _ = os.Stdout.Write([]byte{'\n'})
            return string(buf), false
        }
        switch b {
        case '\t': // Tab 补全
            newBuf, repaint := completeLocalCommand(buf, sshClient)
            if len(repaint) > 0 {
                _, _ = os.Stdout.Write([]byte(repaint))
            }
            buf = newBuf
        case '\r', '\n':
            if b == '\r' {
                if next, _ := r.Peek(1); len(next) == 1 && next[0] == '\n' {
                    _, _ = r.ReadByte()
                }
            }
            _, _ = os.Stdout.Write([]byte{'\n'})
            return string(buf), false
        case 0x03: // Ctrl-C
            // 显示 ^C 并取消本地命令
            _, _ = os.Stdout.Write([]byte{'^', 'C', '\n'})
            return "", true
        case 0x7f, 0x08: // Backspace / DEL
            if len(buf) > 0 {
                buf = buf[:len(buf)-1]
                // 回显删除（光标左移、空格覆盖、再左移）
                _, _ = os.Stdout.Write([]byte{'\b', ' ', '\b'})
            }
        default:
            buf = append(buf, b)
            _, _ = os.Stdout.Write([]byte{b})
        }
    }
}

func handleLocalCommand(cmdline string, sshClient *ssh.Client) {
    cmdline = strings.TrimSpace(cmdline)
    if strings.HasPrefix(cmdline, ":") {
        cmdline = strings.TrimSpace(cmdline[1:])
    }
    fields := splitArgs(cmdline)
    if len(fields) == 0 {
        return
    }
    switch strings.ToLower(fields[0]) {
    case "help":
        fallthrough
    case "?":
        fmt.Fprintln(os.Stderr, ":upload <本地路径> <远端目录或文件>  — 上传文件或目录（目录会作为子目录创建）")
        fmt.Fprintln(os.Stderr, ":download <远端路径> [本地路径或目录] — 下载文件或目录；省略本地路径则存到当前目录")
        fmt.Fprintln(os.Stderr, ":help 或 :? — 显示帮助")
        return
    case "upload":
        if len(fields) != 3 {
            fmt.Fprintln(os.Stderr, "用法: :upload <本地路径> <远端目录或文件>")
            return
        }
        c, err := sftpclient.New(sshClient)
        if err != nil {
            fmt.Fprintln(os.Stderr, "SFTP 初始化失败:", err)
            return
        }
        defer c.Close()
        if err := c.UploadAuto(fields[1], fields[2]); err != nil {
            fmt.Fprintln(os.Stderr, "上传失败:", err)
        }
        return
    case "download":
        if len(fields) < 2 || len(fields) > 3 {
            fmt.Fprintln(os.Stderr, "用法: :download <远端路径> [本地路径或目录]")
            return
        }
        c, err := sftpclient.New(sshClient)
        if err != nil {
            fmt.Fprintln(os.Stderr, "SFTP 初始化失败:", err)
            return
        }
        defer c.Close()
        remote := fields[1]
        // 远端相对路径：按照远端 shell 的当前目录解析
        if !isRemoteAbs(remote) {
            if cwd, err := remoteCWD(sshClient); err == nil && cwd != "" {
                remote = path.Clean(path.Join(cwd, remote))
            }
        }
        local := ""
        if len(fields) == 3 {
            local = fields[2]
        } else {
            // 仅给了远端路径，则下载到当前工作目录（取文件名），或目录递归到当前目录
            local = "./"
        }
        if err := c.Download(remote, local); err != nil {
            fmt.Fprintln(os.Stderr, "下载失败:", err)
        }
        return
    default:
        fmt.Fprintln(os.Stderr, "未知本地命令，输入 :help 查看用法")
        return
    }
}

// splitArgs 将命令行按空白拆分，并支持用引号包裹的参数。
func splitArgs(s string) []string {
    var args []string
    var cur bytes.Buffer
    inQuote := byte(0)
    esc := false
    for i := 0; i < len(s); i++ {
        ch := s[i]
        if esc {
            cur.WriteByte(ch)
            esc = false
            continue
        }
        if ch == '\\' {
            esc = true
            continue
        }
        if inQuote != 0 {
            if ch == inQuote {
                inQuote = 0
            } else {
                cur.WriteByte(ch)
            }
            continue
        }
        switch ch {
        case '\'', '"':
            inQuote = ch
        case ' ', '\t':
            if cur.Len() > 0 {
                args = append(args, cur.String())
                cur.Reset()
            }
        default:
            cur.WriteByte(ch)
        }
    }
    if cur.Len() > 0 {
        args = append(args, cur.String())
    }
    return args
}

// 判断远端路径是否为绝对路径（以 / 或 ~ 开头）。
func isRemoteAbs(p string) bool {
    return strings.HasPrefix(p, "/") || strings.HasPrefix(p, "~")
}

// 通过新会话读取远端当前工作目录（与交互式 shell 一致）。
func remoteCWD(client *ssh.Client) (string, error) {
    s, err := client.NewSession()
    if err != nil {
        return "", err
    }
    defer s.Close()
    var out bytes.Buffer
    s.Stdout = &out
    if err := s.Run("pwd"); err != nil {
        return "", err
    }
    return strings.TrimSpace(out.String()), nil
}

func remoteHOME(client *ssh.Client) (string, error) {
    s, err := client.NewSession()
    if err != nil {
        return "", err
    }
    defer s.Close()
    var out bytes.Buffer
    s.Stdout = &out
    if err := s.Run("printf %s \"$HOME\""); err != nil {
        return "", err
    }
    return strings.TrimSpace(out.String()), nil
}

// 完成本地命令中的远端路径补全（只在 download 的第1个参数、upload 的第2个参数）。
// 返回新的缓冲内容与需要回显的字符（如追加补全或换行+列表+重绘）。
func completeLocalCommand(buf []byte, client *ssh.Client) ([]byte, string) {
    s := string(buf)
    // 计算 tokens，并判断当前是否在输入新参数
    tokens := splitArgs(s)
    typingNew := len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t')

    if len(tokens) == 0 {
        return buf, ""
    }
    cmd := strings.ToLower(tokens[0])
    // 需要补全远端路径的参数索引（从0开始，不含命令本身）
    targetIdx := -1
    switch cmd {
    case "download":
        targetIdx = 1 // tokens[0] 是命令名，远端路径在 tokens[1]
    case "upload":
        targetIdx = 2 // upload 的远端路径在 tokens[2]
    default:
        return buf, ""
    }

    // 当前正在编辑的参数索引
    curIdx := len(tokens) - 1
    if typingNew {
        curIdx++
    }
    if curIdx != targetIdx {
        return buf, "" // 只在远端参数位置才补全
    }

    // 当前片段（可能为空）
    var fragment string
    var before string
    if typingNew {
        fragment = ""
        before = s
        if len(before) > 0 {
            // 保证只保留一个空格分隔
            if before[len(before)-1] != ' ' {
                before += " "
            }
        }
    } else {
        // 取最后一个 token 的原始片段位置：简单处理，只按空格查找
        lastSpace := strings.LastIndexAny(s, " \t")
        if lastSpace >= 0 {
            before = s[:lastSpace+1]
            fragment = s[lastSpace+1:]
        } else {
            before = ""
            fragment = s
        }
    }

    // 解析片段为目录+前缀（基于远端路径规则 '/'' 分隔）。
    slash := strings.LastIndex(fragment, "/")
    dirPart := ""
    prefix := fragment
    if strings.HasPrefix(fragment, "/") {
        if slash == 0 {
            dirPart = "/"
            prefix = fragment[1:]
        } else if slash > 0 {
            dirPart = fragment[:slash]
            prefix = fragment[slash+1:]
        }
    } else if slash >= 0 {
        dirPart = fragment[:slash]
        prefix = fragment[slash+1:]
    }

    // 解析远端基准目录
    baseDir := dirPart
    if baseDir == "" {
        if cwd, err := remoteCWD(client); err == nil && cwd != "" {
            baseDir = cwd
        } else {
            baseDir = "/"
        }
    } else if strings.HasPrefix(baseDir, "~") {
        if home, err := remoteHOME(client); err == nil && home != "" {
            if baseDir == "~" {
                baseDir = home
            } else if strings.HasPrefix(baseDir, "~/") {
                baseDir = path.Clean(path.Join(home, baseDir[2:]))
            }
        }
    } else if !strings.HasPrefix(baseDir, "/") {
        if cwd, err := remoteCWD(client); err == nil && cwd != "" {
            baseDir = path.Clean(path.Join(cwd, baseDir))
        }
    }

    // 列出目录，匹配前缀
    c, err := sftpclient.New(client)
    if err != nil {
        return buf, ""
    }
    defer c.Close()

    list, err := c.ReadDir(baseDir)
    if err != nil {
        return buf, ""
    }
    var cand []string
    for _, fi := range list {
        name := fi.Name()
        if strings.HasPrefix(name, prefix) {
            if fi.IsDir() {
                name += "/"
            }
            cand = append(cand, name)
        }
    }
    if len(cand) == 0 {
        return buf, "\a"
    }
    // 单一候选：直接补全
    if len(cand) == 1 {
        completed := cand[0]
        newFrag := completed
        var newS string
        if dirPart != "" {
            if dirPart == "/" {
                newS = before + dirPart + newFrag
            } else {
                newS = before + dirPart + "/" + newFrag
            }
        } else {
            newS = before + newFrag
        }
        // 仅回显追加部分
        echo := newS[len(s):]
        return []byte(newS), echo
    }

    // 多候选：取公共前缀，若可扩展则扩展；否则列出候选。
    common := cand[0]
    for _, v := range cand[1:] {
        common = commonPrefix(common, v)
        if common == "" {
            break
        }
    }
    if len(common) > len(prefix) {
        // 可扩展公共前缀
        var newS string
        add := common[len(prefix):]
        newS = s + add
        return []byte(newS), add
    }

    // 输出候选列表并重绘当前行
    var b strings.Builder
    b.WriteByte('\n')
    for i, v := range cand {
        if i > 0 {
            b.WriteByte(' ')
        }
        b.WriteString(v)
    }
    b.WriteByte('\n')
    b.WriteString(":")
    b.WriteString(s)
    return buf, b.String()
}

func commonPrefix(a, b string) string {
    n := len(a)
    if len(b) < n {
        n = len(b)
    }
    i := 0
    for i < n && a[i] == b[i] {
        i++
    }
    return a[:i]
}
