package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

var Conf = new(Config)

type MySQLConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

type PanelServiceConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type DockerClientConfig struct {
	Host string `yaml:"host"`
}

type Config struct {
	MySQL        MySQLConfig        `yaml:"mysql"`
	PanelService PanelServiceConfig `yaml:"panel_service"`
	DockerClient DockerClientConfig `yaml:"docker_client"`
}

func LoadConfig() {
	yamlFile, err := ioutil.ReadFile("config/config.yaml")
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(yamlFile, Conf)
	if err != nil {
		panic(err)
	}
}
