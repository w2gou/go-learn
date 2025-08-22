package version1

import (
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

type GoSshConnectModel struct{}

func (m *GoSshConnectModel) StartServer() error {
	log.Printf("启动ssh连接器")

	// 启动 HTTP 服务等逻辑
	startServer()

	return nil
}

func startServer() {
	command()
}

func command() {
	app := &cli.App{
		Name:  "sshctl",
		Usage: "SSH 控制工具：执行命令、上传下载文件",
		Commands: []*cli.Command{
			{
				Name:      "add",
				Usage:     "添加一个连接 sshctl add <name> --host=... --user=...",
				ArgsUsage: "<name>",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "host", Required: true},
					&cli.StringFlag{Name: "user", Required: true},
					&cli.IntFlag{Name: "port", Value: 22},
					&cli.StringFlag{Name: "key", Value: "~/.ssh/id_rsa"},
					&cli.StringFlag{Name: "desc"},
				},
				Action: func(c *cli.Context) error {
					name := c.Args().Get(0)
					if name == "" {
						return cli.ShowCommandHelp(c, "add")
					}
					return addConnection(name, connection{
						Host: c.String("host"),
						User: c.String("user"),
						Port: c.Int("port"),
						Key:  c.String("key"),
						Desc: c.String("desc"),
					})
				},
			},
			{
				Name:  "list",
				Usage: "列出所有连接",
				Action: func(c *cli.Context) error {
					cfg, err := loadConfig()
					if err != nil {
						return err
					}
					for name, conn := range cfg.Connections {
						println(name + "\t" + conn.User + "@" + conn.Host + ":" + string(rune(conn.Port)) + "\t" + conn.Desc)
					}
					return nil
				},
			},
			{
				Name:      "exec",
				Usage:     "执行命令 sshctl exec <name> <command>",
				ArgsUsage: "<name> <command>",
				Action: func(c *cli.Context) error {
					name := c.Args().Get(0)
					cmd := c.Args().Get(1)
					if name == "" || cmd == "" {
						return cli.ShowCommandHelp(c, "exec")
					}
					conn, err := getConnection(name)
					if err != nil {
						return err
					}
					client, err := connect(conn)
					if err != nil {
						return err
					}
					defer client.Close()
					return execCommand(client, cmd)
				},
			},
			{
				Name:      "put",
				Usage:     "上传文件 sshctl put <name> <local> <remote>",
				ArgsUsage: "<name> <local> <remote>",
				Action: func(c *cli.Context) error {
					name, local, remote := c.Args().Get(0), c.Args().Get(1), c.Args().Get(2)
					if name == "" || local == "" || remote == "" {
						return cli.ShowCommandHelp(c, "put")
					}
					conn, err := getConnection(name)
					if err != nil {
						return err
					}
					sshClient, err := connect(conn)
					if err != nil {
						return err
					}
					defer sshClient.Close()
					sftpClient, err := newClient(sshClient)
					if err != nil {
						return err
					}
					defer sftpClient.Close()
					return upload(sftpClient, local, remote)
				},
			},
			{
				Name:      "get",
				Usage:     "下载文件 sshctl get <name> <remote> <local>",
				ArgsUsage: "<name> <remote> <local>",
				Action: func(c *cli.Context) error {
					name, remote, local := c.Args().Get(0), c.Args().Get(1), c.Args().Get(2)
					if name == "" || remote == "" || local == "" {
						return cli.ShowCommandHelp(c, "get")
					}
					conn, err := getConnection(name)
					if err != nil {
						return err
					}
					sshClient, err := connect(conn)
					if err != nil {
						return err
					}
					defer sshClient.Close()
					sftpClient, err := newClient(sshClient)
					if err != nil {
						return err
					}
					defer sftpClient.Close()
					return download(sftpClient, remote, local)
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
