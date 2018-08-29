package main

import (
	"fmt"
	"github.com/KyleBanks/depth"
	"github.com/fatih/color"
	"golang.org/x/crypto/ssh/terminal"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

func exitError(msg string) {
	fmt.Printf("%v\n", msg)
	os.Exit(1)
}

func main() {
	// ターミナルに出力する場合のみカラー出力する
	color.NoColor = !terminal.IsTerminal(int(os.Stdout.Fd()))

	config, err := NewConfig()
	if err != nil {
		exitError(err.Error())
	}

	foundError := false
	for _, c := range config.Constraint {
		// "./hoge/fuga" 形式を "github.com/tomoemon/hoge/fuga" 形式に統一する
		fromOriginalPath := normalizePackagePath(c.From, config.Root)
		allowedPackages := normalizePackagePaths(c.Allow, config.Root)

		fromPaths, err := expandPackagePath(fromOriginalPath)
		if err != nil {
			exitError(err.Error())
		}
		for _, fromPath := range fromPaths {

			t := depth.Tree{
				MaxDepth: config.MaxDepth(),
			}
			err = t.Resolve(fromPath)
			if len(fromPaths) > 1 && err == depth.ErrRootPkgNotResolved {
				continue
			} else if err != nil {
				exitError(err.Error() + ": " + fromPath)
			}
			fmt.Printf("# %s\n", fromPath)
			for _, dep := range t.Root.Deps {
				if e := validate(dep, nil, *optProjectRoot, allowedPackages, config.IgnoreOther, true); e != nil {
					printResult(false, e.Error())
					foundError = true
				} else {
					printResult(true, dep.Name)
				}
			}
			fmt.Printf("\n")
		}
	}
	if foundError {
		os.Exit(1)
	} else {
		os.Exit(0)
	}
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

func validate(pkg depth.Pkg, depStack []depth.Pkg, projectRoot string, allowedPackages []string, ignoreOther bool, ignoreInternal bool) error {
	newDepStack := append(depStack[:], pkg)

	// golang 内部パッケージを無視するかどうか
	if ignoreInternal && pkg.Internal {
		return nil
	}

	// 外部パッケージの場合に無視するかどうか
	if ignoreOther {
		if !strings.HasPrefix(pkg.Name, projectRoot) || strings.HasPrefix(pkg.Name, projectRoot+"/vendor") {
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
		if e := validate(dep, newDepStack, projectRoot, allowedPackages, ignoreOther, ignoreInternal); e != nil {
			return e
		}
	}
	return nil
}
