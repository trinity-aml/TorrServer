package config

import "github.com/ilyakaznacheev/cleanenv"

type ConfParser struct {
	Trackers          []string `yaml:"trackers"`
	Default_url       []string `yaml:"default_url"`
	Blacklist_tracker []string `yaml:"blacklist_tracker"`
}

var cfg ConfParser

func ReadConfigParser(vars string) ([]string, error) {
	err := cleanenv.ReadConfig("config.yml", &cfg)
	if err == nil {
		switch {
		case vars == "Trackers":
			return cfg.Trackers, nil
		case vars == "Default_url":
			return cfg.Default_url, nil
		case vars == "Blacklist_tracker":
			return cfg.Blacklist_tracker, nil
		}
	}
	return nil, err
}
