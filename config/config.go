package config

import (
	"gopkg.in/yaml.v2"
	"log"
	"os"
)

type Config struct {
	Server struct {
		Port string `yaml:"port"`
		Host string `yaml:"host"`
		Root string `yaml:"root"`
	} `yaml:"server"`
	Database struct {
		Hostname   string `yaml:"hostname"`
		Port       string `yaml:"port"`
		Username   string `yaml:"username"`
		Password   string `yaml:"password"`
		Database   string `yaml:"database"`
		Migrations string `yaml:"migrations"`
	} `yaml:"database"`
}

func ReadConfig() Config {
	f, err := os.Open("config.yml")
	if err != nil {
		log.Println(err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Println("Unable to close config file")
		}
	}(f)

	var cfg Config
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		log.Println(err)
	}

	return cfg
}
