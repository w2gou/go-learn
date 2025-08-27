package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type Connection struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password,omitempty"`
	KeyPath  string `json:"key_path,omitempty"`
}

func getConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".gossh")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func LoadConnections() ([]Connection, error) {
	path, err := getConfigPath()
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return []Connection{}, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var conns []Connection
	if err := json.Unmarshal(data, &conns); err != nil {
		return nil, err
	}
	return conns, nil
}

func SaveConnections(conns []Connection) error {
	path, err := getConfigPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(conns, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func AddConnection(conn Connection) error {
	conns, err := LoadConnections()
	if err != nil {
		return err
	}
	for _, c := range conns {
		if c.Name == conn.Name {
			return errors.New("连接名已存在")
		}
	}
	conns = append(conns, conn)
	return SaveConnections(conns)
}
