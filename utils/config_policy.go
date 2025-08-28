package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"

	"github.com/PlakarKorp/kloset/policy"
	"go.yaml.in/yaml/v3"
	"gopkg.in/ini.v1"
)

type ConfigHandler interface {
	ValidateKeyVal(key, val string) error
	ValidateEntry(map[string]string) error
}

type policiesConfig struct {
	Version  string                       `yaml:"version"`
	Policies map[string]map[string]string `yaml:"policies"`
}

func (c *policiesConfig) ValidateKeyVal(key, value string) error {
	if !slices.Contains([]string{
		"keep-minutes",
		"keep-hours",
		"keep-days",
		"keep-weeks",
		"keep-months",
		"keep-years",
		"keep-per-minute",
		"keep-per-hour",
		"keep-per-day",
		"keep-per-week",
		"keep-per-month",
		"keep-per-year",
	}, key) {
		return fmt.Errorf("invalid key")
	}

	i, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("invalid value: %w", err)
	}
	if i < 0 {
		return fmt.Errorf("negative value")
	}

	return nil
}

func (c *policiesConfig) ValidateEntry(entry map[string]string) error {
	return nil
}

func (c *policiesConfig) Has(name string) bool {
	_, ok := c.Policies[name]
	return ok
}

func (c *policiesConfig) Add(name string) {
	c.Policies[name] = make(map[string]string)
}

func (c *policiesConfig) Set(name string, key string, value string) error {
	if err := c.ValidateKeyVal(key, value); err != nil {
		return err
	}
	c.Policies[name][key] = value
	return nil
}

func (c *policiesConfig) Unset(name string, key string) {
	delete(c.Policies[name], key)
}

func (c *policiesConfig) Remove(name string) {
	delete(c.Policies, name)
}

func (c *policiesConfig) SaveToFile(filename string) error {
	tmpFile, err := os.CreateTemp(filepath.Dir(filename), filepath.Base(filename)+".*.tmp")
	if err != nil {
		return err
	}
	err = yaml.NewEncoder(tmpFile).Encode(c)
	tmpFile.Close()
	if err == nil {
		err = os.Rename(tmpFile.Name(), filename)
	}
	os.Remove(tmpFile.Name())
	return err
}

func (c *policiesConfig) Load(rd io.Reader) error {
	return yaml.NewDecoder(rd).Decode(c)
}

func marshalINISections(sectionName string, kv map[string]string, w io.Writer) error {
	cfg := ini.Empty()

	section := cfg.Section(sectionName)
	for key, value := range kv {
		section.Key(key).SetValue(value)
	}
	_, err := cfg.WriteTo(w)
	return err
}

func (c *policiesConfig) Dump(w io.Writer, format string, names []string) error {

	if len(names) == 0 {
		for name := range c.Policies {
			names = append(names, name)
		}
	}

	for _, name := range names {
		if !c.Has(name) {
			return fmt.Errorf("entry %q not found", name)
		}
		var err error
		switch format {
		case "json":
			err = json.NewEncoder(w).Encode(map[string]map[string]string{name: c.Policies[name]})
		case "ini":
			err = marshalINISections(name, c.Policies[name], w)
		case "yaml":
			err = yaml.NewEncoder(w).Encode(map[string]map[string]string{name: c.Policies[name]})
		default:
			return fmt.Errorf("unknown format %q", format)
		}
		if err != nil {
			return fmt.Errorf("failed to encode entry %q: %w", name, err)
		}
	}

	return nil
}

func (c *policiesConfig) ApplyConfig(name string, po *policy.PolicyOptions) {
	apply := func(setter func(int) policy.Option, key string) {
		entry, ok := c.Policies[name]
		if !ok {
			return
		}
		value, ok := entry[key]
		if !ok {
			return
		}
		i, err := strconv.Atoi(value)
		if err != nil {
			return
		}
		setter(i)(po)
	}
	apply(policy.WithKeepMinutes, "keep-minutes")
	apply(policy.WithKeepHours, "keep-hours")
	apply(policy.WithKeepDays, "keep-days")
	apply(policy.WithKeepWeeks, "keep-weeks")
	apply(policy.WithKeepMonths, "keep-months")
	apply(policy.WithKeepYears, "keep-years")
	apply(policy.WithPerMinuteCap, "keep-per-minute")
	apply(policy.WithPerHourCap, "keep-per-hour")
	apply(policy.WithPerDayCap, "keep-per-day")
	apply(policy.WithPerWeekCap, "keep-per-week")
	apply(policy.WithPerMonthCap, "keep-per-month")
	apply(policy.WithPerYearCap, "keep-per-year")
}

func LoadPolicyConfigFile(filename string) (*policiesConfig, error) {
	var cfg policiesConfig
	cfg.Policies = make(map[string]map[string]string)

	rd, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return &cfg, nil
		}
		return nil, err
	}
	defer rd.Close()

	if err := cfg.Load(rd); err != nil {
		return nil, err
	}

	return &cfg, nil
}
