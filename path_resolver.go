package main

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
	"golang.org/x/xerrors"
)

type PathResolver struct {
	goModInfo *GoModInfo
}

func NewPathResolver(configPath string) (*PathResolver, error) {
	modInfo, err := findGoModPath(configPath)
	if err != nil {
		return nil, xerrors.Errorf(": %w", err)
	}
	return &PathResolver{
		goModInfo: modInfo,
	}, nil
}

func (r *PathResolver) ModuleName() PackagePath {
	return r.goModInfo.ModuleName
}

func (r *PathResolver) NormalizeImportPath(p string) PackagePath {
	if strings.HasPrefix(p, "./") {
		p = strings.TrimPrefix(p, "./")
		return PackagePath(path.Join(r.goModInfo.ModuleName.String(), p))
	}
	return PackagePath(p)
}

func (r *PathResolver) NormalizeImportPaths(paths []string) []PackagePath {
	result := make([]PackagePath, 0, len(paths))
	for _, p := range paths {
		result = append(result, r.NormalizeImportPath(p))
	}
	return result
}

func (r *PathResolver) ExpandWildCardSuffix(p PackagePath) ([]PackagePath, error) {
	if strings.HasSuffix(p.String(), "**") {
		if !strings.HasPrefix(p.String(), r.goModInfo.ModuleName.String()) {
			return nil, xerrors.Errorf("wildcard suffix can be used within module package: \"%s\"", r.goModInfo.ModuleName)
		}
		pathBase := strings.TrimRight(p.String(), "*")
		pathBase = strings.TrimPrefix(pathBase, r.goModInfo.ModuleName.String())
		searchRootPath := path.Join(r.goModInfo.Dir, pathBase)

		var dirList []PackagePath
		err := filepath.Walk(searchRootPath, func(path string, f os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if f.IsDir() {
				path = filepath.Join(
					r.goModInfo.ModuleName.String(),
					strings.TrimLeft(
						strings.TrimPrefix(path, r.goModInfo.Dir),
						"/",
					),
				)
				dirList = append(dirList, PackagePath(path))
			}
			return nil
		})
		if err != nil {
			return nil, xerrors.Errorf(": %w", err)
		}
		return dirList, nil
	}
	return []PackagePath{p}, nil
}

type PackagePath string

func (p PackagePath) String() string {
	return string(p)
}

type GoModInfo struct {
	Path       string
	Dir        string
	ModuleName PackagePath
}

var goModNotFound = errors.New("go.mod not found")

// return module_name, go.mod path, error
func findGoModPath(base string) (*GoModInfo, error) {
	absPath, err := filepath.Abs(base)
	if err != nil {
		return nil, xerrors.Errorf(": %w", err)
	}
	absPath = filepath.Clean(absPath)
	info, err := os.Stat(absPath)
	if err != nil {
		return nil, xerrors.Errorf(": %w", err)
	}
	var dirName string
	if !info.IsDir() {
		dirName = filepath.Dir(absPath)
	} else {
		dirName = absPath
	}
	lastDirName := ""
	for lastDirName != dirName {
		searchPath := path.Join(dirName, "go.mod")
		info, err := getModInfo(searchPath)
		if err == nil {
			return info, nil
		} else if err != goModNotFound {
			return nil, xerrors.Errorf(": %w", err)
		}
		lastDirName = dirName
		dirName = filepath.Dir(dirName)
	}
	return nil, xerrors.Errorf(": %w", goModNotFound)
}

func getModInfo(searchPath string) (*GoModInfo, error) {
	f, err := os.Open(searchPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, xerrors.Errorf(": %w", err)
		}
		return nil, goModNotFound
	}
	defer f.Close()
	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, xerrors.Errorf(": %w", err)
	}
	mf, err := modfile.Parse(searchPath, bytes, nil)
	if err != nil {
		return nil, xerrors.Errorf(": %w", err)
	}
	return &GoModInfo{
		Path:       searchPath,
		Dir:        filepath.Dir(searchPath),
		ModuleName: PackagePath(mf.Module.Mod.Path),
	}, nil
}
