package main

import (
	"fmt"
	"sort"
	"strings"

	"golang.org/x/xerrors"

	"github.com/KyleBanks/depth"
)

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

func flattenSrcDir(deps []depth.Pkg) []string {
	result := make([]string, 0, len(deps))
	for _, v := range deps {
		result = append(result, v.SrcDir)
	}
	return result
}

type AssertionResult struct {
	name string
	err  error
}

func resolve(fromPath PackagePath, allowedPackages []PackagePath, projectRoot PackagePath, config *Config) ([]AssertionResult, error) {
	depthTree := depth.Tree{
		MaxDepth: config.MaxDepth(),
	}
	if err := depthTree.Resolve(fromPath.String()); err != nil {
		return nil, xerrors.Errorf(": %w", err)
	}

	result := make([]AssertionResult, 0, len(depthTree.Root.Deps))
	for _, dep := range depthTree.Root.Deps {
		err := validate(dep, nil, projectRoot, allowedPackages, config.IgnoreExternal, true)
		result = append(result, AssertionResult{
			name: dep.Name,
			err:  err,
		})
	}
	return result, nil
}

func validate(pkg depth.Pkg, depStack []depth.Pkg, projectRoot PackagePath, allowedPackages []PackagePath, ignoreExternal bool, ignoreInternal bool) error {
	newDepStack := append(depStack[:], pkg)

	if pkg.Internal {
		// golang 内部パッケージ
		if ignoreInternal {
			return nil
		}
	} else if !strings.HasPrefix(pkg.Name, projectRoot.String()) {
		// 外部パッケージ
		if ignoreExternal {
			return nil
		}
	}

	// allowedPackages に含まれないパッケージの場合はエラー
	allowed := false
	for _, allowedPackage := range allowedPackages {
		if strings.HasPrefix(pkg.Name, allowedPackage.String()) {
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
