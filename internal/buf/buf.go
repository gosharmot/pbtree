package buf

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

var externalPlugins = map[string]struct{}{
	"go":   {},
	"grpc": {},
}

// Parse try parse transferred file to GenYAML
func Parse(f string) (GenYAML, error) {
	open, err := os.Open(f)
	if err != nil {
		if os.IsNotExist(err) {
			return GenYAML{}, fmt.Errorf("file '%s' not found", f)
		}
		return GenYAML{}, err
	}

	var res GenYAML
	if err = yaml.NewDecoder(open).Decode(&res); err != nil {
		return GenYAML{}, err
	}

	return res, nil
}

// GenYAML is representation of configuration file used by the buf generate command
type GenYAML struct {
	Version string   `yaml:"version" json:"version,omitempty"`
	Plugins []Plugin `yaml:"plugins" json:"plugins,omitempty"`
}

// Plugin is representation of buf plugin definition
type Plugin struct {
	Name     string   `yaml:"name"`
	Path     string   `yaml:"path"`
	Out      string   `yaml:"out"`
	Opt      []string `yaml:"opt"`
	Strategy string   `yaml:"strategy"`
}

// ExternalPluginsOnly return buf.gen.yaml content with
func (g GenYAML) ExternalPluginsOnly() ([]byte, error) {
	p := make([]Plugin, 0, 2)
	for _, plugin := range g.Plugins {
		if _, ok := externalPlugins[plugin.Name]; ok {
			p = append(p, plugin)
		}
	}
	return yaml.Marshal(GenYAML{Version: g.Version, Plugins: p})
}

// MFlags returns map[proto]dstPath from -M options [go, grpc] plugins only
func (g GenYAML) MFlags() (map[string]string, error) {
	res := make(map[string]string, len(g.Plugins))

	for _, plugin := range g.Plugins {
		for _, opt := range plugin.Opt {
			if mFlag, ok := strings.CutPrefix(opt, "M"); ok {
				split := strings.Split(mFlag, "=")
				if len(split) != 2 {
					return nil, fmt.Errorf("invalid option %q", mFlag)
				}
				res[split[0]] = split[1]
			}
		}
	}

	return res, nil
}
