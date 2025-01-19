package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Redis struct {
		Addrs      []string `mapstructure:"addrs" yaml:"addrs"`
		Password   string   `mapstructure:"password" yaml:"password"`
		ClientName string   `mapstructure:"client-name" yaml:"client-name"`
	} `mapstructure:"redis" yaml:"redis"`
	Server struct {
		HttpPort string `mapstructure:"http-port" yaml:"http-port"`
		WSPort   string `mapstructure:"ws-port" yaml:"ws-port"`
	} `mapstructure:"server" yaml:"server"`
}

func LoadConfig(path string) *Config {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Error opening config file:", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println("Error closing config file:", err)
		}
	}(file)
	cfg := &Config{}
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(cfg)
	if err != nil {
		fmt.Println("Error parsing config file:", err)
	}
	return cfg
}
