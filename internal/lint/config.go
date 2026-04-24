package lint

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Include       []string       `yaml:"include"`
	Exclude       []string       `yaml:"exclude"`
	MaxLinesByExt map[string]int `yaml:"max_lines_by_ext"`
	Commands      []Command      `yaml:"commands"`
}

type Command struct {
	Name string   `yaml:"name"`
	Cmd  string   `yaml:"cmd"`
	Args []string `yaml:"args"`
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
