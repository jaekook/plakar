package plugins

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"time"
)

type Cache[T any] struct {
	ttl  time.Duration
	get  func() (T, error)
	last time.Time
	v    T
}

func (c *Cache[T]) Get() (T, error) {
	if time.Since(c.last) > c.ttl {
		v, err := c.get()
		if err != nil {
			return c.v, err
		}
		c.v = v
		c.last = time.Now()
	}
	return c.v, nil
}

type Manager struct {
	ApiVersion string
	Os         string
	Arch       string

	PluginsDir  string // Where plugins are installed
	PackagesUrl string // Where packages are retrieved from

	packages     Cache[[]Package]     // list of available packages
	integrations Cache[[]Integration] // list of integrations
}

func NewManager(pluginsDir string) *Manager {
	mgr := &Manager{
		ApiVersion:  PLUGIN_API_VERSION,
		Os:          runtime.GOOS,
		Arch:        runtime.GOARCH,
		PluginsDir:  filepath.Join(pluginsDir, PLUGIN_API_VERSION),
		PackagesUrl: "https://plugins.plakar.io/kloset/pkg/" + PLUGIN_API_VERSION + "/",
	}
	mgr.packages = Cache[[]Package]{
		ttl: 5 * time.Minute,
		get: func() ([]Package, error) { return fetchPackages(mgr.PackagesUrl) },
	}
	mgr.integrations = Cache[[]Integration]{
		ttl: 5 * time.Minute,
		get: fetchIntegrationList,
	}
	return mgr
}

func (mgr *Manager) PackageUrl(pkg Package) string {
	return mgr.PackagesUrl + pkg.PkgName()
}

func (mgr *Manager) ListAvailablePackages() ([]Package, error) {
	packages, err := mgr.packages.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to list available packages: %w", err)
	}

	var res []Package
	for _, pkg := range packages {
		if pkg.Os == mgr.Os && pkg.Arch == mgr.Arch {
			res = append(res, pkg)
		}
	}

	return res, nil
}

func (mgr *Manager) IsAvailable(pkg Package) (bool, error) {
	packages, err := mgr.ListAvailablePackages()
	if err != nil {
		return false, err
	}
	for _, entry := range packages {
		if pkg == entry {
			return true, nil
		}
	}
	return false, nil
}

func (mgr *Manager) FindAvailablePackage(name string) (Package, error) {
	packages, err := mgr.packages.Get()
	if err != nil {
		return Package{}, fmt.Errorf("failed to list packages")
	}
	for _, pkg := range packages {
		if pkg.Name == name {
			return pkg, nil
		}
	}
	return Package{}, fmt.Errorf("package not available")
}

func (mgr *Manager) ListInstalledPackages() ([]Package, error) {
	var packages []Package

	dirEntries, err := os.ReadDir(mgr.PluginsDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to list installed packages: %w", err)
	}

	for _, entry := range dirEntries {
		if !entry.Type().IsRegular() {
			continue
		}
		var pkg Package
		err := ParsePackage(entry.Name(), &pkg)
		if err == nil {
			packages = append(packages, pkg)
		}
	}

	return packages, nil
}

func (mgr *Manager) FindInstalledPackage(name string) (Package, error) {
	packages, err := mgr.ListInstalledPackages()
	if err != nil {
		return Package{}, fmt.Errorf("failed to list installed packages")
	}
	for _, pkg := range packages {
		if pkg.Name == name {
			return pkg, nil
		}
	}
	return Package{}, fmt.Errorf("package not installed")
}

func (mgr *Manager) IsInstalled(pkg Package) (bool, error) {
	_, err := os.Stat(filepath.Join(mgr.PluginsDir, pkg.PkgName()))
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return false, err
		}
		return false, nil
	}
	return true, nil
}

func (mgr *Manager) ListIntegrations(filter IntegrationFilter) ([]Integration, error) {
	integrations, err := mgr.integrations.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to list integrations")
	}

	var res []Integration
	for _, info := range integrations {
		if filter.Type == "storage" && !info.Types.Storage {
			continue
		}
		if filter.Type == "source" && !info.Types.Source {
			continue
		}
		if filter.Type == "destination" && !info.Types.Destination {
			continue
		}
		if filter.Tag != "" && !slices.Contains(info.Tags, filter.Tag) {
			continue
		}
		info.Installation.Status = "not-installed"

		pkg, err := mgr.FindInstalledPackage(info.Name)
		if err == nil {
			info.Installation.Status = "installed"
			info.Installation.Version = pkg.Version
		}

		if filter.Status != "" && filter.Status != info.Installation.Status {
			continue
		}

		ok, err := mgr.IsAvailable(mgr.IntegrationAsPackage(&info))
		if ok {
			info.Installation.Available = true
		}

		res = append(res, info)
	}

	return res, nil
}

func (mgr *Manager) IntegrationAsPackage(int *Integration) Package {
	return Package{
		Name:    int.Name,
		Version: int.LatestVersion,
		Os:      mgr.Os,
		Arch:    mgr.Arch,
	}
}
