package main

import (
	"os"
	"path/filepath"

	"sigs.k8s.io/yaml"
)

type config struct {
	Port    int      `json:"port"`
	Modules []string `json:"modules"`
}

func loadConfig(filename string) (config, error) {
	var cfg config

	data, err := os.ReadFile(filepath.Clean(filename))
	if err != nil {
		return config{}, err
	}

	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return config{}, err
	}

	return cfg, nil
}
