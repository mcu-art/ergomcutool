package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

var Version = "1.0.0"

type GlobalConfigT struct {
	Host         string `yaml:"host"`
	Port         string `yaml:"port"`
	BaseURL      string `yaml:"base_url"`
	AssetsDir    string `yaml:"assets_directory"`
	TemplatesDir string `yaml:"templates_directory"`
}

var GlobalConfig = &GlobalConfigT{}

func ReadGlobalConfigFile(file string) error {

	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	serverConfig := &GlobalConfigT{}
	err = yaml.Unmarshal(data, serverConfig)
	if err != nil {
		return err
	}

	return err
}
