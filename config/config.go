package config

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
)

var config Configuration

type Configuration struct {
	Auth struct {
		Secret string `json:"secret"`
		Salt   string `json:"salt"`
	} `json:"user"`
	Server struct {
		Metrics string `json:"metrics"`
		Service string `json:"service"`
		Port    string `json:"port"`
	} `json:"server"`
	Database struct {
		Hostname   string `json:"hostname"`
		Port       string `json:"port"`
		Username   string `json:"username"`
		Password   string `json:"password"`
		Database   string `json:"database"`
		Migrations string `json:"migrations"`
	} `json:"database"`
	Log struct {
		Path string `json:"path"`
	} `json:"log"`
}

func Config() Configuration {
	return config
}

func ReadConfig() Configuration {
	configFile := "config.json"
	f, err := os.Open(configFile)
	if err != nil {
		slog.Error(fmt.Sprintf("Could not open config file: %s", configFile), err)
	}

	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			slog.Warn("Unable to close config file")
		}
	}(f)

	var cfg Configuration
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		msg := "Failed to parse config file."
		slog.Error(msg)
		log.Fatal(msg)
	}
	config = cfg
	return cfg
}
