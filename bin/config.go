package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/spf13/viper"
)

const (
	// DefaultConfigFile specifies the default config location
	DefaultConfigFile = "./assistant.yaml"
)

var (
	_ConfigFile string
)

// LoadConfig loads in the configuration files for the assistant
func LoadConfig(r io.Reader) (*viper.Viper, error) {
	cfg := viper.New()
	cfg.SetConfigType("yaml")
	return cfg, cfg.ReadConfig(r)
}

// LoadConfigFile loads a config from a specific path.
func LoadConfigFile(path string) (*viper.Viper, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open config file %s: %w", path, err)
	}

	return LoadConfig(f)
}

func ParseFlags() {
	flag.StringVar(&_ConfigFile, "config", DefaultConfigFile, "A path to the configuration file")
	flag.StringVar(&_ConfigFile, "c", DefaultConfigFile, "A path to the configuration file")

	flag.Parse()
}
