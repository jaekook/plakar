package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"bufio"

	"github.com/PlakarKorp/kloset/config"
	"github.com/PlakarKorp/plakar/appcontext"
	"gopkg.in/yaml.v3"
)

type configHandler struct {
	Path string
}

func newConfigHandler(path string) *configHandler {
	return &configHandler{
		Path: path,
	}
}

func (cl *configHandler) Load() (*config.Config, error) {
	cfg := config.NewConfig()
	err := cl.load("sources.yml", &cfg.Sources)
	if err != nil {
		if os.IsNotExist(err) {
			goto fallback
		}
		return nil, err
	}
	err = cl.load("destinations.yml", &cfg.Destinations)
	if err != nil {
		if os.IsNotExist(err) {
			goto fallback
		}
		return nil, err
	}
	err = cl.load("klosets.yml", &cfg.Repositories)
	if err != nil {
		if os.IsNotExist(err) {
			goto fallback
		}
		return nil, err
	}

	for k, v := range cfg.Repositories {
		if _, ok := v[".isDefault"]; ok {
			if cfg.DefaultRepository != "" {
				return nil, fmt.Errorf("multiple default store")
			}
			cfg.DefaultRepository = k
			delete(v, ".isDefault")
		}
	}

	return cfg, nil

fallback:
	// Load old config if found
	oldpath := filepath.Join(cl.Path, "plakar.yml")
	cfg, err = LoadOldConfigIfExists(oldpath)
	if err != nil {
		return nil, fmt.Errorf("error reading old config file: %w", err)
	}

	// Save the config in the new format right now
	err = SaveConfig(cl.Path, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to update old config file: %w", err)
	}
	// Do we want to remove the old file?
	return cfg, nil
}

func (cl *configHandler) Save(cfg *config.Config) error {
	err := cl.save("sources.yml", cfg.Sources)
	if err != nil {
		return err
	}
	err = cl.save("destinations.yml", cfg.Destinations)
	if err != nil {
		return err
	}
	for k, v := range cfg.Repositories {
		if k == cfg.DefaultRepository {
			v[".isDefault"] = "yes"
		}
	}
	err = cl.save("klosets.yml", cfg.Repositories)
	if err != nil {
		return err
	}
	return nil
}

func (cl *configHandler) load(filename string, dst any) error {
	path := filepath.Join(cl.Path, filename)
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return err
		}
		return fmt.Errorf("error reading config file: %w", err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return fmt.Errorf("failed to get config file info: %w", err)
	}
	if info.Size() == 0 {
		return nil
	}

	err = yaml.NewDecoder(f).Decode(dst)
	if err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	return nil
}

func (cl *configHandler) save(filename string, src any) error {
	path := filepath.Join(cl.Path, filename)
	tmpFile, err := os.CreateTemp(cl.Path, "config.*.yml")
	if err != nil {
		return err
	}

	err = yaml.NewEncoder(tmpFile).Encode(src)
	tmpFile.Close()

	if err == nil {
		err = os.Rename(tmpFile.Name(), path)
	}

	if err != nil {
		os.Remove(tmpFile.Name())
		return err
	}

	return nil
}

func LoadConfig(configDir string) (*config.Config, error) {
	cl := newConfigHandler(configDir)
	cfg, err := cl.Load()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func SaveConfig(configDir string, cfg *config.Config) error {
	return newConfigHandler(configDir).Save(cfg)
}


func LoadIni(path string) (map[string]map[string]string, error) {
	result := make(map[string]map[string]string)
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var section string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.TrimSpace(line[1 : len(line)-1])
			result[section] = make(map[string]string)
		} else if eq := strings.Index(line, "="); eq != -1 {
			if section == "" {
				return nil, fmt.Errorf("key-value pair outside of section")
			}
			key := strings.TrimSpace(line[:eq])
			value := strings.TrimSpace(line[eq+1:])
			result[section][key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func ImportConfigFromIni(ctx *appcontext.AppContext, name string, iniMap map[string]map[string]string, mode string) error {
	if iniMap[name] == nil {
		return fmt.Errorf("no section found for destination %q in ini file", name)
	}
	
	mapConf := make(map[string]string)
	for k, v := range iniMap[name] {
		mapConf[k] = v
	}
	mapConf["location"] = mapConf["type"] + "://"
	providerSpecialCases := map[string]string{
		"drive": "googledrive",
		"google photos": "googlephotos",
	}
	for t, s := range providerSpecialCases {
		if mapConf["type"] == t {
			mapConf["location"] = s + "://"
			break
		}
	}
	if mode == "destination" {
		ctx.Config.Destinations[name] = mapConf
	} else if mode == "source" {
		ctx.Config.Sources[name] = mapConf
	} else {
		ctx.Config.Repositories[name] = mapConf
	}
	return SaveConfig(ctx.ConfigDir, ctx.Config)
}
