package version1

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

const ConfigDir = ".sshctl"
const ConfigFile = "config.json"

var (
	configPath string
	once       sync.Once
)

type connection struct {
	Host string `json:"host"`
	User string `json:"user"`
	Port int    `json:"port"`
	Key  string `json:"key"` // 支持 ~
	Desc string `json:"desc"`
}

type Config struct {
	Connections map[string]connection `json:"connections"`
}

func init() {
	once.Do(func() {
		home, _ := os.UserHomeDir()
		configPath = filepath.Join(home, ConfigDir, ConfigFile)
	})
}

func loadConfig() (*Config, error) {
	_, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		return &Config{Connections: make(map[string]connection)}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Connections == nil {
		cfg.Connections = make(map[string]connection)
	}
	return &cfg, nil
}

func SaveConfig(cfg *Config) error {
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}

func addConnection(name string, conn connection) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	conn.Key = ExpandHome(conn.Key) // 处理 ~
	cfg.Connections[name] = conn
	return SaveConfig(cfg)
}

func getConnection(name string) (connection, error) {
	cfg, err := loadConfig()
	if err != nil {
		return connection{}, err
	}
	conn, exists := cfg.Connections[name]
	if !exists {
		return connection{}, os.ErrNotExist
	}
	return conn, nil
}
