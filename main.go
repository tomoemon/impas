package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/KyleBanks/depth"
	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
)

func main() {
	// ターミナルに出力する場合のみカラー出力する
	color.NoColor = !isatty.IsTerminal(os.Stdout.Fd())

	if succeeded, err := run(); err != nil {
		fmt.Printf("%+v\n", err.Error())
		os.Exit(1)
	} else if succeeded {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}

func run() (bool, error) {
	config, err := NewConfig()
	if err != nil {
		return false, err
	}
	for _, c := range config.Constraint {
		// "./hoge/fuga" 形式を "github.com/tomoemon/hoge/fuga" 形式に統一する
		fromOriginalPath := normalizePackagePath(c.From, config.Root)
		allowedPackages := normalizePackagePaths(c.Allow, config.Root)

		fromPaths, err := expandPackagePath(fromOriginalPath)
		if err != nil {
			return false, err
		}
		for _, fromPath := range fromPaths {
			fmt.Printf("# %s\n", fromPath)
			names, err := resolve(fromPath, allowedPackages, config)
			if len(fromPaths) > 1 && err == depth.ErrRootPkgNotResolved {
				continue
			} else if err != nil {
				if _, ok := err.(*invalidImportError); ok {
					printResult(false, err.Error())
					return false, nil
				}
				return false, err
			}
			for _, name := range names {
				printResult(true, name)
			}
			fmt.Printf("\n")
		}
	}
	return true, nil
}

func resolve(fromPath string, allowedPackages []string, config *Config) ([]string, error) {
	depthTree := depth.Tree{
		MaxDepth: config.MaxDepth(),
	}
	if err := depthTree.Resolve(fromPath); err != nil {
		return nil, err
	}
	names := make([]string, 0, len(depthTree.Root.Deps))
	for _, dep := range depthTree.Root.Deps {
		if err := validate(dep, nil, config.Root, allowedPackages, config.IgnoreExternal, true); err != nil {
			return nil, err
		} else {
			names = append(names, dep.Name)
		}
	}
	return names, nil
}

func printResult(isOk bool, message string) {
	if isOk {
		c := color.New(color.FgGreen)
		c.Printf("[OK] ")
		fmt.Println(message)
	} else {
		c := color.New(color.FgRed)
		c.Printf("[NG] %s\n", message)
	}
}

func normalizePackagePath(p string, projectRoot string) string {
	if strings.HasPrefix(p, "./") {
		return strings.Replace(path.Join(projectRoot, p), "\\", "/", -1)
	}
	return p
}

func normalizePackagePaths(paths []string, projectRoot string) []string {
	result := make([]string, 0, len(paths))
	for _, p := range paths {
		result = append(result, normalizePackagePath(p, projectRoot))
	}
	return result
}

func expandPackagePath(p string) ([]string, error) {
	if strings.HasSuffix(p, "**") {
		srcRoot := path.Join(os.Getenv("GOPATH"), "src")
		newRootPath := strings.TrimRight(path.Join(srcRoot, p), "*")
		var dirList []string
		err := filepath.Walk(newRootPath, func(path string, f os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if f.IsDir() {
				path = strings.TrimLeft(strings.Replace(path, srcRoot, "", -1), "/")
				dirList = append(dirList, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		return dirList, err
	}
	return []string{p}, nil
}

func flattenSrcDir(deps []depth.Pkg) []string {
	result := make([]string, 0, len(deps))
	for _, v := range deps {
		result = append(result, v.SrcDir)
	}
	return result
}

type invalidImportError struct {
	pkg         depth.Pkg
	importStack []depth.Pkg
}

func (e *invalidImportError) Error() string {
	stack := flattenSrcDir(e.importStack)
	sort.Sort(sort.Reverse(sort.StringSlice(stack)))
	indent := 2
	var result []string
	for _, s := range stack {
		result = append(result, strings.Repeat(" ", indent)+"from "+s)
		indent += 2
	}
	return fmt.Sprintf("%s is imported\n%s", e.pkg.Name, strings.Join(result, "\n"))
}

func validate(pkg depth.Pkg, depStack []depth.Pkg, projectRoot string, allowedPackages []string, ignoreExternal bool, ignoreInternal bool) error {
	newDepStack := append(depStack[:], pkg)

	if pkg.Internal {
		// golang 内部パッケージ
		if ignoreInternal {
			return nil
		}
	} else if !strings.HasPrefix(pkg.Name, projectRoot) || strings.HasPrefix(pkg.Name, projectRoot+"/vendor") {
		// 外部パッケージ
		if ignoreExternal {
			return nil
		}
	}

	// allowedPackages に含まれないパッケージの場合はエラー
	allowed := false
	for _, allowedPackage := range allowedPackages {
		if strings.HasPrefix(pkg.Name, allowedPackage) {
			allowed = true
			break
		}
	}
	if !allowed {
		return &invalidImportError{pkg: pkg, importStack: newDepStack}
	}

	// 再帰的に allowedPackages の依存をチェックする
	for _, dep := range pkg.Deps {
		if err := validate(dep, newDepStack, projectRoot, allowedPackages, ignoreExternal, ignoreInternal); err != nil {
			return err
		}
	}
	return nil
}
