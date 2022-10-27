package main

import (
	"errors"
	"os"

	yaml "gopkg.in/yaml.v3"
)

type Pages struct {
	Blocks []string
}

type Config struct {
	Pages        []Pages `yaml:"pages"`
	Interval     int     `yaml:"interval"`
	CacheTimeout int     `yaml:"cachetimeout"`
	Orientation  string  `yaml:"orientation"`
	TimeZone     string  `yaml:"timezone"`
	Interface    string  `yaml:"interface"`
}

var (
	ErrNoValue      = errors.New("no value for field 'value'")
	ErrInvalidValue = errors.New("invalid value for field 'value'")
)

func (c *Pages) UnmarshalYAML(unmarshal func(interface{}) error) error {
	miface := make(map[interface{}]interface{})
	if err := unmarshal(&miface); err == nil {
		sstr := make([]string, 0)
		for _, p := range miface {
			for _, v := range p.([]interface{}) {
				sstr = append(sstr, v.(string))
			}
		}
		c.Blocks = sstr
		return nil
	}

	return ErrInvalidValue
}

func NewConfig(configPath string) (*Config, error) {
	config := &Config{}

	file, err := os.Open(configPath)

	if err != nil {
		return nil, err
	}
	defer file.Close()

	d := yaml.NewDecoder(file)

	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}
