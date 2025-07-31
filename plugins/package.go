package plugins

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Package struct {
	Name    string
	Version string
	Os      string
	Arch    string
}

func ParsePackageName(name string, pkg *Package) error {
	if !strings.HasSuffix(name, ".ptar") {
		return fmt.Errorf("package name %q does not end with .ptar", name)
	}

	baseName := strings.TrimSuffix(name, ".ptar")
	atoms := strings.Split(baseName, "_")
	if len(atoms) != 4 {
		return fmt.Errorf("package name %q does not contain all atoms (name, version, OS, architecture)", name)
	}

	pkg.Name = atoms[0]
	pkg.Version = atoms[1]
	pkg.Os = atoms[2]
	pkg.Arch = atoms[3]
	return nil
}

func (pkg Package) PkgName() string {
	return fmt.Sprintf("%s_%s_%s_%s.ptar", pkg.Name, pkg.Version, pkg.Os, pkg.Arch)
}

func (pkg Package) PluginName() string {
	return fmt.Sprintf("%s_%s_%s_%s", pkg.Name, pkg.Version, pkg.Os, pkg.Arch)
}

func fetchPackages(url string) ([]Package, error) {
	var packages []Package
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get %s: %w", url, err)
	}
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, resp.Status)
	}
	defer resp.Body.Close()

	var lst []struct {
		Name  string
		Type  string
		Mtime string
		Size  uint64
	}
	err = json.NewDecoder(resp.Body).Decode(&lst)
	if err != nil {
		return nil, fmt.Errorf("failed to decode package list: %w", err)
	}

	for _, e := range lst {
		var pkg Package
		if ParsePackageName(e.Name, &pkg) == nil {
			//fmt.Println(e.Name)
			packages = append(packages, pkg)
		}
	}

	return packages, nil
}
