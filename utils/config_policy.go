package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/PlakarKorp/kloset/locate"
	"go.yaml.in/yaml/v3"
)

type LocatePolicyConfig struct {
	Before time.Time `yaml:"before,omitempty"`
	Since  time.Time `yaml:"since,omitempty"`

	Name        string `yaml:"name,omitempty"`
	Category    string `yaml:"category,omitempty"`
	Environment string `yaml:"environment,omitempty"`
	Perimeter   string `yaml:"perimeter,omitempty"`
	Job         string `yaml:"job,omitempty"`

	Latest bool `yaml:"latest,omitempty"`

	Tags  []string `yaml:"tags,omitempty"`
	IDs   []string `yaml:"ids,omitempty"`
	Roots []string `yaml:"roots,omitempty"`

	Minutes    int `yaml:"minutes,omitempty"`
	Hours      int `yaml:"hours,omitempty"`
	Days       int `yaml:"days,omitempty"`
	Weeks      int `yaml:"weeks,omitempty"`
	Months     int `yaml:"months,omitempty"`
	Years      int `yaml:"years,omitempty"`
	Mondays    int `yaml:"mondays,omitempty"`
	Tuesdays   int `yaml:"tuesdays,omitempty"`
	Wednesdays int `yaml:"wednesdays,omitempty"`
	Thursdays  int `yaml:"thursdays,omitempty"`
	Fridays    int `yaml:"fridays,omitempty"`
	Saturdays  int `yaml:"saturdays,omitempty"`
	Sundays    int `yaml:"sundays,omitempty"`

	PerMinute    int `yaml:"per-minute,omitempty"`
	PerHour      int `yaml:"per-hour,omitempty"`
	PerDay       int `yaml:"per-day,omitempty"`
	PerWeek      int `yaml:"per-week,omitempty"`
	PerMonth     int `yaml:"per-month,omitempty"`
	PerYear      int `yaml:"per-year,omitempty"`
	PerMonday    int `yaml:"per-monday,omitempty"`
	PerTuesday   int `yaml:"per-tuesday,omitempty"`
	PerWednesday int `yaml:"per-wednesday,omitempty"`
	PerThursday  int `yaml:"per-thursday,omitempty"`
	PerFriday    int `yaml:"per-friday,omitempty"`
	PerSaturday  int `yaml:"per-saturday,omitempty"`
	PerSunday    int `yaml:"per-sunday,omitempty"`
}

type policiesConfig struct {
	Version  string                         `yaml:"version"`
	Policies map[string]*LocatePolicyConfig `yaml:"policies"`
}

func (c *policiesConfig) Has(name string) bool {
	_, ok := c.Policies[name]
	return ok
}

func (c *policiesConfig) Add(name string) {
	c.Policies[name] = &LocatePolicyConfig{}
}

func (c *policiesConfig) setInt(value string, p *int) error {
	i, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("invalid value: %w", err)
	}
	if i < 0 {
		return fmt.Errorf("negative value")
	}
	*p = i
	return nil
}

func (c *policiesConfig) setBool(value string, p *bool) error {
	if value == "true" || value == "yes" {
		*p = true
	} else if value == "false" || value == "no" {
		*p = false
	} else {
		return fmt.Errorf("invalid value")
	}
	return nil
}

func (c *policiesConfig) setTime(value string, p *time.Time) error {
	t, err := locate.ParseTimeFlag(value)
	if err != nil {
		return fmt.Errorf("invalid value: %w", err)
	}
	*p = t
	return nil
}

func (c *policiesConfig) setStringList(value string, p *[]string) error {
	*p = strings.Split(value, ",")
	return nil
}

func (c *policiesConfig) locateField(name string, key string) (any, error) {
	p, ok := c.Policies[name]
	if !ok {
		return nil, fmt.Errorf("entry not found")
	}

	switch key {
	case "before":
		return &p.Before, nil
	case "since":
		return &p.Since, nil
	case "name":
		return &p.Name, nil
	case "category":
		return &p.Category, nil
	case "environment":
		return &p.Environment, nil
	case "perimeter":
		return &p.Perimeter, nil
	case "job":
		return &p.Job, nil
	case "tags":
		return &p.Tags, nil
	case "ids":
		return &p.IDs, nil
	case "roots":
		return &p.Roots, nil
	case "latest":
		return &p.Latest, nil
	case "minutes":
		return &p.Minutes, nil
	case "hours":
		return &p.Hours, nil
	case "days":
		return &p.Days, nil
	case "weeks":
		return &p.Weeks, nil
	case "months":
		return &p.Months, nil
	case "years":
		return &p.Years, nil
	case "mondays":
		return &p.Mondays, nil
	case "tuesdays":
		return &p.Tuesdays, nil
	case "wednesdays":
		return &p.Wednesdays, nil
	case "thursdays":
		return &p.Thursdays, nil
	case "fridays":
		return &p.Fridays, nil
	case "saturdays":
		return &p.Saturdays, nil
	case "sundays":
		return &p.Sundays, nil

	case "per-minute":
		return &p.PerMinute, nil
	case "per-hour":
		return &p.PerHour, nil
	case "per-day":
		return &p.PerDay, nil
	case "per-week":
		return &p.PerWeek, nil
	case "per-month":
		return &p.PerMonth, nil
	case "per-year":
		return &p.PerYear, nil
	case "per-monday":
		return &p.PerMonday, nil
	case "per-tuesday":
		return &p.PerTuesday, nil
	case "per-wednesday":
		return &p.PerWednesday, nil
	case "per-thursday":
		return &p.PerThursday, nil
	case "per-friday":
		return &p.PerFriday, nil
	case "per-saturday":
		return &p.PerSaturday, nil
	case "per-sunday":
		return &p.PerSunday, nil

	default:
		return nil, fmt.Errorf("invalid key")
	}
}

func (c *policiesConfig) Set(name string, key string, value string) error {
	field, err := c.locateField(name, key)
	if err != nil {
		return err
	}

	switch p := field.(type) {
	case *int:
		return c.setInt(value, p)
	case *time.Time:
		return c.setTime(value, p)
	case *string:
		*p = value
		return nil
	case *[]string:
		return c.setStringList(value, p)
	case *bool:
		return c.setBool(value, p)
	default:
		return fmt.Errorf("invalid field type")
	}
}

func (c *policiesConfig) Unset(name string, key string) error {
	field, err := c.locateField(name, key)
	if err != nil {
		return err
	}

	switch p := field.(type) {
	case *int:
		*p = 0
	case *time.Time:
		*p = time.Time{}
	case *string:
		*p = ""
	case *[]string:
		*p = nil
	case *bool:
		*p = false
	default:
		return fmt.Errorf("invalid field type")
	}

	return nil
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
			err = json.NewEncoder(w).Encode(map[string]*LocatePolicyConfig{name: c.Policies[name]})
		case "yaml":
			err = yaml.NewEncoder(w).Encode(map[string]*LocatePolicyConfig{name: c.Policies[name]})
		default:
			return fmt.Errorf("unknown format %q", format)
		}
		if err != nil {
			return fmt.Errorf("failed to encode entry %q: %w", name, err)
		}
	}

	return nil
}

func LoadPolicyConfigFile(filename string) (*policiesConfig, error) {
	var cfg policiesConfig
	cfg.Version = "v1.0.0"
	cfg.Policies = make(map[string]*LocatePolicyConfig)

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

func (c *policiesConfig) ApplyConfig(name string, po *locate.LocateOptions) {
	p, ok := c.Policies[name]
	if !ok {
		return
	}

	po.Filters.Before = p.Before
	po.Filters.Since = p.Since

	po.Filters.Name = p.Name
	po.Filters.Category = p.Category
	po.Filters.Environment = p.Environment
	po.Filters.Perimeter = p.Perimeter
	po.Filters.Job = p.Job

	po.Filters.Latest = p.Latest

	po.Filters.Tags = p.Tags
	po.Filters.IDs = p.IDs
	po.Filters.Roots = p.Roots

	po.Minute.Keep = p.Minutes
	po.Hour.Keep = p.Hours
	po.Day.Keep = p.Days
	po.Week.Keep = p.Weeks
	po.Month.Keep = p.Months
	po.Year.Keep = p.Years
	po.Monday.Keep = p.Mondays
	po.Tuesday.Keep = p.Tuesdays
	po.Wednesday.Keep = p.Wednesdays
	po.Thursday.Keep = p.Thursdays
	po.Friday.Keep = p.Fridays
	po.Saturday.Keep = p.Saturdays
	po.Sunday.Keep = p.Sundays

	po.Minute.Cap = p.PerMinute
	po.Hour.Cap = p.PerHour
	po.Day.Cap = p.PerDay
	po.Week.Cap = p.PerWeek
	po.Month.Cap = p.PerMonth
	po.Year.Cap = p.PerYear
	po.Monday.Cap = p.PerMonday
	po.Tuesday.Cap = p.PerTuesday
	po.Wednesday.Cap = p.PerWednesday
	po.Thursday.Cap = p.PerThursday
	po.Friday.Cap = p.PerFriday
	po.Saturday.Cap = p.PerSaturday
	po.Sunday.Cap = p.PerSunday
}
