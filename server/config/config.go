package config

import "github.com/ilyakaznacheev/cleanenv"

type ConfigParser struct {
	Host   string `yaml:"host" env:"HOST_TOR" env-default:"http://rutor.info"`
	ApiKey string `yaml:"api_key" env:"PORT_TOR" env-default:""`
}

var cfg ConfigParser

func ReadConfigParser(vars string) (string, error) {
	err := cleanenv.ReadConfig("config.yml", &cfg)
	if err == nil {
		switch {
		case vars == "Host":
			return cfg.Host, nil
		case vars == "ApiKey":
			return cfg.ApiKey, nil
		}
	}
	return "", err
}
