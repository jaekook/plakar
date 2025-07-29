package plugins

import (
	"fmt"
	"strings"
)

type Package struct {
	Name    string
	Version string
	Os      string
	Arch    string
}

func ParsePackage(name string, pkg *Package) error {
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
