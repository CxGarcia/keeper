package provider_registry

import (
	_ "embed"

	"gopkg.in/yaml.v3"
)

//go:embed provider-registry.yml
var registry []byte

type Registry struct {
	Providers []Provider `yaml:"providers"`
}

type Provider struct {
	Name         string       `yaml:"name"`
	BaseURL      string       `yaml:"base_url"`
	DefaultModel string       `yaml:"default_model"`
	Models       []Model      `yaml:"models"`
	Auth         ProviderAuth `yaml:"auth"`
}

type Model struct {
	Name string `yaml:"name"`
}

type ProviderAuth struct {
	Type  string `yaml:"type"`
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

func New() (Registry, error) {
	return loadRegistry()
}

func loadRegistry() (Registry, error) {
	var reg Registry
	if err := yaml.Unmarshal(registry, &reg); err != nil {
		return Registry{}, err
	}

	return reg, nil
}
