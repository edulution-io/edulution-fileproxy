package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Log struct {
		Level string `yaml:"level"`
		File  string `yaml:"file"`
	} `yaml:"log"`
	LDAP struct {
		Server             string `yaml:"server"`
		InsecureSkipVerify bool   `yaml:"insecure_skip_verify"`
		Domain             string `yaml:"domain"`
	} `yaml:"ldap"`
	SMB struct {
		Server    string `yaml:"server"`
		ShareName string `yaml:"share_name"`
	} `yaml:"smb"`
	HTTP struct {
		Address      string `yaml:"address"`
		WebDAVPrefix string `yaml:"webdav_prefix"`
		APIPrefix    string `yaml:"api_prefix"`
		CertFile     string `yaml:"cert_file"`
		KeyFile      string `yaml:"key_file"`
	} `yaml:"http"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
