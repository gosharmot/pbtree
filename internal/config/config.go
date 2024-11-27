package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Template is file with list of proto
var Template = []byte("local_proto: []\nexternal_proto: []")

// Config is representation of Template file
type Config struct {
	LocalProto    []string `yaml:"local_proto"`
	ExternalProto []string `yaml:"external_proto"`
}

// ParseFromReader parse Config from io.Reader
func ParseFromReader(r io.Reader) (Config, error) {
	var c Config
	if err := yaml.NewDecoder(r).Decode(&c); err != nil {
		return Config{}, err
	}
	return c, nil
}

// Parse try parse transferred file to Config
func Parse(f string) (Config, error) {
	open, err := os.Open(f)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			_, name := filepath.Split(f)
			return Config{}, fmt.Errorf("file '%s' not found. run pbtree init", name)
		}
		return Config{}, err
	}
	defer func() { _ = open.Close() }()

	return ParseFromReader(open)
}

// Marshal is yaml marshall Config
func (c Config) Marshal() ([]byte, error) {
	return yaml.Marshal(c)
}
